package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ekristen/azure-nuke/pkg/types"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ResourceTypes struct {
	Targets  types.Collection `yaml:"targets"`
	Excludes types.Collection `yaml:"excludes"`
}

type Tenant struct {
	Filters       Filters       `yaml:"filters"`
	ResourceTypes ResourceTypes `yaml:"resource-types"`
	Presets       []string      `yaml:"presets"`
}

type Nuke struct {
	Tenants         map[string]Tenant            `yaml:"tenants"`
	TenantBlocklist []string                     `yaml:"tenant-blocklist"`
	ResourceTypes   ResourceTypes                `yaml:"resource-types"`
	FeatureFlags    FeatureFlags                 `yaml:"feature-flags"`
	Presets         map[string]PresetDefinitions `yaml:"presets"`
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
	Filters Filters `yaml:"filters"`
}

func Load(path string) (*Nuke, error) {
	var err error

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(Nuke)
	err = yaml.UnmarshalStrict(raw, config)
	if err != nil {
		return nil, err
	}

	if err := config.resolveDeprecations(); err != nil {
		return nil, err
	}

	return config, nil
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

func (c *Nuke) ValidateTenant(tenantId string) error {
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

func (c *Nuke) Filters(tenantId string) (Filters, error) {
	tenant := c.Tenants[tenantId]
	filters := tenant.Filters

	if filters == nil {
		filters = Filters{}
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

func (c *Nuke) resolveDeprecations() error {
	deprecations := map[string]string{
		"EC2DhcpOptions":                "EC2DHCPOptions",
		"EC2InternetGatewayAttachement": "EC2InternetGatewayAttachment",
		"EC2NatGateway":                 "EC2NATGateway",
		"EC2Vpc":                        "EC2VPC",
		"EC2VpcEndpoint":                "EC2VPCEndpoint",
		"EC2VpnConnection":              "EC2VPNConnection",
		"EC2VpnGateway":                 "EC2VPNGateway",
		"EC2VpnGatewayAttachement":      "EC2VPNGatewayAttachment",
		"ECRrepository":                 "ECRRepository",
		"IamGroup":                      "IAMGroup",
		"IamGroupPolicyAttachement":     "IAMGroupPolicyAttachment",
		"IamInstanceProfile":            "IAMInstanceProfile",
		"IamInstanceProfileRole":        "IAMInstanceProfileRole",
		"IamPolicy":                     "IAMPolicy",
		"IamRole":                       "IAMRole",
		"IamRolePolicyAttachement":      "IAMRolePolicyAttachment",
		"IamServerCertificate":          "IAMServerCertificate",
		"IamUser":                       "IAMUser",
		"IamUserAccessKeys":             "IAMUserAccessKey",
		"IamUserGroupAttachement":       "IAMUserGroupAttachment",
		"IamUserPolicyAttachement":      "IAMUserPolicyAttachment",
		"RDSCluster":                    "RDSDBCluster",
	}

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
