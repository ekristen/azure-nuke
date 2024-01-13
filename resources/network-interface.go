package resources

import (
	"context"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-05-01/network"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const NetworkInterfaceResource = "NetworkInterface"

func init() {
	resource.Register(resource.Registration{
		Name:   NetworkInterfaceResource,
		Scope:  nuke.ResourceGroup,
		Lister: NetworkInterfaceLister{},
	})
}

type NetworkInterfaceLister struct {
}

func (l NetworkInterfaceLister) List(o interface{}) ([]resource.Resource, error) {
	opts := o.(nuke.ListerOpts)

	log := logrus.WithField("r", NetworkInterfaceResource).WithField("s", opts.SubscriptionId)

	client := network.NewInterfacesClient(opts.SubscriptionId)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list network interfaces")

	ctx := context.TODO()

	list, err := client.List(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &NetworkInterface{
				client: client,
				name:   g.Name,
				rg:     &opts.ResourceGroup,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done listing network interfaces")

	return resources, nil
}

type NetworkInterface struct {
	client network.InterfacesClient
	name   *string
	rg     *string
}

func (r *NetworkInterface) Remove() error {
	_, err := r.client.Delete(context.TODO(), *r.rg, *r.name)
	return err
}

func (r *NetworkInterface) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", *r.name)
	properties.Set("ResourceGroup", *r.rg)

	return properties
}

func (r *NetworkInterface) String() string {
	return *r.name
}
