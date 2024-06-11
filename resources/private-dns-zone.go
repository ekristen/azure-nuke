package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns" //nolint:staticcheck

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

	client privatedns.PrivateZonesClient
	Name   *string
	Tags   map[string]*string
}

func (r *PrivateDNSZone) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name, "")
	return err
}

func (r *PrivateDNSZone) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PrivateDNSZone) String() string {
	return *r.Name
}

type PrivateDNSZoneLister struct {
}

func (l PrivateDNSZoneLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithFields(logrus.Fields{
		"r": PrivateDNSZoneResource,
		"s": opts.SubscriptionID,
	})

	log.Trace("start")

	client := privatedns.NewPrivateZonesClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	list, err := client.List(ctx, nil)
	if err != nil {
		log.WithError(err).Error("unable to list")
		return nil, err
	}

	log.Trace("listing entities")

	for list.NotDone() {
		log.WithField("count", len(list.Values())).Trace("list not done")
		for _, g := range list.Values() {
			log.Trace("adding entity to list")

			resources = append(resources, &PrivateDNSZone{
				BaseResource: &BaseResource{
					Region:         g.Location,
					ResourceGroup:  azure.GetResourceGroupFromID(*g.ID),
					SubscriptionID: ptr.String(opts.SubscriptionID),
				},
				client: client,
				Name:   g.Name,
				Tags:   g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
