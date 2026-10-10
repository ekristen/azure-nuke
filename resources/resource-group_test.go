package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/settings"
)

func TestResourceGroup_Filter(t *testing.T) {
	cases := []struct {
		name      string
		managedBy *string
		setting   settings.Setting
		filtered  bool
	}{
		{name: "unmanaged", setting: settings.Setting{}},
		{name: "managed", managedBy: ptr.String("/subscriptions/x/resourceGroups/y"), setting: settings.Setting{}, filtered: true},
		{
			name:      "managed with DeleteManaged",
			managedBy: ptr.String("/subscriptions/x/resourceGroups/y"),
			setting:   settings.Setting{"DeleteManaged": true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &ResourceGroup{Name: ptr.String("rg"), ManagedBy: tc.managedBy}
			r.Settings(&tc.setting)

			if tc.filtered {
				assert.Error(t, r.Filter())
			} else {
				assert.NoError(t, r.Filter())
			}
		})
	}
}
