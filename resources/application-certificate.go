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

const ApplicationCertificateResource = "ApplicationCertificate"

func init() {
	registry.Register(&registry.Registration{
		Name:     ApplicationCertificateResource,
		Scope:    azure.TenantScope,
		Resource: &ApplicationCertificate{},
		Lister:   &ApplicationCertificateLister{},
	})
}

type ApplicationCertificate struct {
	*BaseResource `property:",inline"`

	client *msgraph.ApplicationsClient
	ID     *string
	Name   *string
	AppID  *string
}

func (r *ApplicationCertificate) Filter() error {
	return nil
}

func (r *ApplicationCertificate) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ID)
	return err
}

func (r *ApplicationCertificate) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ApplicationCertificate) String() string {
	return *r.ID
}

type ApplicationCertificateLister struct{}

func (l ApplicationCertificateLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ApplicationCertificateResource).WithField("s", opts.SubscriptionID)

	client := msgraph.NewApplicationsClient()
	client.BaseClient.Authorizer = opts.Authorizers.Graph
	client.BaseClient.DisableRetries = true

	log.Trace("attempting to list application certificates")

	entities, _, err := client.List(ctx, odata.Query{})
	if err != nil {
		return nil, err
	}

	log.Trace("listing application certificate")

	for i := range *entities {
		entity := &(*entities)[i]

		for _, cred := range *entity.KeyCredentials {
			resources = append(resources, &ApplicationCertificate{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client: client,
				ID:     cred.KeyId,
				Name:   cred.DisplayName,
				AppID:  entity.ID(),
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
