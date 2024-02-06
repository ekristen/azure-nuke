package nuke

import (
	"context"
	"fmt"
	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/commands/global"
	"github.com/ekristen/azure-nuke/pkg/common"
	"github.com/ekristen/azure-nuke/pkg/config"
	"github.com/ekristen/azure-nuke/pkg/nuke"
	libconfig "github.com/ekristen/libnuke/pkg/config"
	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"slices"
	"time"
)

type log2LogrusWriter struct {
	entry *logrus.Entry
}

func (w *log2LogrusWriter) Write(b []byte) (int, error) {
	n := len(b)
	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	w.entry.Trace(string(b))
	return n, nil
}

func execute(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	// This is to purposefully capture the output from the standard logger that is written to by several
	// of the azure sdk golang libraries by hashicorp
	log.SetOutput(&log2LogrusWriter{
		entry: logrus.WithField("source", "standard-logger"),
	})

	logrus.Tracef("tenant id: %s", c.String("tenant-id"))

	authorizers, err := azure.ConfigureAuth(ctx,
		c.String("environment"), c.String("tenant-id"), c.String("client-id"),
		c.String("client-secret"), c.String("client-certificate-file"),
		c.String("client-federated-token-file"))
	if err != nil {
		return err
	}

	logrus.Trace("preparing to run nuke")

	params := &libnuke.Parameters{
		Force:      c.Bool("force"),
		ForceSleep: c.Int("force-sleep"),
		Quiet:      c.Bool("quiet"),
		NoDryRun:   c.Bool("no-dry-run"),
		Includes:   c.StringSlice("include"),
		Excludes:   c.StringSlice("exclude"),
	}

	parsedConfig, err := config.New(libconfig.Options{
		Path:         c.Path("config"),
		Deprecations: resource.GetDeprecatedResourceTypeMapping(),
	})
	if err != nil {
		return err
	}

	tenant, err := azure.NewTenant(ctx,
		authorizers, c.String("tenant-id"), c.StringSlice("subscription-id"), parsedConfig.Regions)
	if err != nil {
		return err
	}

	filters, err := parsedConfig.Filters(c.String("tenant-id"))
	if err != nil {
		return err
	}

	n := libnuke.New(params, filters, parsedConfig.Settings)

	n.SetRunSleep(5 * time.Second)

	tenantConfig := parsedConfig.Accounts[c.String("tenant-id")]
	tenantResourceTypes := types.ResolveResourceTypes(
		resource.GetNamesForScope(nuke.Tenant),
		[]types.Collection{
			n.Parameters.Includes,
			parsedConfig.ResourceTypes.GetIncludes(),
			tenantConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			n.Parameters.Excludes,
			parsedConfig.ResourceTypes.Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
		nil,
		nil,
	)

	subResourceTypes := types.ResolveResourceTypes(
		resource.GetNamesForScope(nuke.Subscription),
		[]types.Collection{
			n.Parameters.Includes,
			parsedConfig.ResourceTypes.GetIncludes(),
			tenantConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			n.Parameters.Excludes,
			parsedConfig.ResourceTypes.Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
		nil,
		nil,
	)

	rgResourceTypes := types.ResolveResourceTypes(
		resource.GetNamesForScope(nuke.ResourceGroup),
		[]types.Collection{
			n.Parameters.Includes,
			parsedConfig.ResourceTypes.GetIncludes(),
			tenantConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			n.Parameters.Excludes,
			parsedConfig.ResourceTypes.Excludes,
			tenantConfig.ResourceTypes.Excludes,
		},
		nil,
		nil,
	)

	if slices.Contains(parsedConfig.Regions, "global") {
		if err := n.RegisterScanner(nuke.Tenant, libnuke.NewScanner("tenant/all", tenantResourceTypes, &nuke.ListerOpts{
			Authorizers: authorizers,
			TenantId:    tenant.ID,
		})); err != nil {
			return err
		}
	}

	logrus.Debug("registering scanner for tenant subscription resources")
	for _, subscriptionId := range tenant.SubscriptionIds {
		logrus.Debug("registering scanner for subscription resources")
		if err := n.RegisterScanner(nuke.Subscription, libnuke.NewScanner("tenant/sub", subResourceTypes, &nuke.ListerOpts{
			Authorizers:    tenant.Authorizers,
			TenantId:       tenant.ID,
			SubscriptionId: subscriptionId,
		})); err != nil {
			return err
		}
	}

	for _, region := range parsedConfig.Regions {
		if region == "global" {
			continue
		}

		for _, subscriptionId := range tenant.SubscriptionIds {
			logrus.Debug("registering scanner for subscription resources")

			for i, resourceGroup := range tenant.ResourceGroups[subscriptionId] {
				logrus.Debugf("registering scanner for resource group resources: rg/%s", resourceGroup)
				if err := n.RegisterScanner(nuke.ResourceGroup, libnuke.NewScanner(fmt.Sprintf("%s/rg%d", region, i), rgResourceTypes, &nuke.ListerOpts{
					Authorizers:    tenant.Authorizers,
					TenantId:       tenant.ID,
					SubscriptionId: subscriptionId,
					ResourceGroup:  resourceGroup,
					Location:       region,
				})); err != nil {
					return err
				}
			}
		}
	}

	logrus.Debug("running ...")
	return n.Run(c.Context)
}

func init() {
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:  "config",
			Usage: "path to config file",
			Value: "config.yaml",
		},
		&cli.StringSliceFlag{
			Name:  "include",
			Usage: "only include this specific resource",
		},
		&cli.StringSliceFlag{
			Name:  "exclude",
			Usage: "exclude this specific resource (this overrides everything)",
		},
		&cli.BoolFlag{
			Name:  "quiet",
			Usage: "hide filtered messages",
		},
		&cli.BoolFlag{
			Name:  "no-dry-run",
			Usage: "actually run the removal of the resources after discovery",
		},
		&cli.BoolFlag{
			Name:    "no-prompt",
			Usage:   "disable prompting for verification to run",
			Aliases: []string{"force"},
		},
		&cli.IntFlag{
			Name:    "prompt-delay",
			Usage:   "seconds to delay after prompt before running (minimum: 3 seconds)",
			Value:   10,
			Aliases: []string{"force-sleep"},
		},
		&cli.StringSliceFlag{
			Name:  "feature-flag",
			Usage: "enable experimental behaviors that may not be fully tested or supported",
		},
		&cli.StringFlag{
			Name:    "environment",
			Usage:   "Azure Environment",
			EnvVars: []string{"AZURE_ENVIRONMENT"},
			Value:   "global",
		},
		&cli.StringFlag{
			Name:     "tenant-id",
			Usage:    "the tenant-id to nuke",
			EnvVars:  []string{"AZURE_TENANT_ID"},
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     "subscription-id",
			Usage:    "the subscription-id to nuke (this filters to 1 or more subscription ids)",
			EnvVars:  []string{"AZURE_SUBSCRIPTION_ID"},
			Required: false,
		},
		&cli.StringFlag{
			Name:     "client-id",
			Usage:    "the client-id to use for authentication",
			EnvVars:  []string{"AZURE_CLIENT_ID"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "client-secret",
			Usage:   "the client-secret to use for authentication",
			EnvVars: []string{"AZURE_CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:    "client-certificate-file",
			Usage:   "the client-certificate-file to use for authentication",
			EnvVars: []string{"AZURE_CLIENT_CERTIFICATE_FILE"},
		},
		&cli.StringFlag{
			Name:    "client-federated-token-file",
			Usage:   "the client-federated-token-file to use for authentication",
			EnvVars: []string{"AZURE_FEDERATED_TOKEN_FILE"},
		},
	}

	cmd := &cli.Command{
		Name:    "run",
		Aliases: []string{"nuke"},
		Usage:   "run nuke against an azure tenant to remove all configured resources",
		Flags:   append(flags, global.Flags()...),
		Before:  global.Before,
		Action:  execute,
	}

	common.RegisterCommand(cmd)
}
