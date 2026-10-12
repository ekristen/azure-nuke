package resources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const SecurityAlertResource = "SecurityAlert"

const SecurityAlertLocation = "/Microsoft.Security/locations/(?P<region>.*)/alerts/"

func init() {
	registry.Register(&registry.Registration{
		Name:     SecurityAlertResource,
		Scope:    azure.SubscriptionScope,
		Resource: &SecurityAlert{},
		Lister:   &SecurityAlertsLister{},
	})
}

type SecurityAlert struct {
	*BaseResource `property:",inline"`

	client      *armsecurity.AlertsClient
	ID          string
	Name        string
	DisplayName string
	Status      string
}

func (r *SecurityAlert) Filter() error {
	if r.Status == "Dismissed" {
		return fmt.Errorf("alert already dismissed")
	}

	return nil
}

func (r *SecurityAlert) Remove(ctx context.Context) error {
	// Note: we cannot actually remove alerts :(
	// So we just have to dismiss them instead
	_, err := r.client.UpdateSubscriptionLevelStateToDismiss(ctx, *r.Region, r.Name, nil)
	return err
}

func (r *SecurityAlert) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SecurityAlert) String() string {
	return r.Name
}

// ------------------------------------

type SecurityAlertsLister struct{}

func (l SecurityAlertsLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.
		WithField("r", SecurityAlertResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	locationRe := regexp.MustCompile(SecurityAlertLocation)

	client, err := armsecurity.NewAlertsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			matches := locationRe.FindStringSubmatch(ptr.ToString(entity.ID))

			var displayName, status string
			if entity.Properties != nil {
				displayName = ptr.ToString(entity.Properties.AlertDisplayName)
				if entity.Properties.Status != nil {
					status = string(*entity.Properties.Status)
				}
			}

			resources = append(resources, &SecurityAlert{
				BaseResource: &BaseResource{
					Region: ptr.String(matches[1]),
				},
				client:      client,
				ID:          ptr.ToString(entity.ID),
				Name:        ptr.ToString(entity.Name),
				DisplayName: displayName,
				Status:      status,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
