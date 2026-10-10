package resources

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const RoleAssignmentResource = "RoleAssignment"

// ARM returns role assignment scopes with /resourceGroups/ or /resourcegroups/
var roleAssignmentResourceGroupRegex = regexp.MustCompile(`(?i)/resourceGroups/([^/]+)`)

func init() {
	registry.Register(&registry.Registration{
		Name:     RoleAssignmentResource,
		Scope:    azure.SubscriptionScope,
		Resource: &RoleAssignment{},
		Lister:   &RoleAssignmentLister{},
	})
}

// RoleAssignmentLister lists the role assignments on resource groups and resources of the subscription.
// Subscription-scope assignments are SubscriptionRoleAssignment's.
type RoleAssignmentLister struct{}

func (l RoleAssignmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", RoleAssignmentResource).WithField("s", opts.SubscriptionID)

	client, err := armauthorization.NewRoleAssignmentsClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	defClient, err := armauthorization.NewRoleDefinitionsClient(opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	roleNames := make(map[string]*string)
	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list role assignments")

	pager := client.NewListForSubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			if entity.Properties == nil {
				continue
			}

			definitionID := ptr.ToString(entity.Properties.RoleDefinitionID)
			if _, ok := roleNames[definitionID]; !ok {
				// a deleted custom role keeps its assignments: name them by the definition id
				roleNames[definitionID] = ptr.String(definitionID)
				definition, err := defClient.GetByID(ctx, definitionID, nil)
				if err != nil && !azure.IsNotFound(err) {
					return nil, err
				}
				if err == nil && definition.Properties != nil && definition.Properties.RoleName != nil {
					roleNames[definitionID] = definition.Properties.RoleName
				}
			}

			var principalType *string
			if entity.Properties.PrincipalType != nil {
				principalType = ptr.String(string(*entity.Properties.PrincipalType))
			}

			resources = append(resources, &RoleAssignment{
				BaseResource: &BaseResource{
					Region:         ptr.String("global"),
					ResourceGroup:  scopeResourceGroup(ptr.ToString(entity.Properties.Scope)),
					SubscriptionID: ptr.String(opts.SubscriptionID),
				},
				client:        client,
				ID:            entity.ID,
				Name:          entity.Name,
				Scope:         entity.Properties.Scope,
				RoleName:      roleNames[definitionID],
				PrincipalID:   entity.Properties.PrincipalID,
				PrincipalType: principalType,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}

// ---------------------------------------------

type RoleAssignment struct {
	*BaseResource `property:",inline"`

	client        *armauthorization.RoleAssignmentsClient
	ID            *string `description:"The ID of the role assignment."`
	Name          *string `description:"The name (GUID) of the role assignment."`
	Scope         *string `description:"The scope the role is assigned at."`
	RoleName      *string `description:"The name of the assigned role."`
	PrincipalID   *string `description:"The object ID of the principal." property:"name=PrincipalId"`
	PrincipalType *string `description:"The type of the principal: User, Group, ServicePrincipal, ..."`
}

// Filter keeps every assignment outside the subscription's resource groups: on the subscription itself (owners, the
// identity running the nuke) or inherited from a management group or the tenant root.
func (r *RoleAssignment) Filter() error {
	if !strings.HasPrefix(strings.ToLower(ptr.ToString(r.Scope)),
		strings.ToLower("/subscriptions/"+ptr.ToString(r.SubscriptionID)+"/resourceGroups/")) {
		return errors.New("not assigned within a resource group of the subscription")
	}
	return nil
}

func (r *RoleAssignment) Remove(ctx context.Context) error {
	_, err := r.client.DeleteByID(ctx, *r.ID, nil)
	return azure.IgnoreNotFound(err)
}

func (r *RoleAssignment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func scopeResourceGroup(scope string) *string {
	if matches := roleAssignmentResourceGroupRegex.FindStringSubmatch(scope); len(matches) == 2 {
		return &matches[1]
	}
	return nil
}

func (r *RoleAssignment) UniqueKey() string {
	return ptr.ToString(r.ID)
}

func (r *RoleAssignment) String() string {
	return ptr.ToString(r.PrincipalID) + " -> " + ptr.ToString(r.RoleName) + " @ " + ptr.ToString(r.Scope)
}
