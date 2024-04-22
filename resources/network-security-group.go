package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const NetworkSecurityGroupResource = "NetworkSecurityGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkSecurityGroupResource,
		Scope:    nuke.ResourceGroup,
		Resource: &NetworkSecurityGroup{},
		Lister:   &NetworkSecurityGroupLister{},
	})
}

type NetworkSecurityGroup struct {
	client        network.SecurityGroupsClient
	Region        *string
	ResourceGroup *string
	Name          *string
	Tags          map[string]*string
}

func (r *NetworkSecurityGroup) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *NetworkSecurityGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkSecurityGroup) String() string {
	return *r.Name
}

type NetworkSecurityGroupLister struct {
}

func (l NetworkSecurityGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", NetworkSecurityGroupResource).WithField("s", opts.SubscriptionID)

	client := network.NewSecurityGroupsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &NetworkSecurityGroup{
				client:        client,
				Region:        g.Location,
				ResourceGroup: &opts.ResourceGroup,
				Name:          g.Name,
				Tags:          g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
