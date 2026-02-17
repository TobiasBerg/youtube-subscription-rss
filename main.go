package main

import (
	"context"
	"log"
	"os"

	"github.com/TobiasBerg/youtube-subscription-rss/cmd"
	"github.com/TobiasBerg/youtube-subscription-rss/config"
	"github.com/urfave/cli/v3"
)

func main() {
	cfg, err := config.CreateConfig()
	if err != nil {
		panic(err)
	}

	startCMD := &cli.Command{
		Name:   "start",
		Usage:  "Start the application server (default command)",
		Action: cmd.CreateFeedCMD(cfg),
	}

	listSubscriptionsCMD := &cli.Command{
		Name:   "create-secret",
		Usage:  "",
		Action: cmd.CreateSecretCMD,
	}

	app := cli.Command{
		Name:           "youtube-subscription-rss",
		Description:    "Generate an Atom feed of the most recent videos from your YouTube subscriptions",
		DefaultCommand: "start",
		Commands: []*cli.Command{
			startCMD,
			listSubscriptionsCMD,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
