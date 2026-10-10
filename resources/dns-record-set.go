package resources

import (
	"context"
	"errors"
	"path"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const DNSRecordSetResource = "DNSRecordSet"

func init() {
	registry.Register(&registry.Registration{
		Name:     DNSRecordSetResource,
		Scope:    azure.SubscriptionScope,
		Resource: &DNSRecordSet{},
		Lister:   &DNSRecordSetLister{},
	})
}

type DNSRecordSetLister struct{}

func (l DNSRecordSetLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", DNSRecordSetResource).WithField("s", opts.SubscriptionID)

	zonesClient, err := armdns.NewZonesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	client, err := armdns.NewRecordSetsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list dns zones")

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

			pager := client.NewListAllByDNSZonePager(*resourceGroup, *zone.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if azure.IsNotFound(err) {
					// the zone went with its resource group since it was listed
					break
				}
				if err != nil {
					return nil, err
				}

				for _, entity := range page.Value {
					recordSet := &DNSRecordSet{
						BaseResource: &BaseResource{
							Region:         ptr.String("global"),
							ResourceGroup:  resourceGroup,
							SubscriptionID: ptr.String(opts.SubscriptionID),
						},
						client: client,
						ID:     entity.ID,
						Name:   entity.Name,
						Zone:   zone.Name,
						// ARM returns Microsoft.Network/dnszones/NS
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

type DNSRecordSet struct {
	*BaseResource `property:",inline"`

	client *armdns.RecordSetsClient
	ID     *string            `description:"The ID of the record set."`
	Name   *string            `description:"The name of the record set relative to the zone, @ for the apex."`
	Zone   *string            `description:"The DNS zone the record set belongs to."`
	Type   *string            `description:"The record type, e.g. A, CNAME, NS."`
	TTL    *int64             `description:"The TTL of the records."`
	Tags   map[string]*string `description:"The metadata of the record set."`
}

func (r *DNSRecordSet) Filter() error {
	if ptr.ToString(r.Type) == string(armdns.RecordTypeSOA) {
		return errors.New("the SOA record is managed by the zone")
	}
	if ptr.ToString(r.Name) == "@" && ptr.ToString(r.Type) == string(armdns.RecordTypeNS) {
		return errors.New("the apex NS record is managed by the zone")
	}
	return nil
}

func (r *DNSRecordSet) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Zone, *r.Name, armdns.RecordType(*r.Type), nil)
	return azure.IgnoreNotFound(err)
}

func (r *DNSRecordSet) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

// UniqueKey tells record sets apart that share a name: a delegation and the apex of its child zone, or several
// record types of one name.
func (r *DNSRecordSet) UniqueKey() string {
	return ptr.ToString(r.ID)
}

func (r *DNSRecordSet) String() string {
	name := ptr.ToString(r.Name) + "." + ptr.ToString(r.Zone)
	if ptr.ToString(r.Name) == "@" {
		name = ptr.ToString(r.Zone)
	}
	return name + " " + ptr.ToString(r.Type)
}
