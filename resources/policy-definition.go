package resources

import (
	"context"
	"fmt"
	"github.com/aws/smithy-go/ptr"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/resources/mgmt/2021-06-01-preview/policy"
)

const PolicyDefinitionResource = "PolicyDefinition"

func init() {
	resource.Register(resource.Registration{
		Name:   PolicyDefinitionResource,
		Scope:  nuke.Subscription,
		Lister: PolicyDefinitionLister{},
	})
}

type PolicyDefinition struct {
	client      policy.DefinitionsClient
	name        string
	displayName string
	policyType  string
}

func (r *PolicyDefinition) Filter() error {
	if r.policyType == "BuiltIn" || r.policyType == "Static" {
		return fmt.Errorf("cannot delete policies with type %s", r.policyType)
	}
	return nil
}

func (r *PolicyDefinition) Remove() error {
	_, err := r.client.Delete(context.TODO(), r.name)
	return err
}

func (r *PolicyDefinition) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", r.name)
	properties.Set("DisplayName", r.displayName)
	properties.Set("Type", r.policyType)

	return properties
}

func (r *PolicyDefinition) String() string {
	return r.name
}

type PolicyDefinitionLister struct {
}

func (l PolicyDefinitionLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", PolicyDefinitionResource).WithField("s", opts.SubscriptionId)

	client := policy.NewDefinitionsClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list policy definitions")

	ctx := context.TODO()
	list, err := client.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	log.Trace("listing policy definitions")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &PolicyDefinition{
				client:      client,
				name:        *g.Name,
				displayName: ptr.ToString(g.DisplayName),
				policyType:  string(g.PolicyType),
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.WithField("total", len(resources)).Trace("done")

	return resources, nil
}
