package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/TobiasBerg/youtube-subscription-rss/config"
	"github.com/TobiasBerg/youtube-subscription-rss/service"
	"github.com/urfave/cli/v3"
)

func CreateFeedCMD(cfg config.AppConfig) func(ctx context.Context, c *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		data, err := service.GenerateFeed(ctx, cfg)
		if err != nil {
			return err
		}

		if err := os.MkdirAll("outputs", 0o755); err != nil {
			return fmt.Errorf("error creating outputs directory: %w", err)
		}

		if err := os.WriteFile("outputs/feed.xml", data, 0o644); err != nil {
			return fmt.Errorf("error writing feed.xml: %w", err)
		}

		return nil
	}
}
