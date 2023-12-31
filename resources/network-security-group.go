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

func init() {
	resource.Register(resource.Registration{
		Name:   "NetworkSecurityGroup",
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
	opts nuke.ListerOpts
}

func (l NetworkSecurityGroupLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l NetworkSecurityGroupLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := network.NewSecurityGroupsClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list groups")

	ctx := context.Background()

	list, err := client.List(ctx, l.opts.ResourceGroup)
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
				rg:       &l.opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}
