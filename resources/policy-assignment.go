package resources

import (
	"context"
	"fmt"
	"strings"

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

const PolicyAssignmentResource = "PolicyAssignment"

func init() {
	registry.Register(&registry.Registration{
		Name:     PolicyAssignmentResource,
		Scope:    azure.SubscriptionScope,
		Resource: &PolicyAssignment{},
		Lister:   &PolicyAssignmentLister{},
	})
}

type PolicyAssignment struct {
	*BaseResource `property:",inline"`

	client          *armpolicy.AssignmentsClient
	Name            string
	Scope           string
	EnforcementMode string
}

func (r *PolicyAssignment) Filter() error {
	if strings.HasPrefix(r.Name, "sys.") {
		return fmt.Errorf("cannot remove built-in policy")
	}
	return nil
}

func (r *PolicyAssignment) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, r.Scope, r.Name, nil)
	return err
}

func (r *PolicyAssignment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PolicyAssignment) String() string {
	return r.Name
}

type PolicyAssignmentLister struct{}

func (l PolicyAssignmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", PolicyAssignmentResource).WithField("s", opts.SubscriptionID)

	client, err := armpolicy.NewAssignmentsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds,
		&arm.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				APIVersion: "2024-04-01",
			},
		})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list policy assignments")

	pager := client.NewListPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		log.Trace("listing policy assignments")

		for _, g := range page.Value {
			enforcementMode := ""
			if g.Properties != nil && g.Properties.EnforcementMode != nil {
				enforcementMode = string(*g.Properties.EnforcementMode)
			}

			scope := ""
			if g.Properties != nil && g.Properties.Scope != nil {
				scope = *g.Properties.Scope
			}

			resources = append(resources, &PolicyAssignment{
				BaseResource: &BaseResource{
					Region: ptr.String("global"),
				},
				client:          client,
				Name:            ptr.ToString(g.Name),
				Scope:           scope,
				EnforcementMode: enforcementMode,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
