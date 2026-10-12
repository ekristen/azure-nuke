package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ApplicationSecretResource = "ApplicationSecret"

func init() {
	registry.Register(&registry.Registration{
		Name:     ApplicationSecretResource,
		Scope:    azure.TenantScope,
		Resource: &ApplicationSecret{},
		Lister:   &ApplicationSecretLister{},
	})
}

type ApplicationSecret struct {
	*BaseResource `property:",inline"`

	client  *msgraph.ApplicationsClient
	KeyID   *string `description:"The unique ID of the Application Secret Key"`
	Name    *string `description:"The display name of the Application Secret"`
	AppID   *string `description:"The unique ID of the Application to which the secret belongs"`
	AppName *string `description:"The display name of the Application to which the secret belongs"`
}

func (r *ApplicationSecret) Filter() error {
	return nil
}

func (r *ApplicationSecret) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.KeyID)
	return err
}

func (r *ApplicationSecret) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ApplicationSecret) String() string {
	return fmt.Sprintf("%s -> %s", *r.AppName, *r.KeyID)
}

type ApplicationSecretLister struct{}

func (l ApplicationSecretLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ApplicationSecretResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	log.Trace("attempting to list application secrets")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application secrets")

	for i := range *entities {
		entity := &(*entities)[i]

		for _, cred := range *entity.PasswordCredentials {
			resources = append(resources, &ApplicationSecret{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client:  client,
				KeyID:   cred.KeyId,
				Name:    cred.DisplayName,
				AppID:   entity.ID(),
				AppName: entity.DisplayName,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
