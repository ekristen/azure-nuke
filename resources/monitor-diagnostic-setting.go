package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const MonitorDiagnosticSettingResource = "MonitorDiagnosticSetting"

func init() {
	registry.Register(&registry.Registration{
		Name:     MonitorDiagnosticSettingResource,
		Scope:    azure.SubscriptionScope,
		Lister:   &MonitorDiagnosticSettingLister{},
		Resource: &MonitorDiagnosticSetting{},
	})
}

type MonitorDiagnosticSetting struct {
	*BaseResource `property:",inline"`

	client *armmonitor.DiagnosticSettingsClient
	Name   *string
}

func (r *MonitorDiagnosticSetting) Remove(ctx context.Context) error {
	resourceURI := fmt.Sprintf("/subscriptions/%s", *r.SubscriptionID)
	_, err := r.client.Delete(ctx, resourceURI, *r.Name, nil)
	return err
}

func (r *MonitorDiagnosticSetting) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *MonitorDiagnosticSetting) String() string {
	return *r.Name
}

type MonitorDiagnosticSettingLister struct{}

func (l MonitorDiagnosticSettingLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", MonitorDiagnosticSettingResource).WithField("s", opts.SubscriptionID)

	client, err := armmonitor.NewDiagnosticSettingsClient(opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	resourceURI := fmt.Sprintf("/subscriptions/%s", opts.SubscriptionID)

	log.Trace("attempting to list diagnostic settings")

	pager := client.NewListPager(resourceURI, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		if page.Value != nil {
			for _, ds := range page.Value {
				resources = append(resources, &MonitorDiagnosticSetting{
					BaseResource: &BaseResource{
						Region:         ptr.String("global"),
						SubscriptionID: &opts.SubscriptionID,
					},
					client: client,
					Name:   ds.Name,
				})
			}
		}
	}

	log.Trace("done")

	return resources, nil
}
