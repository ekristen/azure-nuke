package nuke

import (
	"context"
	"io/ioutil"

	authWrapper "github.com/manicminer/hamilton-autorest/auth"
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/environments"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

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

	env, err := environments.EnvironmentFromString(c.String("environment"))
	if err != nil {
		return err
	}

	authConfig := &auth.Config{
		Environment: env,
		TenantID:    c.String("tenant-id"),
		ClientID:    c.String("client-id"),
	}

	if c.String("client-secret") != "" {
		logrus.Debug("authentication type: client secret")
		authConfig.EnableClientSecretAuth = true
		authConfig.ClientSecret = c.String("client-secret")
	} else if c.String("client-certificate-file") != "" {
		logrus.Debug("authentication type: client certificate")
		authConfig.EnableClientCertAuth = true
		authConfig.ClientCertPath = c.String("client-certificate-file")
	} else if c.String("client-federated-token-file") != "" {
		logrus.Debug("authentication type: federated token")
		token, err := ioutil.ReadFile(c.String("client-federated-token-file"))
		if err != nil {
			return err
		}
		authConfig.EnableClientFederatedAuth = true
		authConfig.FederatedAssertion = string(token)
	}

	graphAuthorizer, err := authConfig.NewAuthorizer(ctx, env.MsGraph)
	if err != nil {
		return err
	}

	mgmtAuthorizer, err := authConfig.NewAuthorizer(ctx, env.ResourceManager)
	if err != nil {
		return err
	}

	var authorizers azure.Authorizers

	authorizers.Management = &authWrapper.Authorizer{Authorizer: mgmtAuthorizer}
	authorizers.Graph = graphAuthorizer

	logrus.Trace("preparing to run nuke")

	params := nuke.NukeParameters{
		ConfigPath: c.Path("config"),
		ForceSleep: c.Int("force-sleep"),
		Quiet:      c.Bool("quiet"),
		NoDryRun:   c.Bool("no-dry-run"),
		Targets:    c.StringSlice("only-resource"),
	}

	tenant, err := azure.NewTenant(ctx, authorizers, c.String("tenant-id"), c.StringSlice("subscription-id"))
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
			EnvVars:  []string{"AZURE_CLIENT_ID"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "client-secret",
			EnvVars: []string{"AZURE_CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:    "client-certificate-file",
			EnvVars: []string{"AZURE_CLIENT_CERTIFICATE_FILE"},
		},
		&cli.StringFlag{
			Name:    "client-federated-token-file",
			EnvVars: []string{"AZURE_FEDERATED_TOKEN_FILE"},
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
		&cli.StringSliceFlag{
			Name:  "only-resource",
			Usage: "only resource",
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
