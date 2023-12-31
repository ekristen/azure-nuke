package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
)

func init() {
	resource.Register(resource.Registration{
		Name:   "ApplicationSecret",
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
	opts nuke.ListerOpts
}

func (l ApplicationSecretLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l ApplicationSecretLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = l.opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	logrus.Trace("attempting to list service principals")

	ctx := context.Background()

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	logrus.Trace("listing ....")

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

	return resources, nil
}
