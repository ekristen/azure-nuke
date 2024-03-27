package config

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/config"
)

// New creates a new extended configuration from a file. This is necessary because we are extended the default
// libnuke configuration to contain additional attributes that are specific to the AWS Nuke tool.
func New(opts config.Options) (*Config, error) {
	// Step 1 - Create the libnuke config
	cfg, err := config.New(opts)
	if err != nil {
		return nil, err
	}

	// Step 2 - Instantiate the extended config
	c := &Config{}

	// Step 3 - Load the same config file against the extended config
	if err := c.Load(opts.Path); err != nil {
		return nil, err
	}

	// Step 4 - Set the libnuke config on the extended config
	c.Config = *cfg

	if len(c.Tenants) > 0 {
		logrus.Warn("use of deprecated `tenants` configuration key. Please use `accounts` instead")

		if len(c.Accounts) > 0 && len(c.Tenants) > 0 {
			return nil, fmt.Errorf("cannot use both `accounts` and `tenants` configuration keys")
		}

		c.Accounts = c.Tenants
	}

	if len(c.TenantBlocklist) > 0 {
		logrus.Warn("use of deprecated `tenant-blocklist` configuration key. Please use `blocklist` instead")

		if len(c.Blocklist) > 0 && len(c.TenantBlocklist) > 0 {
			return nil, fmt.Errorf("cannot use both `blocklist` and `tenant-blocklist` configuration keys")
		}

		c.Blocklist = c.TenantBlocklist
	}

	return c, nil
}

// Config is an extended configuration implementation that adds some additional features on top of the libnuke config.
type Config struct {
	// Config is the underlying libnuke configuration.
	config.Config `yaml:",inline"`

	// These are tenants that are configured. There is a more generic Accounts in the config that is available that
	// should be used instead of this.
	// Deprecated: Use Accounts instead. Will be removed in 2.x
	Tenants map[string]*config.Account `yaml:"tenants"`

	// TenantBlocklist is a list of tenant IDs that should be blocklisted. This is used to prevent you from accidentally
	// nuking your production account.
	// Deprecated: Use Blocklist instead. Will be removed in 2.x
	TenantBlocklist []string `yaml:"tenant-blocklist"`
}
