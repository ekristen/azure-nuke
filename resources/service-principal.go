package resources

import (
	"context"
	"fmt"
	"github.com/aws/smithy-go/ptr"
	"strings"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type ServicePrincipal struct {
	client   *msgraph.ServicePrincipalsClient
	id       *string
	name     *string
	appOwner *string
	spType   *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "ServicePrincipal",
		Scope:  resource.Tenant,
		Lister: ListServicePrincipal,
	})
}

func ListServicePrincipal(opts resource.ListerOpts) ([]resource.Resource, error) {
	log := logrus.
		WithField("resource", "ServicePrincipal").
		WithField("scope", resource.Subscription).
		WithField("subscription", opts.SubscriptionId)

	client := msgraph.NewServicePrincipalsClient()
	client.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list service principals")

	ctx := context.TODO()
	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing entities")

	for _, entity := range *entities {
		resources = append(resources, &ServicePrincipal{
			client:   client,
			id:       entity.ID(),
			name:     entity.DisplayName,
			appOwner: entity.AppOwnerOrganizationId,
			spType:   entity.ServicePrincipalType,
		})
	}

	return resources, nil
}

func (r *ServicePrincipal) Filter() error {
	if ptr.ToString(r.spType) == "ManagedIdentity" {
		return fmt.Errorf("cannot delete managed service principals")
	}

	if ptr.ToString(r.appOwner) == "f8cdef31-a31e-4b4a-93e4-5f571e91255a" {
		return fmt.Errorf("cannot delete built-in service principals")
	}

	if ptr.ToString(r.name) == "O365 LinkedIn Connection" {
		return fmt.Errorf("cannot delete built-in service principals")
	}

	if strings.Contains(ptr.ToString(r.name), "securityOperators/Defender") {
		return fmt.Errorf("cannot delete defender linked service principals")
	}

	return nil
}

func (r *ServicePrincipal) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.id)
	return err
}

func (r *ServicePrincipal) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("AppOwnerId", r.appOwner)
	properties.Set("ServicePrincipalType", r.spType)

	return properties
}

func (r *ServicePrincipal) String() string {
	return ptr.ToString(r.id)
}
