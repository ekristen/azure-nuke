package resources

import (
	"context"
	"errors"
	"path"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns/v2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const PrivateDNSRecordSetResource = "PrivateDNSRecordSet"

func init() {
	registry.Register(&registry.Registration{
		Name:     PrivateDNSRecordSetResource,
		Scope:    azure.SubscriptionScope,
		Resource: &PrivateDNSRecordSet{},
		Lister:   &PrivateDNSRecordSetLister{},
	})
}

type PrivateDNSRecordSetLister struct{}

func (l PrivateDNSRecordSetLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", PrivateDNSRecordSetResource).WithField("s", opts.SubscriptionID)

	zonesClient, err := armprivatedns.NewPrivateZonesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	client, err := armprivatedns.NewRecordSetsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list private dns zones")

	zonesPager := zonesClient.NewListPager(nil)
	for zonesPager.More() {
		zonesPage, err := zonesPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, zone := range zonesPage.Value {
			resourceGroup := azure.GetResourceGroupFromID(ptr.ToString(zone.ID))
			if resourceGroup == nil {
				continue
			}

			pager := client.NewListPager(*resourceGroup, *zone.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if azure.IsNotFound(err) {
					break
				}
				if err != nil {
					return nil, err
				}

				for _, entity := range page.Value {
					recordSet := &PrivateDNSRecordSet{
						BaseResource: &BaseResource{
							Region:         ptr.String("global"),
							ResourceGroup:  resourceGroup,
							SubscriptionID: ptr.String(opts.SubscriptionID),
						},
						client: client,
						ID:     entity.ID,
						Name:   entity.Name,
						Zone:   zone.Name,
						Type: ptr.String(path.Base(ptr.ToString(entity.Type))),
					}
					if entity.Properties != nil {
						recordSet.TTL = entity.Properties.TTL
						recordSet.Tags = entity.Properties.Metadata
					}
					resources = append(resources, recordSet)
				}
			}
		}
	}

	log.Trace("done")

	return resources, nil
}

// ---------------------------------------------

type PrivateDNSRecordSet struct {
	*BaseResource `property:",inline"`

	client *armprivatedns.RecordSetsClient
	ID     *string            `description:"The ID of the record set."`
	Name   *string            `description:"The name of the record set relative to the zone, @ for the apex."`
	Zone   *string            `description:"The private DNS zone the record set belongs to."`
	Type   *string            `description:"The record type, e.g. A, CNAME, TXT."`
	TTL    *int64             `description:"The TTL of the records."`
	Tags   map[string]*string `description:"The metadata of the record set."`
}

func (r *PrivateDNSRecordSet) Filter() error {
	if ptr.ToString(r.Type) == string(armprivatedns.RecordTypeSOA) {
		return errors.New("the SOA record is managed by the zone")
	}
	return nil
}

func (r *PrivateDNSRecordSet) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Zone, armprivatedns.RecordType(*r.Type), *r.Name, nil)
	return azure.IgnoreNotFound(err)
}

func (r *PrivateDNSRecordSet) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PrivateDNSRecordSet) UniqueKey() string {
	return ptr.ToString(r.ID)
}

func (r *PrivateDNSRecordSet) String() string {
	name := ptr.ToString(r.Name) + "." + ptr.ToString(r.Zone)
	if ptr.ToString(r.Name) == "@" {
		name = ptr.ToString(r.Zone)
	}
	return name + " " + ptr.ToString(r.Type)
}
