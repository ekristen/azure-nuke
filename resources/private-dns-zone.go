package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const PrivateDNSZoneResource = "PrivateDNSZone"

func init() {
	registry.Register(&registry.Registration{
		Name:   PrivateDNSZoneResource,
		Scope:  nuke.Subscription,
		Lister: &PrivateDNSZoneLister{},
	})
}

type PrivateDNSZone struct {
	client   privatedns.PrivateZonesClient
	name     *string
	location *string
	rg       *string
}

func (r *PrivateDNSZone) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.rg, *r.name, "")
	return err
}

func (r *PrivateDNSZone) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("ResourceGroup", *r.rg)

	return properties
}

func (r *PrivateDNSZone) String() string {
	return *r.name
}

type PrivateDNSZoneLister struct {
}

func (l PrivateDNSZoneLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithFields(logrus.Fields{
		"r": PrivateDNSZoneResource,
		"s": opts.SubscriptionId,
	})

	log.Trace("start")

	client := privatedns.NewPrivateZonesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

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
