package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const DNSZoneResource = "DNSZone"

func init() {
	registry.Register(&registry.Registration{
		Name:   DNSZoneResource,
		Scope:  nuke.ResourceGroup,
		Lister: &DNSZoneLister{},
	})
}

type DNSZoneLister struct {
}

func (l DNSZoneLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithFields(logrus.Fields{
		"r": DNSZoneResource,
		"s": opts.SubscriptionID,
	})

	log.Trace("start")

	client := dns.NewZonesClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup, nil)
	if err != nil {
		log.WithError(err).Error("unable to list")
		return nil, err
	}

	log.Trace("listing entities")

	for list.NotDone() {
		log.WithField("count", len(list.Values())).Trace("list not done")
		for _, g := range list.Values() {
			log.Trace("adding entity to list")
			resources = append(resources, &DNSZone{
				rg: &opts.ResourceGroup,

				client: client,
				name:   g.Name,
				region: g.Location,
				tags:   g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}

type DNSZone struct {
	client dns.ZonesClient
	name   *string
	region *string
	rg     *string
	tags   map[string]*string
}

func (r *DNSZone) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.rg, *r.name, "")
	return err
}

func (r *DNSZone) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("ResourceGroup", r.rg)
	properties.Set("Region", r.region)

	for k, v := range r.tags {
		properties.SetTag(&k, v)
	}

	return properties
}

func (r *DNSZone) String() string {
	return *r.name
}
