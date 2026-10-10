package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns/v2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const PrivateDNSZoneResource = "PrivateDNSZone"

func init() {
	registry.Register(&registry.Registration{
		Name:     PrivateDNSZoneResource,
		Scope:    azure.SubscriptionScope,
		Resource: &PrivateDNSZone{},
		Lister:   &PrivateDNSZoneLister{},
	})
}

type PrivateDNSZone struct {
	*BaseResource `property:",inline"`

	client *armprivatedns.PrivateZonesClient
	Name   *string
	Tags   map[string]*string
}

func (r *PrivateDNSZone) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func (r *PrivateDNSZone) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PrivateDNSZone) String() string {
	return *r.Name
}

type PrivateDNSZoneLister struct{}

func (l PrivateDNSZoneLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithFields(logrus.Fields{
		"r": PrivateDNSZoneResource,
		"s": opts.SubscriptionID,
	})

	log.Trace("start")

	client, err := armprivatedns.NewPrivateZonesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing entities")

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			resources = append(resources, &PrivateDNSZone{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  azure.GetResourceGroupFromID(*entity.ID),
					SubscriptionID: ptr.String(opts.SubscriptionID),
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
