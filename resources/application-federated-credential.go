package resources

import (
	"context"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
)

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "ApplicationFederatedCredential",
		Scope:  resource.Tenant,
		Lister: ListApplicationFederatedCredential,
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

func ListApplicationFederatedCredential(opts resource.ListerOpts) ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
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

	return resources, nil
}
