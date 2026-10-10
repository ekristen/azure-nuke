package azure

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenant_RemoveSubscriptions(t *testing.T) {
	const test, prod = "test", "prod"

	tenant := &Tenant{
		SubscriptionIds: []string{test, prod},
		ResourceGroups: map[string][]string{
			test: {"rg-test"},
			prod: {"rg-prod"},
		},
		Regions: map[string][]string{},
	}

	removed := tenant.RemoveSubscriptions(func(id string) bool { return slices.Contains([]string{prod}, id) })

	assert.Equal(t, []string{prod}, removed)
	assert.Equal(t, []string{test}, tenant.SubscriptionIds)
	assert.NotContains(t, tenant.ResourceGroups, prod)
	assert.Contains(t, tenant.ResourceGroups, test)
}
