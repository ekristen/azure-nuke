package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/preview/resources/mgmt/2021-06-01-preview/policy" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
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

	client      policy.DefinitionsClient
	Name        *string
	DisplayName string
	PolicyType  string `property:"name=Type"`
}

func (r *PolicyDefinition) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.Name)
	return err
}

func (r *PolicyDefinition) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PolicyDefinition) String() string {
	return *r.Name
}

type PolicyDefinitionLister struct {
}

func (l PolicyDefinitionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", PolicyDefinitionResource).WithField("s", opts.SubscriptionID)

	client := policy.NewDefinitionsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list policy definitions")

	list, err := client.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing policy definitions")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			// Filtering out BuiltIn Policy Definitions, because otherwise it needlessly adds 3000+
			// resources that have to get filtered out later. This instead does it optimistically here.
			// Ideally we'd be able to use filter above, but it does not work. Thanks, Azure. :facepalm:
			if g.PolicyType == "BuiltIn" || g.PolicyType == "Static" {
				continue
			}

			resources = append(resources, &PolicyDefinition{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client:      client,
				Name:        g.Name,
				DisplayName: ptr.ToString(g.DisplayName),
				PolicyType:  string(g.PolicyType),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.WithField("total", len(resources)).Trace("done")

	return resources, nil
}
