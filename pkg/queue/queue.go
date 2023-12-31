package queue

import (
	"github.com/ekristen/azure-nuke/pkg/nuke"
	"github.com/ekristen/cloud-nuke-sdk/pkg/queue"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/cloud-nuke-sdk/pkg/log"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"

	_ "github.com/ekristen/azure-nuke/resources"
)

// An Item describes an actual AWS resource entity with the current state and
// some metadata.
type Item struct {
	queue.Item

	Authorizers    azure.Authorizers
	TenantId       string
	SubscriptionId string
	ResourceGroup  string

	Region *string
	Type   string
}

func (i *Item) Print() {
	switch i.State {
	case queue.ItemStateNew:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonWaitPending, "would remove")
	case queue.ItemStatePending:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonWaitPending, "triggered remove")
	case queue.ItemStateWaiting:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonWaitPending, "waiting")
	case queue.ItemStateFailed:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonError, "failed")
	case queue.ItemStateFiltered:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonSkip, i.Reason)
	case queue.ItemStateFinished:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonSuccess, "removed")
	}
}

// List gets all resource items of the same resource type like the Item.
func (i *Item) List() ([]resource.Resource, error) {
	listers := resource.GetListers()
	/*
		sess, err := i.Region.Session(i.Type)
		if err != nil {
			return nil, err
		}
		return listers[i.Type](sess)
	*/
	return listers[i.Type](nuke.ListerOpts{
		Authorizers:    i.Authorizers,
		TenantId:       i.TenantId,
		SubscriptionId: i.SubscriptionId,
		ResourceGroup:  i.ResourceGroup,
	})
}

type Queue []*Item

func (q Queue) CountTotal() int {
	return len(q)
}

func (q Queue) Count(states ...queue.ItemState) int {
	count := 0
	for _, item := range q {
		for _, state := range states {
			if item.State == state {
				count = count + 1
				break
			}
		}
	}
	return count
}
