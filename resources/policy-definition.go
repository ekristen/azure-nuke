package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy/v2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const PolicyDefinitionResource = "PolicyDefinition"

func init() {
	registry.Register(&registry.Registration{
		Name:     PolicyDefinitionResource,
		Scope:    azure.SubscriptionScope,
		Resource: &PolicyDefinition{},
		Lister:   &PolicyDefinitionLister{},
	})
}

type PolicyDefinition struct {
	*BaseResource `property:",inline"`

	client      *armpolicy.DefinitionsClient
	Name        *string
	DisplayName string
	PolicyType  string `property:"name=Type"`
}

func (r *PolicyDefinition) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.Name, nil)
	return err
}

func (r *PolicyDefinition) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PolicyDefinition) String() string {
	return *r.Name
}

type PolicyDefinitionLister struct{}

func (l PolicyDefinitionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", PolicyDefinitionResource).WithField("s", opts.SubscriptionID)

	client, err := armpolicy.NewDefinitionsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds,
		&arm.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				APIVersion: "2023-04-01",
			},
		})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list policy definitions")

	pager := client.NewListPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		log.Trace("listing policy definitions")

		for _, g := range page.Value {
			// Filtering out BuiltIn Policy Definitions, because otherwise it needlessly adds 3000+
			// resources that have to get filtered out later. This instead does it optimistically here.
			// Ideally we'd be able to use filter above, but it does not work. Thanks, Azure. :facepalm:
			if g.Properties != nil && g.Properties.PolicyType != nil {
				policyType := string(*g.Properties.PolicyType)
				if policyType == "BuiltIn" || policyType == "Static" {
					continue
				}
			}

			policyType := ""
			if g.Properties != nil && g.Properties.PolicyType != nil {
				policyType = string(*g.Properties.PolicyType)
			}

			displayName := ""
			if g.Properties != nil && g.Properties.DisplayName != nil {
				displayName = *g.Properties.DisplayName
			}

			resources = append(resources, &PolicyDefinition{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client:      client,
				Name:        g.Name,
				DisplayName: displayName,
				PolicyType:  policyType,
			})
		}
	}

	log.WithField("total", len(resources)).Trace("done")

	return resources, nil
}
