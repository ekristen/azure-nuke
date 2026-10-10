package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v12"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const PublicIPAddressesResource = "PublicIPAddress"

func init() {
	registry.Register(&registry.Registration{
		Name:     PublicIPAddressesResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &PublicIPAddresses{},
		Lister:   &PublicIPAddressesLister{},
		DeprecatedAliases: []string{
			"PublicIPAddresses",
		},
	})
}

type PublicIPAddresses struct {
	*BaseResource `property:",inline"`

	client   *armnetwork.PublicIPAddressesClient
	removeOp *runtime.Poller[armnetwork.PublicIPAddressesClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *PublicIPAddresses) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	r.removeOp = poller
	return nil
}

func (r *PublicIPAddresses) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *PublicIPAddresses) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PublicIPAddresses) String() string {
	return *r.Name
}

type PublicIPAddressesLister struct{}

func (l PublicIPAddressesLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", PublicIPAddressesResource).WithField("s", opts.SubscriptionID)

	client, err := armnetwork.NewPublicIPAddressesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list public ip addresses")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &PublicIPAddresses{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   entity.Name,
				Tags:   entity.Tags,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
