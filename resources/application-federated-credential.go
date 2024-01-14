package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const ApplicationFederatedCredentialResource = "ApplicationFederatedCredential"

func init() {
	resource.Register(resource.Registration{
		Name:   ApplicationFederatedCredentialResource,
		Scope:  nuke.Tenant,
		Lister: ApplicationFederatedCredentialLister{},
	})
}

type ApplicationFederatedCredential struct {
	client     *msgraph.ApplicationsClient
	id         *string
	name       *string
	appId      *string
	uniqueName *string
}

func (r *ApplicationFederatedCredential) Filter() error {
	return nil
}

func (r *ApplicationFederatedCredential) Remove() error {
	_, err := r.client.DeleteFederatedIdentityCredential(context.TODO(), *r.appId, *r.id)
	return err
}

func (r *ApplicationFederatedCredential) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("AppID", *r.appId)
	properties.Set("AppUniqueName", r.uniqueName)

	return properties
}

func (r *ApplicationFederatedCredential) String() string {
	return *r.id
}

type ApplicationFederatedCredentialLister struct {
}

func (l ApplicationFederatedCredentialLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", ApplicationFederatedCredentialResource).WithField("s", opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list application federated creds")

	ctx := context.TODO()

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application federated creds")

	for _, entity := range *entities {
		creds, _, err := client.ListFederatedIdentityCredentials(ctx, *entity.ID(), odata.Query{})
		if err != nil {
			return nil, err
		}
		for _, cred := range *creds {
			resources = append(resources, &ApplicationFederatedCredential{
				client:     client,
				id:         cred.ID,
				name:       cred.Name,
				appId:      entity.ID(),
				uniqueName: entity.UniqueName,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
