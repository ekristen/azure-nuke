package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ServicePrincipalResource = "ServicePrincipal"

func init() {
	registry.Register(&registry.Registration{
		Name:     ServicePrincipalResource,
		Scope:    azure.TenantScope,
		Resource: &ServicePrincipal{},
		Lister:   &ServicePrincipalsLister{},
	})
}

type ServicePrincipal struct {
	*BaseResource `property:",inline"`

	client   *msgraph.ServicePrincipalsClient
	ID       *string
	Name     *string
	AppOwner *string `property:"name=AppOwnerId"`
	SPType   *string `property:"name=ServicePrincipalType"`
}

func (r *ServicePrincipal) Filter() error {
	if ptr.ToString(r.SPType) == "ManagedIdentity" {
		return fmt.Errorf("cannot delete managed service principals")
	}

	if ptr.ToString(r.AppOwner) == "f8cdef31-a31e-4b4a-93e4-5f571e91255a" {
		return fmt.Errorf("cannot delete built-in service principals")
	}

	if ptr.ToString(r.Name) == "O365 LinkedIn Connection" {
		return fmt.Errorf("cannot delete built-in service principals")
	}

	if strings.Contains(ptr.ToString(r.Name), "securityOperators/Defender") {
		return fmt.Errorf("cannot delete defender linked service principals")
	}

	return nil
}

func (r *ServicePrincipal) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ID)
	return err
}

func (r *ServicePrincipal) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ServicePrincipal) String() string {
	return ptr.ToString(r.Name)
}

// -------------------------------------------------------------

type ServicePrincipalsLister struct {
}

func (l ServicePrincipalsLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ServicePrincipalResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewServicePrincipalsClient()
	client.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	client.BaseClient.DisableRetries = true

	log.Trace("attempting to list service principals")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing resource start")

	for i := range *entities {
		entity := &(*entities)[i]

		// Filtering out Microsoft owned Service Principals, because otherwise it needlessly adds 3000+
		// resources that have to get filtered out later. This instead does it optimistically here.
		// Ideally we'd be able to use odata.Query above, but it's not supported by the graph at this time.
		if ptr.ToString(entity.AppOwnerOrganizationId) == "f8cdef31-a31e-4b4a-93e4-5f571e91255a" {
			continue
		}

		resources = append(resources, &ServicePrincipal{
			BaseResource: &BaseResource{
				Region: ptr.String("global"),
			},
			client:   client,
			ID:       entity.ID(),
			Name:     entity.DisplayName,
			AppOwner: entity.AppOwnerOrganizationId,
			SPType:   entity.ServicePrincipalType,
		})
	}

	log.Trace("listing resources end")

	return resources, nil
}
