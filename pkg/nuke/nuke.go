package nuke

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/config"
	"github.com/ekristen/azure-nuke/pkg/queue"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
	"github.com/ekristen/azure-nuke/pkg/utils"
)

type NukeParameters struct {
	ConfigPath string

	Targets  []string
	Excludes []string

	NoDryRun   bool
	Force      bool
	ForceSleep int
	Quiet      bool

	MaxWaitRetries int
}

type Nuke struct {
	Parameters     NukeParameters
	TenantId       string
	SubscriptionId string
	Config         *config.Nuke
	Tenant         *azure.Tenant

	ResourceTypes types.Collection

	items queue.Queue
}

func New(params NukeParameters, tenant *azure.Tenant) *Nuke {
	n := Nuke{
		Parameters: params,
		Tenant:     tenant,
	}

	return &n
}

func (n *Nuke) Run() error {
	var err error

	if n.Parameters.ForceSleep < 3 {
		return fmt.Errorf("Value for --force-sleep cannot be less than 3 seconds. This is for your own protection.")
	}
	forceSleep := time.Duration(n.Parameters.ForceSleep) * time.Second

	//fmt.Printf("aws-nuke version %s - %s - %s\n\n", BuildVersion, BuildDate, BuildHash)

	err = n.Config.ValidateTenant(n.Tenant.ID)
	if err != nil {
		return err
	}

	/*
		err = n.Config.ValidateAccount(n.Account.ID(), n.Account.Aliases())
		if err != nil {
			return err
		}

		fmt.Printf("Do you really want to nuke the account with "+
			"the ID %s and the alias '%s'?\n", n.Account.ID(), n.Account.Alias())
		if n.Parameters.Force {
			fmt.Printf("Waiting %v before continuing.\n", forceSleep)
			time.Sleep(forceSleep)
		} else {
			fmt.Printf("Do you want to continue? Enter account alias to continue.\n")
			err = Prompt(n.Account.Alias())
			if err != nil {
				return err
			}
		}
	*/

	err = n.Scan()
	if err != nil {
		return err
	}

	if n.items.Count(queue.ItemStateNew) == 0 {
		fmt.Println("No resource to delete.")
		return nil
	}

	if !n.Parameters.NoDryRun {
		fmt.Println("The above resources would be deleted with the supplied configuration. Provide --no-dry-run to actually destroy resources.")
		return nil
	}

	/*
		fmt.Printf("Do you really want to nuke these resources on the account with "+
			"the ID %s and the alias '%s'?\n", n.Account.ID(), n.Account.Alias())
		if n.Parameters.Force {
			fmt.Printf("Waiting %v before continuing.\n", forceSleep)
			time.Sleep(forceSleep)
		} else {
			fmt.Printf("Do you want to continue? Enter account alias to continue.\n")
			err = Prompt(n.Account.Alias())
			if err != nil {
				return err
			}
		}
	*/

	// TODO: temporary location
	time.Sleep(forceSleep)

	failCount := 0
	waitingCount := 0

	for {
		n.HandleQueue()

		if n.items.Count(queue.ItemStatePending, queue.ItemStateWaiting, queue.ItemStateNew) == 0 && n.items.Count(queue.ItemStateFailed) > 0 {
			if failCount >= 2 {
				logrus.Errorf("There are resources in failed state, but none are ready for deletion, anymore.")
				fmt.Println()

				for _, item := range n.items {
					if item.State != queue.ItemStateFailed {
						continue
					}

					item.Print()
					logrus.Error(item.Reason)
				}

				return fmt.Errorf("failed")
			}

			failCount = failCount + 1
		} else {
			failCount = 0
		}
		if n.Parameters.MaxWaitRetries != 0 && n.items.Count(queue.ItemStateWaiting, queue.ItemStatePending) > 0 && n.items.Count(queue.ItemStateNew) == 0 {
			if waitingCount >= n.Parameters.MaxWaitRetries {
				return fmt.Errorf("Max wait retries of %d exceeded.\n\n", n.Parameters.MaxWaitRetries)
			}
			waitingCount = waitingCount + 1
		} else {
			waitingCount = 0
		}
		if n.items.Count(queue.ItemStateNew, queue.ItemStatePending, queue.ItemStateFailed, queue.ItemStateWaiting) == 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Printf("Nuke complete: %d failed, %d skipped, %d finished.\n\n",
		n.items.Count(queue.ItemStateFailed), n.items.Count(queue.ItemStateFiltered), n.items.Count(queue.ItemStateFinished))

	return nil
}

func (n *Nuke) Scan() error {
	tenantConfig := n.Config.Tenants[n.Tenant.ID]

	tenantResourceTypes := utils.ResolveResourceTypes(
		resource.GetListersNameForScope(resource.Tenant),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes.Targets,
			tenantConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes.Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
	)

	subscriptionResourceTypes := utils.ResolveResourceTypes(
		resource.GetListersNameForScope(resource.Subscription),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes.Targets,
			tenantConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes.Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
	)

	resourceGroupResourceTypes := utils.ResolveResourceTypes(
		resource.GetListersNameForScope(resource.ResourceGroup),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes.Targets,
			tenantConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes.Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
	)

	itemqueue := make(queue.Queue, 0)

	// Handle Tenant ResourceTypes
	items := Scan(n.Tenant.Authorizers, n.Tenant.ID, "tenant", "", tenantResourceTypes)
	for item := range items {
		ffGetter, ok := item.Resource.(resource.FeatureFlagGetter)
		if ok {
			ffGetter.FeatureFlags(n.Config.FeatureFlags)
		}

		itemqueue = append(itemqueue, item)
		err := n.Filter(item)
		if err != nil {
			return err
		}

		if item.State != queue.ItemStateFiltered || !n.Parameters.Quiet {
			item.Print()
		}
	}

	// Loop Subscriptions
	// Loop ResourceGroups
	for _, subscriptionId := range n.Tenant.SubscriptionIds {

		// Scan for Subscription Resources
		items := Scan(n.Tenant.Authorizers, n.Tenant.ID, subscriptionId, "", subscriptionResourceTypes)
		for item := range items {
			ffGetter, ok := item.Resource.(resource.FeatureFlagGetter)
			if ok {
				ffGetter.FeatureFlags(n.Config.FeatureFlags)
			}

			itemqueue = append(itemqueue, item)
			err := n.Filter(item)
			if err != nil {
				return err
			}

			if item.State != queue.ItemStateFiltered || !n.Parameters.Quiet {
				item.Print()
			}
		}

		// Scan for ResourceGroup resources
		for _, resourceGroup := range n.Tenant.ResourceGroups[subscriptionId] {
			logrus.WithField("subscription", subscriptionId).WithField("group", resourceGroup).Trace("scanning")
			items := Scan(n.Tenant.Authorizers, n.Tenant.ID, subscriptionId, resourceGroup, resourceGroupResourceTypes)
			for item := range items {
				ffGetter, ok := item.Resource.(resource.FeatureFlagGetter)
				if ok {
					ffGetter.FeatureFlags(n.Config.FeatureFlags)
				}

				itemqueue = append(itemqueue, item)
				err := n.Filter(item)
				if err != nil {
					return err
				}

				if item.State != queue.ItemStateFiltered || !n.Parameters.Quiet {
					item.Print()
				}
			}
		}
	}

	fmt.Printf("Scan complete: %d total, %d nukeable, %d filtered.\n\n",
		itemqueue.CountTotal(), itemqueue.Count(queue.ItemStateNew), itemqueue.Count(queue.ItemStateFiltered))

	n.items = itemqueue

	return nil
}

func (n *Nuke) Filter(item *queue.Item) error {
	checker, ok := item.Resource.(resource.Filter)
	if ok {
		err := checker.Filter()
		if err != nil {
			item.State = queue.ItemStateFiltered
			item.Reason = err.Error()

			// Not returning the error, since it could be because of a failed
			// request to the API. We do not want to block the whole nuking,
			// because of an issue on AWS side.
			return nil
		}
	}

	tenantFilters, err := n.Config.Filters(n.Tenant.ID)
	if err != nil {
		return err
	}

	itemFilters, ok := tenantFilters[item.Type]
	if !ok {
		return nil
	}

	for _, filter := range itemFilters {
		prop, err := item.GetProperty(filter.Property)
		if err != nil {
			return err
		}

		match, err := filter.Match(prop)
		if err != nil {
			return err
		}

		if utils.IsTrue(filter.Invert) {
			match = !match
		}

		if match {
			item.State = queue.ItemStateFiltered
			item.Reason = "filtered by config"
			return nil
		}
	}

	return nil
}

func (n *Nuke) HandleQueue() {
	listCache := make(map[string]map[string][]resource.Resource)

	for _, item := range n.items {
		switch item.State {
		case queue.ItemStateNew:
			n.HandleRemove(item)
			item.Print()
		case queue.ItemStateFailed:
			n.HandleRemove(item)
			n.HandleWait(item, listCache)
			item.Print()
		case queue.ItemStatePending:
			n.HandleWait(item, listCache)
			item.State = queue.ItemStateWaiting
			item.Print()
		case queue.ItemStateWaiting:
			n.HandleWait(item, listCache)
			item.Print()
		}

	}

	fmt.Println()
	fmt.Printf("Removal requested: %d waiting, %d failed, %d skipped, %d finished\n\n",
		n.items.Count(queue.ItemStateWaiting, queue.ItemStatePending), n.items.Count(queue.ItemStateFailed),
		n.items.Count(queue.ItemStateFiltered), n.items.Count(queue.ItemStateFinished))
}

func (n *Nuke) HandleRemove(item *queue.Item) {
	err := item.Resource.Remove()
	if err != nil {
		item.State = queue.ItemStateFailed
		item.Reason = err.Error()
		return
	}

	item.State = queue.ItemStatePending
	item.Reason = ""
}

func (n *Nuke) HandleWait(item *queue.Item, cache map[string]map[string][]resource.Resource) {
	var err error
	subscriptionId := item.SubscriptionId
	_, ok := cache[subscriptionId]
	if !ok {
		cache[subscriptionId] = map[string][]resource.Resource{}
	}
	left, ok := cache[subscriptionId][item.Type]
	if !ok {
		left, err = item.List()
		if err != nil {
			item.State = queue.ItemStateFailed
			item.Reason = err.Error()
			return
		}
		cache[subscriptionId][item.Type] = left
	}

	for _, r := range left {
		if item.Equals(r) {
			checker, ok := r.(resource.Filter)
			if ok {
				err := checker.Filter()
				if err != nil {
					break
				}
			}

			return
		}
	}

	item.State = queue.ItemStateFinished
	item.Reason = ""
}
