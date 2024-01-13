package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const ApplicationSecretResource = "ApplicationSecret"

func init() {
	resource.Register(resource.Registration{
		Name:   ApplicationSecretResource,
		Scope:  nuke.Tenant,
		Lister: ApplicationSecretLister{},
	})
}

type ApplicationSecret struct {
	client  *msgraph.ApplicationsClient
	id      *string
	name    *string
	appId   *string
	appName *string
}

func (r *ApplicationSecret) Filter() error {
	return nil
}

func (r *ApplicationSecret) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.id)
	return err
}

func (r *ApplicationSecret) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("AppName", r.appName)

	return properties
}

func (r *ApplicationSecret) String() string {
	return *r.id
}

type ApplicationSecretLister struct {
}

func (l ApplicationSecretLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", ApplicationSecretResource).WithField("s", opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list application secrets")

	ctx := context.TODO()

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application secrets")

	for _, entity := range *entities {
		for _, cred := range *entity.PasswordCredentials {
			resources = append(resources, &ApplicationSecret{
				client:  client,
				id:      cred.KeyId,
				name:    cred.DisplayName,
				appId:   entity.ID(),
				appName: entity.DisplayName,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
