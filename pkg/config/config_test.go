package config

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	libconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"
)

func TestLoadExampleConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	entry := logrus.WithField("test", true)

	config, err := New(libconfig.Options{
		Path: "testdata/example.yaml",
		Log:  entry,
	})
	if err != nil {
		t.Fatal(err)
	}

	expect := Config{
		Config: libconfig.Config{
			Blocklist: []string{"382ee010-63bb-428b-b0f4-3c9081e32ddb"},
			Regions:   []string{"global", "eastus"},
			Accounts: map[string]*libconfig.Account{
				"efda01a1-e2e4-4024-89f0-eb29793c605b": {
					Presets: []string{"common"},
				},
			},
			ResourceTypes: libconfig.ResourceTypes{
				Includes: types.Collection{"ResourceGroup", "ServicePrincipal"},
				Excludes: types.Collection{"AzureADGroup"},
			},
			Presets: map[string]libconfig.Preset{
				"common": {
					Filters: filter.Filters{
						"ResourceGroup": {
							filter.Filter{
								Group: "default",
								Type:  filter.Exact,
								Value: "Default",
							},
						},
						"ServicePrincipal": {
							filter.Filter{
								Group: "default",
								Type:  filter.Exact,
								Value: "some-management-account",
							},
						},
					},
				},
			},
			Settings:     &settings.Settings{},
			Deprecations: make(map[string]string),
			Log:          entry,
		},
	}

	assert.Equal(t, expect, *config)
}
