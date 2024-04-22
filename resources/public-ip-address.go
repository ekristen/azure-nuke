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

const PublicIPAddressesResource = "PublicIPAddresses"

func init() {
	registry.Register(&registry.Registration{
		Name:     PublicIPAddressesResource,
		Scope:    nuke.ResourceGroup,
		Resource: &PublicIPAddresses{},
		Lister:   &PublicIPAddressesLister{},
	})
}

type PublicIPAddresses struct {
	client        network.PublicIPAddressesClient
	Region        *string
	ResourceGroup *string
	Name          *string
	Tags          map[string]*string
}

func (r *PublicIPAddresses) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *PublicIPAddresses) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PublicIPAddresses) String() string {
	return *r.Name
}

type PublicIPAddressesLister struct {
}

func (l PublicIPAddressesLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", PublicIPAddressesResource).WithField("s", opts.SubscriptionID)

	client := network.NewPublicIPAddressesClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list public ip addresses")

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing public ip addresses")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &PublicIPAddresses{
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
