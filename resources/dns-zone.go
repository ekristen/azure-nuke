package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type DNSZone struct {
	client   dns.ZonesClient
	name     *string
	location *string
	rg       *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "DNSZone",
		Scope:  resource.ResourceGroup,
		Lister: ListDNSZone,
	})
}

func ListDNSZone(opts resource.ListerOpts) ([]resource.Resource, error) {
	log := logrus.WithFields(logrus.Fields{
		"subscription": opts.SubscriptionId,
		"handler":      "ListDNSZone",
	})

	log.Trace("start")

	client := dns.NewZonesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	ctx := context.Background()

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

	return resources, nil
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
