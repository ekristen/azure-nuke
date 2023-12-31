package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	sdknuke "github.com/ekristen/cloud-nuke-sdk/pkg/nuke"
)

type Nuke struct {
	sdknuke.Nuke

	TenantId       string
	SubscriptionId string
	Tenant         *azure.Tenant
}

func New(params sdknuke.Parameters, tenant *azure.Tenant) *Nuke {
	n := Nuke{
		Nuke: sdknuke.Nuke{
			Parameters: params,
		},
		Tenant: tenant,
	}

	return &n
}

/*
func (n *Nuke) ssscan() error {
	azureConfig := n.Config.(config.Config)
	tenantConfig := azureConfig.Tenants[n.Tenant.ID]

	tenantResourceTypes := utils.ResolveResourceTypes(
		resource.GetNamesForScope(Tenant),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes().Targets,
			tenantConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes().Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
	)

	subscriptionResourceTypes := utils.ResolveResourceTypes(
		resource.GetNamesForScope(Subscription),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes().Targets,
			tenantConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes().Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
	)

	resourceGroupResourceTypes := utils.ResolveResourceTypes(
		resource.GetNamesForScope(ResourceGroup),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes().Targets,
			tenantConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes().Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
	)

	itemQueue := queue.Queue{}

	// Handle Tenant ResourceTypes
	items := Scan(tenantResourceTypes, ListerOpts{
		Authorizers:    n.Tenant.Authorizers,
		TenantId:       n.Tenant.ID,
		SubscriptionId: "tenant",
		ResourceGroup:  "",
	})
	for item := range items {
		ffGetter, ok := item.Resource.(resource.FeatureFlagGetter)
		if ok {
			ffGetter.FeatureFlags(n.Config.FeatureFlags())
		}

		itemQueue.Items = append(itemQueue.Items, item)
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
		items := Scan(subscriptionResourceTypes, ListerOpts{
			Authorizers:    n.Tenant.Authorizers,
			TenantId:       n.Tenant.ID,
			SubscriptionId: subscriptionId,
			ResourceGroup:  "",
		})
		for item := range items {
			ffGetter, ok := item.Resource.(resource.FeatureFlagGetter)
			if ok {
				ffGetter.FeatureFlags(n.Config.FeatureFlags())
			}

			itemQueue.Items = append(itemQueue.Items, item)
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
			items := Scan(resourceGroupResourceTypes, ListerOpts{
				Authorizers:    n.Tenant.Authorizers,
				TenantId:       n.Tenant.ID,
				SubscriptionId: subscriptionId,
				ResourceGroup:  resourceGroup,
			})
			for item := range items {
				ffGetter, ok := item.Resource.(resource.FeatureFlagGetter)
				if ok {
					ffGetter.FeatureFlags(n.Config.FeatureFlags())
				}

				itemQueue.Items = append(itemQueue.Items, item)
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
		itemQueue.Count(), itemQueue.Count(queue.ItemStateNew), itemQueue.Count(queue.ItemStateFiltered))

	n.Queue = itemQueue

	return nil
}

*/
