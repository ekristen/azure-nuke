package resources

import (
	"context"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/resource-manager/insights/2021-05-01-preview/diagnosticsettings"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"

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

	client diagnosticsettings.DiagnosticSettingsClient
	id     *string
	Name   *string
}

func (r *MonitorDiagnosticSetting) Remove(ctx context.Context) error {
	// Note: we do this because the regex on the ParsedScope is case-sensitive, and the ID returned from
	// the API is not always consistent with the casing.
	cleanedID := strings.ReplaceAll(*r.id, "microsoft.insights", "Microsoft.Insights")

	id, err := diagnosticsettings.ParseScopedDiagnosticSettingID(cleanedID)
	if err != nil {
		return err
	}

	if _, err := r.client.Delete(ctx, *id); err != nil {
		return err
	}

	return nil
}

func (r *MonitorDiagnosticSetting) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *MonitorDiagnosticSetting) String() string {
	return *r.Name
}

type MonitorDiagnosticSettingLister struct {
}

func (l MonitorDiagnosticSettingLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", MonitorDiagnosticSettingResource).WithField("s", opts.SubscriptionID)

	endpoint, _ := environments.AzurePublic().ResourceManager.Endpoint()
	client := diagnosticsettings.NewDiagnosticSettingsClientWithBaseURI(*endpoint)
	client.Client.Authorizer = opts.Authorizers.Management

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list diagnostic settings")
	list, err := client.List(ctx, commonids.NewScopeID(commonids.NewSubscriptionID(opts.SubscriptionID).ID()))
	if err != nil {
		return nil, err
	}

	log.Trace("listing diagnostic settings")

	for _, ds := range *list.Model.Value {
		resources = append(resources, &MonitorDiagnosticSetting{
			BaseResource: &BaseResource{
				Region:         ptr.String("global"),
				SubscriptionID: &opts.SubscriptionID,
			},
			client: client,
			id:     ds.Id,
			Name:   ds.Name,
		})
	}

	log.Trace("done")

	return resources, nil
}
