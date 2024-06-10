package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const SubscriptionRoleAssignmentResource = "SubscriptionRoleAssignment"

func init() {
	registry.Register(&registry.Registration{
		Name:     SubscriptionRoleAssignmentResource,
		Scope:    nuke.Subscription,
		Resource: &SubscriptionRoleAssignment{},
		Lister: &SubscriptionRoleAssignmentLister{
			roleNameCache:      make(map[*string]*string),
			principalNameCache: make(map[*string]*string),
		},
	})
}

type SubscriptionRoleAssignment struct {
	client *armauthorization.RoleAssignmentsClient

	ID               *string `property:"-"`
	Name             *string
	Type             *string `property:"-"`
	RoleName         *string
	RoleDefinitionID *string
	PrincipalID      *string
	PrincipalName    *string
	PrincipalType    *string
	scope            *string
	subscriptionID   *string
}

func (r *SubscriptionRoleAssignment) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.scope, *r.Name, nil)
	return err
}

func (r *SubscriptionRoleAssignment) Filter() error {
	if *r.scope != fmt.Sprintf("/subscriptions/%s", *r.subscriptionID) {
		return fmt.Errorf("role assigned at a different level than the subscription")
	}

	return nil
}

func (r *SubscriptionRoleAssignment) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SubscriptionRoleAssignment) String() string {
	return fmt.Sprintf("%s -> %s", *r.PrincipalName, *r.RoleName)
}

type SubscriptionRoleAssignmentLister struct {
	roleNameCache      map[*string]*string
	principalNameCache map[*string]*string
}

func (l *SubscriptionRoleAssignmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) { //nolint:gocyclo,funlen
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	log := logrus.WithField("r", SubscriptionRoleAssignmentResource).WithField("s", opts.SubscriptionID)

	client, err := armauthorization.NewRoleAssignmentsClient(
		opts.SubscriptionID, opts.Authorizers.IdentityCreds, &arm.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				APIVersion: "2022-04-01",
			},
		})
	if err != nil {
		return resources, nil
	}

	defClient, err := armauthorization.NewRoleDefinitionsClient(opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return resources, nil
	}

	userClient := msgraph.NewUsersClient()
	userClient.BaseClient.Authorizer = opts.Authorizers.Graph
	userClient.BaseClient.DisableRetries = true

	groupClient := msgraph.NewGroupsClient()
	groupClient.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	groupClient.BaseClient.DisableRetries = true

	spClient := msgraph.NewServicePrincipalsClient()
	spClient.BaseClient.Authorizer = opts.Authorizers.MicrosoftGraph
	spClient.BaseClient.DisableRetries = true

	log.Debug("listing subscription role assignments")
	pager := client.NewListPager(&armauthorization.RoleAssignmentsClientListOptions{Filter: ptr.String("atScope()")})

	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return resources, nil
		}
		for _, t := range nextResult.Value {
			rel, defErr := defClient.GetByID(ctx, *t.Properties.RoleDefinitionID, nil)
			if defErr != nil {
				return nil, defErr
			}

			if _, ok := l.roleNameCache[t.Properties.RoleDefinitionID]; !ok {
				l.roleNameCache[t.Properties.RoleDefinitionID] = rel.RoleDefinition.Properties.RoleName
			}

			var principalName = ptr.String("unknown")
			var principalType = ptr.String("unknown")

			if v, ok := l.principalNameCache[t.Properties.PrincipalID]; !ok {
				u, _, err := userClient.Get(ctx, *t.Properties.PrincipalID, odata.Query{})
				if err != nil && !strings.Contains(err.Error(), "ResourceNotFound") {
					return nil, err
				}
				if u != nil {
					log.Debug("found user")
					principalName = u.DisplayName
					principalType = ptr.String("user")
				}

				g, _, err := groupClient.Get(ctx, *t.Properties.PrincipalID, odata.Query{})
				if err != nil && !strings.Contains(err.Error(), "ResourceNotFound") {
					return nil, err
				}
				if g != nil {
					log.Debug("found group")
					principalName = g.DisplayName
					principalType = ptr.String("group")
				}

				sp, _, err := spClient.Get(ctx, *t.Properties.PrincipalID, odata.Query{})
				if err != nil && !strings.Contains(err.Error(), "ResourceNotFound") {
					return nil, err
				}
				if sp != nil {
					log.Debug("found service principal")
					principalName = sp.DisplayName
					principalType = ptr.String("service_principal")
				}

				l.principalNameCache[t.Properties.PrincipalID] = principalName
			} else {
				principalName = v
			}

			roleDefinitionIDParts := strings.Split(*t.Properties.RoleDefinitionID, "/")
			roleDefinitionID := roleDefinitionIDParts[len(roleDefinitionIDParts)-1]

			resources = append(resources, &SubscriptionRoleAssignment{
				client:           client,
				scope:            t.Properties.Scope,
				subscriptionID:   ptr.String(opts.SubscriptionID),
				ID:               t.ID,
				Name:             t.Name,
				Type:             t.Type,
				RoleName:         l.roleNameCache[t.Properties.RoleDefinitionID],
				RoleDefinitionID: ptr.String(roleDefinitionID),
				PrincipalID:      t.Properties.PrincipalID,
				PrincipalName:    principalName,
				PrincipalType:    principalType,
			})
		}
	}

	return resources, nil
}
