package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/libnuke/pkg/registry"
)

const (
	Tenant        registry.Scope = "tenant"
	Subscription  registry.Scope = "subscription"
	ResourceGroup registry.Scope = "resource-group"
)

type ListerOpts struct {
	Authorizers    *azure.Authorizers
	TenantId       string
	SubscriptionId string
	ResourceGroup  string
	Locations      []string
}
