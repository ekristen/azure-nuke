package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const DNSZoneResource = "DNSZone"

func init() {
	registry.Register(&registry.Registration{
		Name:     DNSZoneResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &DNSZone{},
		Lister:   &DNSZoneLister{},
	})
}

type DNSZoneLister struct{}

func (l DNSZoneLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithFields(logrus.Fields{
		"r": DNSZoneResource,
		"s": opts.SubscriptionID,
	})

	log.Trace("start")

	client, err := armdns.NewZonesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("listing entities")

	pager := client.NewListByResourceGroupPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.WithError(err).Error("unable to list")
			return nil, err
		}

		for _, g := range page.Value {
			resources = append(resources, &DNSZone{
				BaseResource: &BaseResource{
					Region:         g.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				Name:   g.Name,
				Tags:   g.Tags,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}

type DNSZone struct {
	*BaseResource `property:",inline"`

	client   *armdns.ZonesClient
	removeOp *runtime.Poller[armdns.ZonesClientDeleteResponse]
	Name     *string
	Tags     map[string]*string
}

func (r *DNSZone) Remove(ctx context.Context) error {
	r.removeOp = nil

	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return azure.IgnoreNotFound(err)
	}

	r.removeOp = poller
	return nil
}

func (r *DNSZone) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *DNSZone) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DNSZone) String() string {
	return *r.Name
}
