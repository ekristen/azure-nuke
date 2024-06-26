package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const AzureADUserResource = "AzureADUser"

func init() {
	registry.Register(&registry.Registration{
		Name:     AzureADUserResource,
		Scope:    azure.TenantScope,
		Resource: &AzureADUser{},
		Lister:   &AzureADUserLister{},
		DependsOn: []string{
			AzureAdGroupResource,
		},
	})
}

type AzureADUserLister struct {
}

func (l AzureADUserLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", AzureADUserResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewUsersClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	log.Trace("attempting to list azure ad users")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for i := range *entities {
		entity := &(*entities)[i]

		resources = append(resources, &AzureADUser{
			BaseResource: &BaseResource{
				Region: ptr.String("global"),
			},
			client: client,
			ID:     entity.ID(),
			Name:   entity.DisplayName,
			UPN:    entity.UserPrincipalName,
		})
	}

	return resources, nil
}

type AzureADUser struct {
	*BaseResource `property:",inline"`

	client *msgraph.UsersClient
	ID     *string `description:"The ID of the Entra ID User"`
	Name   *string `description:"The DisplayName of the Entra ID User"`
	UPN    *string `description:"This is the user principal name of the Entra ID user, usually in the format of email"`
}

func (r *AzureADUser) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ID)
	return err
}

func (r *AzureADUser) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *AzureADUser) String() string {
	return *r.Name
}
