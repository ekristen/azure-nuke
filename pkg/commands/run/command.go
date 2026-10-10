package run

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	libconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/filter"
	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/scanner"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/commands/global"
	"github.com/ekristen/azure-nuke/pkg/common"
	"github.com/ekristen/azure-nuke/pkg/config"
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

func execute(ctx context.Context, cmd *cli.Command) error { //nolint:funlen,gocyclo
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// This is to purposefully capture the output from the standard logger that is written to by several
	// of the azure sdk golang libraries by hashicorp
	log.SetOutput(&log2LogrusWriter{
		entry: logrus.WithField("source", "standard-logger"),
	})

	logger := logrus.StandardLogger()
	logger.SetOutput(os.Stdout)

	logger.Tracef("tenant id: %s", cmd.String("tenant-id"))

	if cmd.String("client-id") == "" &&
		(cmd.String("client-secret") != "" || cmd.String("client-certificate-file") != "" || cmd.String("client-federated-token-file") != "") {
		return fmt.Errorf("--client-id is required when using --client-secret, --client-certificate-file, or --client-federated-token-file")
	}

	authorizers, err := azure.ConfigureAuth(ctx,
		cmd.String("environment"), cmd.String("tenant-id"), cmd.String("client-id"),
		cmd.String("client-secret"), cmd.String("client-certificate-file"),
		cmd.String("client-federated-token-file"))
	if err != nil {
		return err
	}

	logger.Trace("preparing to run nuke")

	params := &libnuke.Parameters{
		Force:              cmd.Bool("no-prompt"),
		ForceSleep:         int(cmd.Int("prompt-delay")), //nolint:unconvert
		Quiet:              cmd.Bool("quiet"),
		NoDryRun:           cmd.Bool("no-dry-run"),
		Includes:           cmd.StringSlice("include"),
		Excludes:           cmd.StringSlice("exclude"),
		WaitOnDependencies: cmd.Bool("wait-on-dependencies"),
		MaxWaitRetries:     int(cmd.Int("max-wait-retries")), //nolint:unconvert
	}

	parsedConfig, err := config.New(libconfig.Options{
		Path:         cmd.String("config"),
		Deprecations: registry.GetDeprecatedResourceTypeMapping(),
		Log:          logger.WithField("component", "config"),
	})
	if err != nil {
		logger.Errorf("Failed to parse config file %s", cmd.String("config"))
		return err
	}

	if err := parsedConfig.ValidateAccount(cmd.String("tenant-id")); err != nil {
		return err
	}

	// Azure returns subscription ids in lowercase, the blocklist may not be
	inBlocklist := func(id string) bool {
		return slices.ContainsFunc(parsedConfig.ResolveBlocklist(), func(b string) bool { return strings.EqualFold(b, id) })
	}

	for _, subscriptionID := range cmd.StringSlice("subscription-id") {
		if inBlocklist(subscriptionID) {
			return fmt.Errorf("subscription %s is in the blocklist", subscriptionID)
		}
	}

	tenant, err := azure.NewTenant(ctx,
		authorizers, cmd.String("tenant-id"), cmd.StringSlice("subscription-id"), parsedConfig.Regions)
	if err != nil {
		return err
	}

	// Without --subscription-id every subscription of the tenant is nuked, except the blocklisted ones
	for _, subscriptionID := range tenant.RemoveSubscriptions(inBlocklist) {
		logger.Warnf("skipping subscription %s (reason: in the blocklist)", subscriptionID)
	}

	filters, err := parsedConfig.Filters(cmd.String("tenant-id"))
	if err != nil {
		return err
	}

	// Setup Region Filters as Global Filters
	if len(filters[filter.Global]) == 0 {
		filters[filter.Global] = []filter.Filter{}
	}
	if !slices.Contains(parsedConfig.Regions, "all") {
		filters[filter.Global] = append(filters[filter.Global], filter.Filter{
			Property: "Region",
			Type:     filter.NotIn,
			Values:   parsedConfig.Regions,
		})
	}

	// Initialize the underlying nuke process
	n := libnuke.New(params, filters, parsedConfig.Settings)

	n.SetRunSleep(5 * time.Second)
	n.SetLogger(logger.WithField("component", "nuke"))

	n.RegisterVersion(fmt.Sprintf("> %s", common.AppVersion.String()))

	p := &azure.Prompt{Parameters: params, Tenant: tenant}
	n.RegisterPrompt(p.Prompt)

	tenantConfig := parsedConfig.Accounts[cmd.String("tenant-id")]
	tenantResourceTypes := types.ResolveResourceTypes(
		registry.GetNamesForScope(azure.TenantScope),
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
		registry.GetNamesForScope(azure.SubscriptionScope),
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
		registry.GetNamesForScope(azure.ResourceGroupScope),
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

	if slices.Contains(parsedConfig.Regions, "global") || slices.Contains(parsedConfig.Regions, "all") {
		tenantScanner, scanErr := scanner.New(&scanner.Config{
			Owner:         "tenant",
			ResourceTypes: tenantResourceTypes,
			Opts: &azure.ListerOpts{
				Authorizers: authorizers,
				TenantID:    tenant.ID,
			},
			Logger: logger,
		})
		if scanErr != nil {
			return scanErr
		}

		if err := n.RegisterScanner(azure.TenantScope, tenantScanner); err != nil {
			return err
		}

		logger.
			WithField("component", "run").
			WithField("scope", "tenant").
			Debug("registering scanner")
		for _, subscriptionID := range tenant.SubscriptionIds {
			logger.
				WithField("component", "run").
				WithField("scope", "subscription").
				WithField("subscription_id", subscriptionID).
				Debug("registering scanner")

			parts := strings.Split(subscriptionID, "-")
			subScanner, scanErr := scanner.New(&scanner.Config{
				Owner:         fmt.Sprintf("sub/%s", parts[:1][0]),
				ResourceTypes: subResourceTypes,
				Opts: &azure.ListerOpts{
					Authorizers:    tenant.Authorizers,
					TenantID:       tenant.ID,
					SubscriptionID: subscriptionID,
					Regions:        parsedConfig.Regions,
				},
				Logger: logger,
			})
			if scanErr != nil {
				return scanErr
			}

			if err := n.RegisterScanner(azure.SubscriptionScope, subScanner); err != nil {
				return err
			}
		}
	}

	for subscriptionID, resourceGroups := range tenant.ResourceGroups {
		for _, rg := range resourceGroups {
			logger.
				WithField("component", "run").
				WithField("scope", "resource-group").
				WithField("subscription_id", subscriptionID).
				WithField("resource_group", rg).
				Debug("registering scanner")

			rgScanner, scanErr := scanner.New(&scanner.Config{
				Owner:         fmt.Sprintf("sub/%s/rg/%s", subscriptionID, rg),
				ResourceTypes: rgResourceTypes,
				Opts: &azure.ListerOpts{
					Authorizers:    tenant.Authorizers,
					TenantID:       tenant.ID,
					SubscriptionID: subscriptionID,
					ResourceGroup:  rg,
					Regions:        parsedConfig.Regions,
				},
				Logger: logger,
			})
			if scanErr != nil {
				return scanErr
			}

			if err := n.RegisterScanner(azure.ResourceGroupScope, rgScanner); err != nil {
				return err
			}
		}
	}

	logrus.Debug("running ...")

	return n.Run(ctx)
}

func init() {
	flags := []cli.Flag{
		&cli.StringFlag{
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
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "hide filtered messages",
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
		&cli.IntFlag{
			Name:  "max-wait-retries",
			Usage: "maximum number of wait rounds for resources being removed before failing (0 = unlimited)",
		},
		&cli.BoolFlag{
			Name:  "wait-on-dependencies",
			Usage: "wait for dependent resources to be deleted before deleting resources that depend on them",
		},
		&cli.StringSliceFlag{
			Name:    "feature-flag",
			Usage:   "enable experimental behaviors that may not be fully tested or supported",
			Sources: cli.EnvVars("AZURE_NUKE_FEATURE_FLAGS"),
		},
		&cli.StringFlag{
			Name:    "environment",
			Usage:   "Azure Environment",
			Sources: cli.EnvVars("AZURE_ENVIRONMENT"),
			Value:   "global",
		},
		&cli.StringFlag{
			Name:     "tenant-id",
			Usage:    "the tenant-id to nuke",
			Sources:  cli.EnvVars("AZURE_TENANT_ID"),
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     "subscription-id",
			Usage:    "the subscription-id to nuke (this filters to 1 or more subscription ids)",
			Sources:  cli.EnvVars("AZURE_SUBSCRIPTION_ID"),
			Required: false,
		},
		&cli.StringFlag{
			Name:    "client-id",
			Usage:   "the client-id to use for authentication (optional when using Azure CLI auth)",
			Sources: cli.EnvVars("AZURE_CLIENT_ID"),
		},
		&cli.StringFlag{
			Name:    "client-secret",
			Usage:   "the client-secret to use for authentication",
			Sources: cli.EnvVars("AZURE_CLIENT_SECRET"),
		},
		&cli.StringFlag{
			Name:    "client-certificate-file",
			Usage:   "the client-certificate-file to use for authentication",
			Sources: cli.EnvVars("AZURE_CLIENT_CERTIFICATE_FILE"),
		},
		&cli.StringFlag{
			Name:    "client-federated-token-file",
			Usage:   "the client-federated-token-file to use for authentication",
			Sources: cli.EnvVars("AZURE_FEDERATED_TOKEN_FILE"),
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
