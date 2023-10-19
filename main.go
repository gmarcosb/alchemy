package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/hasty/matterfmt/cmd"
	"github.com/hasty/matterfmt/disco"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {

	logrus.SetLevel(logrus.ErrorLevel)

	cxt := context.Background()

	var dryRun bool
	var serial bool

	var linkAttributes bool
	var dumpAscii bool

	app := &cli.App{
		Name:  "matterfmt",
		Usage: "builds stuff",
		Action: func(c *cli.Context) error {
			return cmd.Format(cxt, c.Args().Slice(), dryRun, serial)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "dryrun",
				Aliases:     []string{"dry"},
				Usage:       "whether or not to actually output files",
				Destination: &dryRun,
			},
			&cli.BoolFlag{
				Name:        "serial",
				Usage:       "process files one-by-one",
				Destination: &serial,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "disco",
				Aliases: []string{"c"},
				Usage:   "Discoball documents",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "linkAttributes",
						Usage:       "whether or not to actually output files",
						Destination: &linkAttributes,
					},
				},
				Action: func(cCtx *cli.Context) error {
					var options []disco.Option
					if linkAttributes {
						options = append(options, disco.LinkAttributes)
					}

					err := cmd.DiscoBall(cxt, cCtx.Args().Slice(), dryRun, serial, options...)
					if err != nil {
						return cli.Exit(err, -1)
					}
					return nil
				},
			},
			{
				Name:    "format",
				Aliases: []string{"fmt"},
				Usage:   "just format Matter documents",
				Action: func(cCtx *cli.Context) error {
					return cmd.Format(cxt, cCtx.Args().Slice(), dryRun, serial)
				},
			},
			{
				Name:  "zcl",
				Usage: "translate Matter spec to ZCL",
				Action: func(cCtx *cli.Context) error {
					return cmd.ZCL(cxt, cCtx.Args().Slice(), dryRun, serial)
				},
			},
			{
				Name:  "db",
				Usage: "just format Matter documents",
				Action: func(cCtx *cli.Context) error {
					return cmd.Database(cxt, cCtx.Args().Slice(), serial)
				},
			},
			{
				Name:    "dump",
				Aliases: []string{"c"},
				Usage:   "dump the parse tree of Matter documents",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "ascii",
						Usage:       "dump asciidoc object model",
						Destination: &dumpAscii,
					},
				},
				Action: func(cCtx *cli.Context) error {
					return cmd.Dump(cxt, cCtx.Args().Slice(), dumpAscii)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("failed running", "error", err)
	}
}
