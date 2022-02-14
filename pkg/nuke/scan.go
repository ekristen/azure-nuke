package nuke

import (
	"context"
	"fmt"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/rebuy-de/aws-nuke/pkg/awsutil"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/queue"
	"github.com/ekristen/azure-nuke/pkg/resource"
	"github.com/ekristen/azure-nuke/pkg/utils"

	_ "github.com/ekristen/azure-nuke/resources"
)

const ScannerParallelQueries = 16

func Scan(authorizers azure.Authorizers, tenantId, subscriptionId, resourceGroup string, resourceTypes []string) <-chan *queue.Item {
	s := &scanner{
		items:     make(chan *queue.Item, 100),
		semaphore: semaphore.NewWeighted(ScannerParallelQueries),
	}
	go s.run(authorizers, tenantId, subscriptionId, resourceGroup, resourceTypes)

	return s.items
}

type scanner struct {
	items     chan *queue.Item
	semaphore *semaphore.Weighted
}

func (s *scanner) run(authorizers azure.Authorizers, tenantId, subscriptionId, resourceGroup string, resourceTypes []string) {
	ctx := context.Background()

	for _, resourceType := range resourceTypes {
		s.semaphore.Acquire(ctx, 1)
		go s.list(authorizers, tenantId, subscriptionId, resourceGroup, resourceType)
	}

	// Wait for all routines to finish.
	s.semaphore.Acquire(ctx, ScannerParallelQueries)

	close(s.items)
}

func (s *scanner) list(authorizers azure.Authorizers, tenantId, subscriptionId, resourceGroup, resourceType string) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n\n%s", r.(error), string(debug.Stack()))
			dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
			log.Errorf("Listing %s failed:\n%s", resourceType, dump)
		}
	}()
	defer s.semaphore.Release(1)

	lister := resource.GetListerV2(resourceType)
	var rs []resource.Resource
	rs, err := lister(resource.ListerOpts{
		Authorizers:    authorizers,
		TenantId:       tenantId,
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
	})
	if err != nil {
		_, ok := err.(awsutil.ErrSkipRequest)
		if ok {
			log.Debugf("skipping request: %v", err)
			return
		}

		_, ok = err.(awsutil.ErrUnknownEndpoint)
		if ok {
			log.Warnf("skipping request: %v", err)
			return
		}

		dump := utils.Indent(fmt.Sprintf("%v", err), "    ")
		log.Errorf("Listing %s failed:\n%s", resourceType, dump)
		return
	}

	for _, r := range rs {
		s.items <- &queue.Item{
			Authorizers:    authorizers,
			TenantId:       tenantId,
			SubscriptionId: subscriptionId,
			ResourceGroup:  resourceGroup,
			Resource:       r,
			State:          queue.ItemStateNew,
			Type:           resourceType,
		}
	}
}
