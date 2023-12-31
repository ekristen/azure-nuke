package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "ResourceGroup",
		Lister: ResourceGroupLister{},
		Scope:  nuke.Subscription,
	})
}

type ResourceGroup struct {
	client resources.GroupsClient
	name   *string
}

func (r *ResourceGroup) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.name)
	return err
}

func (r *ResourceGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *ResourceGroup) String() string {
	return *r.name
}

// -------------------

type ResourceGroupLister struct {
	opts nuke.ListerOpts
}

func (l ResourceGroupLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l ResourceGroupLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := resources.NewGroupsClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list groups")

	ctx := context.Background()

	list, err := client.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, entity := range list.Values() {
			resources = append(resources, &ResourceGroup{
				client: client,
				name:   entity.Name,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}
