package cmd

import (
	"context"

	"github.com/TobiasBerg/youtube-subscription-rss/config"
	"github.com/TobiasBerg/youtube-subscription-rss/service"
	"github.com/urfave/cli/v3"
)

func CreateFeedCMD(cfg config.AppConfig) func(ctx context.Context, c *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		err := service.GenerateFeed(ctx, cfg)
		if err != nil {
			return err
		}

		return nil
	}
}
