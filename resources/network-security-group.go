package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type NetworkSecurityGroup struct {
	client   network.SecurityGroupsClient
	name     *string
	location *string
	rg       *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "NetworkSecurityGroup",
		Scope:  resource.ResourceGroup,
		Lister: ListNetworkSecurityGroup,
	})
}

func ListNetworkSecurityGroup(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := network.NewSecurityGroupsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list groups")

	ctx := context.Background()

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for list.NotDone() {
		logrus.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &NetworkSecurityGroup{
				client:   client,
				name:     g.Name,
				location: g.Location,
				rg:       &opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *NetworkSecurityGroup) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.rg, *r.name)
	return err
}

func (r *NetworkSecurityGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("Location", *r.location)

	return properties
}

func (r *NetworkSecurityGroup) String() string {
	return *r.name
}
