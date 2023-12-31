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
		Name:   "ApplicationCertificate",
		Scope:  nuke.Tenant,
		Lister: ApplicationCertificateLister{},
	})
}

type ApplicationCertificate struct {
	client *msgraph.ApplicationsClient
	id     *string
	name   *string
	appId  *string
}

func (r *ApplicationCertificate) Filter() error {
	return nil
}

func (r *ApplicationCertificate) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.id)
	return err
}

func (r *ApplicationCertificate) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)

	return properties
}

func (r *ApplicationCertificate) String() string {
	return *r.id
}

type ApplicationCertificateLister struct {
	opts nuke.ListerOpts
}

func (l ApplicationCertificateLister) SetOptions(opts interface{}) {
	l.opts = opts.(nuke.ListerOpts)
}

func (l ApplicationCertificateLister) List() ([]resource.Resource, error) {
	logrus.Tracef("subscription id: %s", l.opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = l.opts.Authorizers.Graph
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
		for _, cred := range *entity.KeyCredentials {
			resources = append(resources, &ApplicationCertificate{
				client: client,
				id:     cred.KeyId,
				name:   cred.DisplayName,
				appId:  entity.ID(),
			})
		}
	}

	return resources, nil
}
