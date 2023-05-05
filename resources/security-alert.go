package resources

import (
	"context"
	"fmt"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/types"
	"github.com/sirupsen/logrus"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
)

type SecurityAlert struct {
	client      security.AlertsClient
	id          string
	name        string
	displayName string
	location    string
	status      string
}

var SecurityAlertLocation = "/Microsoft.Security/locations/(?P<location>.*)/alerts/"

func init() {
	resource.RegisterV2(resource.Registration{
		Name:   "SecurityAlert",
		Scope:  resource.Subscription,
		Lister: ListSecurityAlert,
	})
}

func ListSecurityAlert(opts resource.ListerOpts) ([]resource.Resource, error) {
	log := logrus.
		WithField("resource", "SecurityAlert").
		WithField("scope", resource.Subscription).
		WithField("subscription", opts.SubscriptionId)

	log.Trace("creating client")

	locationRe := regexp.MustCompile(SecurityAlertLocation)

	client := security.NewAlertsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("listing resources")

	ctx := context.TODO()
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
				location:    matches[1],
				status:      string(g.AlertProperties.Status),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *SecurityAlert) Filter() error {
	if r.status == "Dismissed" {
		return fmt.Errorf("alert already dismissed")
	}

	return nil
}

func (r *SecurityAlert) Remove() error {
	// Note: we cannot actually remove alerts :(
	// So we just have to dismiss them instead
	_, err := r.client.UpdateSubscriptionLevelStateToDismiss(context.TODO(), r.location, r.name)
	return err
}

func (r *SecurityAlert) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("DisplayName", r.displayName)
	properties.Set("Location", r.location)
	properties.Set("Status", r.status)

	return properties
}

func (r *SecurityAlert) String() string {
	return r.name
}
