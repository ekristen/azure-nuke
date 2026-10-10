package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const SSHPublicKeyResource = "SSHPublicKey"

func init() {
	registry.Register(&registry.Registration{
		Name:     SSHPublicKeyResource,
		Scope:    azure.SubscriptionScope,
		Resource: &SSHPublicKey{},
		Lister:   &SSHPublicKeyLister{},
	})
}

type SSHPublicKey struct {
	*BaseResource `property:",inline"`

	client *armcompute.SSHPublicKeysClient
	Name   *string
	Tags   map[string]*string
}

func (r *SSHPublicKey) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name, nil)
	return err
}

func (r *SSHPublicKey) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SSHPublicKey) String() string {
	return *r.Name
}

// --------------------------------------

type SSHPublicKeyLister struct{}

func (l SSHPublicKeyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", SSHPublicKeyResource).WithField("s", opts.SubscriptionID)

	client, err := armcompute.NewSSHPublicKeysClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ssh keys")

	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &SSHPublicKey{
				BaseResource: &BaseResource{
					Region:         &opts.Region,
					SubscriptionID: &opts.SubscriptionID,
					ResourceGroup:  azure.GetResourceGroupFromID(*entity.ID),
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
