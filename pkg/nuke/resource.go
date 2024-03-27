package nuke

import (
	"regexp"

	"github.com/ekristen/libnuke/pkg/registry"

	"github.com/ekristen/azure-nuke/pkg/azure"
)

const (
	Tenant        registry.Scope = "tenant"
	Subscription  registry.Scope = "subscription"
	ResourceGroup registry.Scope = "resource-group"
)

var (
	ResourceGroupRegex = regexp.MustCompile(`/resourceGroups/([^/]+)`)
)

type ListerOpts struct {
	Authorizers    *azure.Authorizers
	TenantID       string
	SubscriptionID string
	ResourceGroup  string
	Regions        []string
}

func GetResourceGroupFromID(id string) *string {
	matches := ResourceGroupRegex.FindStringSubmatch(id)
	if len(matches) == 2 {
		return &matches[1]
	}

	return nil
}
