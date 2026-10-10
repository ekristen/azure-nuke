package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v8"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const VirtualMachineResource = "VirtualMachine"

func init() {
	registry.Register(&registry.Registration{
		Name:     VirtualMachineResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &VirtualMachine{},
		Lister:   &VirtualMachineLister{},
	})
}

type VirtualMachine struct {
	*BaseResource `property:",inline"`

	client       *armcompute.VirtualMachinesClient
	Name         *string
	Tags         map[string]*string
	CreationDate *time.Time
}

func (r *VirtualMachine) Remove(ctx context.Context) error {
	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, &armcompute.VirtualMachinesClientBeginDeleteOptions{
		ForceDeletion: ptr.Bool(true),
	})
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func (r *VirtualMachine) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VirtualMachine) String() string {
	return *r.Name
}

// -----------------------------------------

type VirtualMachineLister struct{}

func (l VirtualMachineLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", VirtualMachineResource).WithField("s", opts.SubscriptionID)

	client, err := armcompute.NewVirtualMachinesClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list virtual machines")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entity := range page.Value {
			instanceView, err := client.InstanceView(ctx, opts.ResourceGroup, *entity.Name, nil)
			if err != nil {
				return nil, err
			}

			var creationDate *time.Time
			for _, status := range instanceView.Statuses {
				if status.Code != nil && *status.Code == "ProvisioningState/succeeded" {
					creationDate = status.Time
					break
				}
			}

			resources = append(resources, &VirtualMachine{
				BaseResource: &BaseResource{
					Region:         entity.Location,
					ResourceGroup:  &opts.ResourceGroup,
					SubscriptionID: &opts.SubscriptionID,
				},
				client:       client,
				Name:         entity.Name,
				Tags:         entity.Tags,
				CreationDate: creationDate,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}
