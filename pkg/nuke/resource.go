package nuke

import (
	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/cloud-nuke-sdk/pkg/resource"
)

const (
	Tenant        resource.Scope = "tenant"
	Subscription  resource.Scope = "subscription"
	ResourceGroup resource.Scope = "resource-group"
)

type ListerOpts struct {
	Authorizers    azure.Authorizers
	TenantId       string
	SubscriptionId string
	ResourceGroup  string
}

func (o ListerOpts) ID() string {
	return ""
}

type Lister struct {
	opts ListerOpts
}

func (l Lister) SetOptions(opts interface{}) {
	l.opts = opts.(ListerOpts)
}
