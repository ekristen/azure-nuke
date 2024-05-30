package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/hashicorp/go-azure-sdk/sdk/odata"
	"github.com/manicminer/hamilton/msgraph"

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
	client authorization.RoleAssignmentsClient

	ID             *string `property:"-"`
	Name           *string
	Type           *string `property:"-"`
	RoleName       *string
	PrincipalID    *string
	PrincipalName  *string
	PrincipalType  *string
	scope          *string
	subscriptionID *string
}

func (r *SubscriptionRoleAssignment) Remove(ctx context.Context) error {
	return nil
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

	client := authorization.NewRoleAssignmentsClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management

	defClient := authorization.NewRoleDefinitionsClient(opts.SubscriptionID)
	defClient.Authorizer = opts.Authorizers.Management

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
	lister, err := client.List(ctx, "atScope()")
	if err != nil {
		return nil, err
	}

	for lister.NotDone() {
		for _, t := range lister.Values() {
			rel, defErr := defClient.GetByID(ctx, *t.Properties.RoleDefinitionID)
			if defErr != nil {
				return nil, defErr
			}

			if _, ok := l.roleNameCache[t.Properties.RoleDefinitionID]; !ok {
				l.roleNameCache[t.Properties.RoleDefinitionID] = rel.RoleName
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

			resources = append(resources, &SubscriptionRoleAssignment{
				client:         client,
				scope:          t.Properties.Scope,
				subscriptionID: ptr.String(opts.SubscriptionID),
				ID:             t.ID,
				Name:           t.Name,
				Type:           t.Type,
				RoleName:       l.roleNameCache[t.Properties.RoleDefinitionID],
				PrincipalID:    t.Properties.PrincipalID,
				PrincipalName:  principalName,
				PrincipalType:  principalType,
			})
		}

		if listErr := lister.NextWithContext(ctx); listErr != nil {
			return nil, listErr
		}
	}

	return resources, nil
}
