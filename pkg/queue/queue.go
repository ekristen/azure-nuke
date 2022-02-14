package queue

import (
	"fmt"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/log"
	"github.com/ekristen/azure-nuke/pkg/resource"

	_ "github.com/ekristen/azure-nuke/resources"
)

type ItemState int

// States of Items based on the latest request to AWS.
const (
	ItemStateNew ItemState = iota
	ItemStatePending
	ItemStateWaiting
	ItemStateFailed
	ItemStateFiltered
	ItemStateFinished
)

// An Item describes an actual AWS resource entity with the current state and
// some metadata.
type Item struct {
	Resource resource.Resource

	State  ItemState
	Reason string

	Authorizers    azure.Authorizers
	TenantId       string
	SubscriptionId string
	ResourceGroup  string

	Region *string
	Type   string
}

func (i *Item) Print() {
	switch i.State {
	case ItemStateNew:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonWaitPending, "would remove")
	case ItemStatePending:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonWaitPending, "triggered remove")
	case ItemStateWaiting:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonWaitPending, "waiting")
	case ItemStateFailed:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonError, "failed")
	case ItemStateFiltered:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonSkip, i.Reason)
	case ItemStateFinished:
		log.Log(i.SubscriptionId, i.Type, i.Resource, log.ReasonSuccess, "removed")
	}
}

// List gets all resource items of the same resource type like the Item.
func (i *Item) List() ([]resource.Resource, error) {
	listers := resource.GetListersV2()
	/*
		sess, err := i.Region.Session(i.Type)
		if err != nil {
			return nil, err
		}
		return listers[i.Type](sess)
	*/
	return listers[i.Type](resource.ListerOpts{
		Authorizers:    i.Authorizers,
		TenantId:       i.TenantId,
		SubscriptionId: i.SubscriptionId,
		ResourceGroup:  i.ResourceGroup,
	})
}

func (i *Item) GetProperty(key string) (string, error) {
	if key == "" {
		stringer, ok := i.Resource.(resource.LegacyStringer)
		if !ok {
			return "", fmt.Errorf("%T does not support legacy IDs", i.Resource)
		}
		return stringer.String(), nil
	}

	getter, ok := i.Resource.(resource.ResourcePropertyGetter)
	if !ok {
		return "", fmt.Errorf("%T does not support custom properties", i.Resource)
	}

	return getter.Properties().Get(key), nil
}

func (i *Item) Equals(o resource.Resource) bool {
	iType := fmt.Sprintf("%T", i.Resource)
	oType := fmt.Sprintf("%T", o)
	if iType != oType {
		return false
	}

	iStringer, iOK := i.Resource.(resource.LegacyStringer)
	oStringer, oOK := o.(resource.LegacyStringer)
	if iOK != oOK {
		return false
	}
	if iOK && oOK {
		return iStringer.String() == oStringer.String()
	}

	iGetter, iOK := i.Resource.(resource.ResourcePropertyGetter)
	oGetter, oOK := o.(resource.ResourcePropertyGetter)
	if iOK != oOK {
		return false
	}
	if iOK && oOK {
		return iGetter.Properties().Equals(oGetter.Properties())
	}

	return false
}

type Queue []*Item

func (q Queue) CountTotal() int {
	return len(q)
}

func (q Queue) Count(states ...ItemState) int {
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
