package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type ResourceGroup struct {
	client resources.GroupsClient
	name   *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "ResourceGroup",
		Lister: ListResourceGroup,
		Scope:  resource.Subscription,
	})
}

func ListResourceGroup(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := resources.NewGroupsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
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
		for _, g := range list.Values() {
			resources = append(resources, &ResourceGroup{
				client: client,
				name:   g.Name,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
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
