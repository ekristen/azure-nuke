package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const ResourceGroupResource = "ResourceGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   ResourceGroupResource,
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
}

func (l ResourceGroupLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", ResourceGroupResource).WithField("s", opts.SubscriptionId)

	client := resources.NewGroupsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	ctx := context.TODO()

	list, err := client.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for list.NotDone() {
		log.Trace("list not done")
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

	log.Trace("done")

	return resources, nil
}
