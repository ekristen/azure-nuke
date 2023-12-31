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
		Name:   "ApplicationFederatedCredential",
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
	opts nuke.ListerOpts
}

func (l ApplicationFederatedCredentialLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l ApplicationFederatedCredentialLister) List() ([]resource.Resource, error) {
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
