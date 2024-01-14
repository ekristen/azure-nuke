package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/libnuke/pkg/resource"
)

const (
	Tenant        resource.Scope = "tenant"
	Subscription  resource.Scope = "subscription"
	ResourceGroup resource.Scope = "resource-group"
)

type ListerOpts struct {
	Authorizers    *azure.Authorizers
	TenantId       string
	SubscriptionId string
	ResourceGroup  string
}
