package config

import (
	"fmt"
	"github.com/ekristen/cloud-nuke-sdk/pkg/config"
	"github.com/ekristen/cloud-nuke-sdk/pkg/filter"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Tenant struct {
	Filters       filter.Filters       `yaml:"filters"`
	ResourceTypes config.ResourceTypes `yaml:"resource-types"`
	Presets       []string             `yaml:"presets"`
}

type Nuke struct {
	Tenants         map[string]Tenant                   `yaml:"tenants"`
	TenantBlocklist []string                            `yaml:"tenant-blocklist"`
	ResourceTypes   config.ResourceTypes                `yaml:"resource-types"`
	FeatureFlags    config.FeatureFlags                 `yaml:"feature-flags"`
	Presets         map[string]config.PresetDefinitions `yaml:"presets"`
}

type FeatureFlags struct {
	DisableDeletionProtection  DisableDeletionProtection `yaml:"disable-deletion-protection"`
	ForceDeleteLightsailAddOns bool                      `yaml:"force-delete-lightsail-addons"`
}

type DisableDeletionProtection struct {
	RDSInstance         bool `yaml:"RDSInstance"`
	EC2Instance         bool `yaml:"EC2Instance"`
	CloudformationStack bool `yaml:"CloudformationStack"`
}

type PresetDefinitions struct {
	Filters filter.Filters `yaml:"filters"`
}

func Load(path string) (*Nuke, error) {
	var err error

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := new(Nuke)
	err = yaml.UnmarshalStrict(raw, cfg)
	if err != nil {
		return nil, err
	}

	if err := cfg.ResolveDeprecations(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Nuke) GetResourceTypes() config.ResourceTypes {
	return c.ResourceTypes
}

func (c *Nuke) GetFeatureFlags() config.FeatureFlags {
	return c.FeatureFlags
}

func (c *Nuke) GetPresets() map[string]config.PresetDefinitions {
	return c.Presets
}

func (c *Nuke) ResolveBlocklist() []string {
	if c.TenantBlocklist != nil {
		return c.TenantBlocklist
	}

	return c.TenantBlocklist
}

func (c *Nuke) HasBlocklist() bool {
	var blocklist = c.ResolveBlocklist()
	return len(blocklist) > 0
}

func (c *Nuke) InBlocklist(searchID string) bool {
	for _, blocklistID := range c.ResolveBlocklist() {
		if blocklistID == searchID {
			return true
		}
	}

	return false
}

func (c *Nuke) Validate(tenantId string) error {
	if !c.HasBlocklist() {
		return fmt.Errorf("the config file contains an empty blocklist. " +
			"For safety reasons you need to specify at least one tenant ID. " +
			"This should be your production account")
	}

	if c.InBlocklist(tenantId) {
		return fmt.Errorf("you are trying to nuke the tenant with the ID %s, "+
			"but it is blocklisted. Aborting", tenantId)
	}

	/*
		if len(aliases) == 0 {
			return fmt.Errorf("the specified account doesn't have an alias. " +
				"For safety reasons you need to specify an account alias. " +
				"Your production account should contain the term 'prod'")
		}
	*/

	/*
		for _, alias := range aliases {
			if strings.Contains(strings.ToLower(alias), "prod") {
				return fmt.Errorf("you are trying to nuke an tenant with the alias, '%s' "+
					"but it has the substring 'prod' in it. aborting", alias)
			}
		}
	*/

	if _, ok := c.Tenants[tenantId]; !ok {
		return fmt.Errorf("your tenant ID '%s' isn't listed in the config. aborting", tenantId)
	}

	return nil
}

func (c *Nuke) Filters(tenantId string) (filter.Filters, error) {
	tenant := c.Tenants[tenantId]
	filters := tenant.Filters

	if filters == nil {
		filters = filter.Filters{}
	}

	if tenant.Presets == nil {
		return filters, nil
	}

	for _, presetName := range tenant.Presets {
		notFound := fmt.Errorf("could not find filter preset '%s'", presetName)
		if c.Presets == nil {
			return nil, notFound
		}

		preset, ok := c.Presets[presetName]
		if !ok {
			return nil, notFound
		}

		filters.Merge(preset.Filters)
	}

	return filters, nil
}

func (c *Nuke) ResolveDeprecations() error {
	deprecations := map[string]string{}

	for _, t := range c.Tenants {
		for resourceType, resources := range t.Filters {
			replacement, ok := deprecations[resourceType]
			if !ok {
				continue
			}
			log.Warnf("deprecated resource type '%s' - converting to '%s'\n", resourceType, replacement)

			if _, ok := t.Filters[replacement]; ok {
				return fmt.Errorf("using deprecated resource type and replacement: '%s','%s'", resourceType, replacement)
			}

			t.Filters[replacement] = resources
			delete(t.Filters, resourceType)
		}
	}
	return nil
}
