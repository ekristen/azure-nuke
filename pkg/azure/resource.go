package azure

import (
	"regexp"

	"github.com/ekristen/libnuke/pkg/registry"
)

const (
	TenantScope        registry.Scope = "tenant"
	SubscriptionScope  registry.Scope = "subscription"
	ResourceGroupScope registry.Scope = "resource-group"
)

var ResourceGroupRegex = regexp.MustCompile(`/resourceGroups/([^/]+)`)

type ListerOpts struct {
	Authorizers    *Authorizers
	TenantID       string
	SubscriptionID string
	ResourceGroup  string
	ResourceGroups []string
	Region         string
	Regions        []string
}

func GetResourceGroupFromID(id string) *string {
	matches := ResourceGroupRegex.FindStringSubmatch(id)
	if len(matches) == 2 {
		return &matches[1]
	}

	return nil
}
