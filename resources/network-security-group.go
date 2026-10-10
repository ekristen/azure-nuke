package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v12"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const NetworkSecurityGroupResource = "NetworkSecurityGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkSecurityGroupResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &NetworkSecurityGroup{},
		Lister:   &NetworkSecurityGroupLister{},
	})
}

type NetworkSecurityGroup struct {
	*BaseResource `property:",inline"`

	client *armnetwork.SecurityGroupsClient
	Name   *string
	Tags   map[string]*string
}

func (r *NetworkSecurityGroup) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func (r *NetworkSecurityGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkSecurityGroup) String() string {
	return *r.Name
}

type NetworkSecurityGroupLister struct{}

func (l NetworkSecurityGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", NetworkSecurityGroupResource).WithField("s", opts.SubscriptionID)

	client, err := armnetwork.NewSecurityGroupsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list groups")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &NetworkSecurityGroup{
				BaseResource: &BaseResource{
					Region:        entity.Location,
					ResourceGroup: &opts.ResourceGroup,
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
