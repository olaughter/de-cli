package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

type shortcutStoryConfig struct {
	withTimes          bool
	numberOfStories    int
}

func main() {
	// shortcut
	var deleteApiKey bool
	var setApiKey bool
	var scStoryConfig shortcutStoryConfig

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "sc",
				Usage: "shortcut related commands",
				Subcommands: []*cli.Command{
					{
						Name:    "auth",
						Aliases: []string{"a"},
						Usage:   "Set or delete the shortcut ApiKey for de-cli",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:        "delete",
								Aliases:     []string{"d"},
								Usage:       "delete the shortcut ApiKey",
								Value:       false,
								Destination: &deleteApiKey,
							},
							&cli.BoolFlag{
								Name:        "set",
								Aliases:     []string{"s"},
								Usage:       "Set the shortcut ApiKey",
								Value:       false,
								Destination: &setApiKey,
							},
						},
						Action: func(cCtx *cli.Context) error {
							if deleteApiKey {
								err := deleteShortcutApiKey()
								if err != nil {
									return cli.Exit(err, 1)
								}
							} else if setApiKey {
								_, err := setShortcutApiKey()
								if err != nil {
									return cli.Exit(err, 1)
								}
							} else {
								return cli.Exit("Please specify either --delete or --set", 1)
							}
							return nil
						},
					},
					{
						Name:    "stories",
						Aliases: []string{"s"},
						Usage:   "prints headline details of shortcut stories belonging to me",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:        "with-times",
								Aliases:     []string{"wt"},
								Usage:       "Include timestamps of when the story last moved in the output",
								Value:       false,
								Destination: &scStoryConfig.withTimes,
							},
							&cli.IntFlag{
								Name:        "number",
								Aliases:     []string{"n", "num"},
								Usage:       "Number of stories to print, max 25",
								Value:       8,
								Destination: &scStoryConfig.numberOfStories,
							},
						},
						Action: func(cCtx *cli.Context) error {
							err := myStories(scStoryConfig)
							if err != nil {
								return cli.Exit(err, 1)
							}
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
