package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

const NetworkSecurityGroupResource = "NetworkSecurityGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   NetworkSecurityGroupResource,
		Scope:  nuke.ResourceGroup,
		Lister: NetworkSecurityGroupLister{},
	})
}

type NetworkSecurityGroup struct {
	client   network.SecurityGroupsClient
	name     *string
	location *string
	rg       *string
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

type NetworkSecurityGroupLister struct {
}

func (l NetworkSecurityGroupLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", NetworkSecurityGroupResource).WithField("s", opts.SubscriptionId)

	client := network.NewSecurityGroupsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	ctx := context.TODO()

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing")

	for list.NotDone() {
		log.Trace("list not done")
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

	log.Trace("done")

	return resources, nil
}
