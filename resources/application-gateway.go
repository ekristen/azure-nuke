package resources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v12"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const ApplicationGatewayResource = "ApplicationGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     ApplicationGatewayResource,
		Scope:    azure.ResourceGroupScope,
		Resource: &ApplicationGateway{},
		Lister:   &ApplicationGatewayLister{},
	})
}

type ApplicationGatewayLister struct{}

func (l ApplicationGatewayLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	var resources []resource.Resource
	opts := o.(*azure.ListerOpts)

	log := logrus.WithField("r", ApplicationGatewayResource).WithField("s", opts.SubscriptionID)

	client, err := armnetwork.NewApplicationGatewaysClient(opts.SubscriptionID, opts.Authorizers.IdentityCreds, nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	log.Trace("attempting to list application gateways")

	pager := client.NewListPager(opts.ResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, entry := range page.Value {
			resources = append(resources, &ApplicationGateway{
				BaseResource: &BaseResource{
					Region:         ptr.String("global"),
					SubscriptionID: ptr.String(opts.SubscriptionID),
					ResourceGroup:  ptr.String(opts.ResourceGroup),
				},
				client: client,
				ID:     entry.ID,
				Name:   entry.Name,
			})
		}
	}

	log.Trace("done")

	return resources, nil
}

type ApplicationGateway struct {
	*BaseResource `property:",inline"`

	client   *armnetwork.ApplicationGatewaysClient
	removeOp *runtime.Poller[armnetwork.ApplicationGatewaysClientDeleteResponse]
	ID       *string
	Name     *string
}

func (r *ApplicationGateway) Filter() error {
	return nil
}

func (r *ApplicationGateway) Remove(ctx context.Context) error {
	r.removeOp = nil

	poller, err := r.client.BeginDelete(ctx, *r.ResourceGroup, *r.Name, nil)
	if err != nil {
		return azure.IgnoreNotFound(err)
	}

	r.removeOp = poller
	return nil
}

func (r *ApplicationGateway) HandleWait(ctx context.Context) error {
	return azure.WaitForPoller(ctx, r.removeOp)
}

func (r *ApplicationGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ApplicationGateway) String() string {
	return *r.Name
}
