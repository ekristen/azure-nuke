package resources

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const SecurityAlertResource = "SecurityAlert"

const SecurityAlertLocation = "/Microsoft.Security/locations/(?P<region>.*)/alerts/"

func init() {
	registry.Register(&registry.Registration{
		Name:   SecurityAlertResource,
		Scope:  nuke.Subscription,
		Lister: &SecurityAlertsLister{},
	})
}

type SecurityAlert struct {
	client      security.AlertsClient
	id          string
	name        string
	displayName string
	region      string
	status      string
}

func (r *SecurityAlert) Filter() error {
	if r.status == "Dismissed" {
		return fmt.Errorf("alert already dismissed")
	}

	return nil
}

func (r *SecurityAlert) Remove(ctx context.Context) error {
	// Note: we cannot actually remove alerts :(
	// So we just have to dismiss them instead
	_, err := r.client.UpdateSubscriptionLevelStateToDismiss(ctx, r.region, r.name)
	return err
}

func (r *SecurityAlert) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("DisplayName", r.displayName)
	properties.Set("Region", r.region)
	properties.Set("Status", r.status)

	return properties
}

func (r *SecurityAlert) String() string {
	return r.name
}

// ------------------------------------

type SecurityAlertsLister struct {
}

func (l SecurityAlertsLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.
		WithField("r", SecurityAlertResource).
		WithField("s", opts.SubscriptionID)

	log.Trace("creating client")

	locationRe := regexp.MustCompile(SecurityAlertLocation)

	client := security.NewAlertsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	list, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	for list.NotDone() {
		log.Trace("listing not done")
		for _, g := range list.Values() {
			matches := locationRe.FindStringSubmatch(ptr.ToString(g.ID))
			resources = append(resources, &SecurityAlert{
				client:      client,
				id:          *g.ID,
				name:        *g.Name,
				displayName: ptr.ToString(g.AlertDisplayName),
				region:      matches[1],
				status:      string(g.AlertProperties.Status),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
