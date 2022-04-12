package resources

import (
	"context"
	"fmt"

	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/manicminer/hamilton/odata"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

type ServicePrincipal struct {
	client   *msgraph.ServicePrincipalsClient
	id       *string
	name     *string
	appOwner *string
}

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "ServicePrincipal",
		Scope:  resource.Tenant,
		Lister: ListServicePrincipal,
	})
}

func ListServicePrincipal(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	wrappedAuth, err := auth.NewAutorestAuthorizerWrapper(opts.Authorizers.Graph)
	if err != nil {
		return nil, err
	}

	client := msgraph.NewServicePrincipalsClient(opts.TenantId)
	client.BaseClient.Authorizer = wrappedAuth
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list service principals")

	ctx := context.Background()

	entites, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

	for _, entity := range *entites {
		resources = append(resources, &ServicePrincipal{
			client:   client,
			id:       entity.ID,
			name:     entity.DisplayName,
			appOwner: entity.AppOwnerOrganizationId,
		})
	}

	return resources, nil
}

func (r *ServicePrincipal) Filter() error {
	if r.appOwner != nil && *r.appOwner == "f8cdef31-a31e-4b4a-93e4-5f571e91255a" {
		return fmt.Errorf("cannot delete built-in service principals")
	}
	if r.name != nil && *r.name == "O365 LinkedIn Connection" {
		return fmt.Errorf("cannot delete built-in service principals")
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

	return properties
}

func (r *ServicePrincipal) String() string {
	return *r.id
}
