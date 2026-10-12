package main

import (
	"context"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"github.com/ekristen/azure-nuke/pkg/common"

	_ "github.com/ekristen/azure-nuke/pkg/commands/list"
	_ "github.com/ekristen/azure-nuke/pkg/commands/run"

	_ "github.com/ekristen/azure-nuke/resources"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// log panics forces exit
			if _, ok := r.(*logrus.Entry); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	cmd := &cli.Command{
		Name:    path.Base(os.Args[0]),
		Usage:   "remove everything from an azure tenant",
		Version: common.AppVersion.Summary,
		Authors: []any{
			"Erik Kristensen <erik@erikkristensen.com>",
		},
		Commands: common.GetCommands(),
		CommandNotFound: func(_ context.Context, _ *cli.Command, command string) {
			logrus.Fatalf("command %s not found.", command)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}
