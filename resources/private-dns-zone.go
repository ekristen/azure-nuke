package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const PrivateDNSZoneResource = "PrivateDNSZone"

func init() {
	resource.Register(resource.Registration{
		Name:   PrivateDNSZoneResource,
		Scope:  nuke.Subscription,
		Lister: PrivateDNSZoneLister{},
	})
}

type PrivateDNSZone struct {
	client   privatedns.PrivateZonesClient
	name     *string
	location *string
	rg       *string
}

func (r *PrivateDNSZone) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.rg, *r.name, "")
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

func (l PrivateDNSZoneLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

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

	ctx := context.TODO()

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
