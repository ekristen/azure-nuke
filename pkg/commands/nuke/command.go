package nuke

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/Azure/go-autorest/autorest"

	"github.com/ekristen/azure-nuke/pkg/azure"
	"github.com/ekristen/azure-nuke/pkg/commands"
	"github.com/ekristen/azure-nuke/pkg/common"
	"github.com/ekristen/azure-nuke/pkg/config"
	"github.com/ekristen/azure-nuke/pkg/nuke"
)

func execute(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	logrus.Tracef("tenant id: %s", c.String("tenant-id"))

	ttype, callback, err := azure.AcquireToken(ctx, c.String("tenant-id"))
	if err != nil {
		return err
	}

	var authorizers azure.Authorizers

	if ttype == "token" {
		authorizers.Management, err = callback(c.String("tenant-id"), "https://management.azure.com/")
		if err != nil {
			return err
		}
		authorizers.Graph, err = callback(c.String("tenant-id"), "https://graph.microsoft.com")
		if err != nil {
			return err
		}

	} else {
		authorizers.Management = autorest.NewBearerAuthorizerCallback(nil, callback)
		authorizers.Graph = autorest.NewBearerAuthorizerCallback(nil, callback)
	}

	logrus.Trace("preparing to run nuke")

	params := nuke.NukeParameters{
		ConfigPath: c.Path("config"),
		ForceSleep: c.Int("force-sleep"),
		Quiet:      c.Bool("quiet"),
		NoDryRun:   c.Bool("no-dry-run"),
	}

	tenant, err := azure.NewTenant(ctx, c.String("tenant-id"), authorizers)
	if err != nil {
		return err
	}

	n := nuke.New(params, tenant)

	config, err := config.Load(params.ConfigPath)
	if err != nil {
		return err
	}

	n.Config = config

	return n.Run()
}

func init() {
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:  "config",
			Usage: "path to config file",
			Value: "config.yaml",
		},
		&cli.StringFlag{
			Name:     "tenant-id",
			Usage:    "the tenant-id to nuke",
			EnvVars:  []string{"AZURE_TENANT_ID"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "resource-id",
			EnvVars: []string{"AZURE_RESOURCE_ID"},
			Value:   "https://management.azure.com/",
		},
		&cli.IntFlag{
			Name:  "force-sleep",
			Usage: "seconds to sleep",
			Value: 10,
		},
		&cli.BoolFlag{
			Name:  "quiet",
			Usage: "hide filtered messages",
		},
		&cli.BoolFlag{
			Name:  "no-dry-run",
			Usage: "no dry run",
		},
	}

	cmd := &cli.Command{
		Name:   "nuke",
		Usage:  "nuke an azure tenant",
		Flags:  append(flags, commands.GlobalFlags()...),
		Before: commands.GlobalBefore,
		Action: execute,
	}

	common.RegisterCommand(cmd)
}
