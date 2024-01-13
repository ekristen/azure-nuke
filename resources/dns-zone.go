package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const DNSZoneResource = "DNSZone"

func init() {
	resource.Register(resource.Registration{
		Name:   DNSZoneResource,
		Scope:  nuke.ResourceGroup,
		Lister: DNSZoneLister{},
	})
}

type DNSZoneLister struct {
}

func (l DNSZoneLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithFields(logrus.Fields{
		"r": DNSZoneResource,
		"s": opts.SubscriptionId,
	})

	log.Trace("start")

	client := dns.NewZonesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	ctx := context.TODO()

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

type DNSZone struct {
	client   dns.ZonesClient
	name     *string
	location *string
	rg       *string
}

func (r *DNSZone) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.rg, *r.name, "")
	return err
}

func (r *DNSZone) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *DNSZone) String() string {
	return *r.name
}
