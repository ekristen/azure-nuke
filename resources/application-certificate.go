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

const ApplicationCertificateResource = "ApplicationCertificate"

func init() {
	resource.Register(&resource.Registration{
		Name:   ApplicationCertificateResource,
		Scope:  nuke.Tenant,
		Lister: &ApplicationCertificateLister{},
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

func (r *ApplicationCertificate) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.id)
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
}

func (l ApplicationCertificateLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", ApplicationCertificateResource).WithField("s", opts.SubscriptionId)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list application certificates")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application certificate")

	for _, entity := range *entities {
		for _, cred := range *entity.KeyCredentials {
			resources = append(resources, &ApplicationCertificate{
				client: client,
				id:     cred.KeyId,
				name:   cred.DisplayName,
				appId:  entity.ID(),
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
