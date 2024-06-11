package resources

import (
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/queue"
)

// BaseResource is a base struct that all Azure resources should embed to provide common fields and methods.
type BaseResource struct {
	Region         *string `description:"The region that the resource group belongs to."`
	SubscriptionID *string `description:"The subscription ID that the resource group belongs to."`
	ResourceGroup  *string `description:"The resource group that the resource belongs to."`
}

// GetRegion returns the region that the resource belongs to.
func (r *BaseResource) GetRegion() string {
	return ptr.ToString(r.Region)
}

// GetSubscriptionID returns the subscription ID that the resource belongs to.
func (r *BaseResource) GetSubscriptionID() string {
	return ptr.ToString(r.SubscriptionID)
}

// GetResourceGroup returns the resource group that the resource belongs to.
func (r *BaseResource) GetResourceGroup() string {
	return ptr.ToString(r.ResourceGroup)
}

// BeforeEnqueue is a special hook that is called from github.com/ekristen/libnuke that allows the resource to
// modify the queue item before it is put on the queue, in this case it allows us to modify the owner field to
// set it as the region so the behavior of this tool is consistent with the other tools based on libnuke and regions
func (r *BaseResource) BeforeEnqueue(item interface{}) {
	i := item.(*queue.Item)
	i.Owner = ptr.ToString(r.Region)
}
