package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "PrivateDNSZone",
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
	opts nuke.ListerOpts
}

func (l PrivateDNSZoneLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l PrivateDNSZoneLister) List() ([]resource.Resource, error) {
	log := logrus.WithFields(logrus.Fields{
		"subscription": l.opts.SubscriptionId,
		"handler":      "ListPrivateDNSZone",
	})

	log.Trace("start")

	client := privatedns.NewPrivateZonesClient(l.opts.SubscriptionId)
	client.Authorizer = l.opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	ctx := context.Background()

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
				rg:       &l.opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}
