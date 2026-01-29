package resources

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-07-01/containerservice" //nolint:staticcheck

	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ManagedClusterResource = "ManagedCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     ManagedClusterResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &ManagedCluster{},
		Lister:   &ManagedClusterLister{},
	})
}

type ManagedCluster struct {
	*BaseResource `property:",inline"`

	client containerservice.ManagedClustersClient
	ID     *string
	Name   *string
	Tags   map[string]*string
}

func (r *ManagedCluster) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *ManagedCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ManagedCluster) String() string {
	return *r.Name
}

// -----------------------------------------

type ManagedClusterLister struct {
}

func (l ManagedClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ManagedClusterResource).WithField("s", opts.SubscriptionID)

	client := containerservice.NewManagedClustersClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list managed clusters")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing resources")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &ManagedCluster{
				BaseResource: &BaseResource{
					Region:         g.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client: client,
				ID:     g.ID,
				Name:   g.Name,
				Tags:   g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
