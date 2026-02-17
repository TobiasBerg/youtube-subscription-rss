package cmd

import (
	"context"
	"fmt"

	"github.com/TobiasBerg/youtube-subscription-rss/service"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"
)

const (
	credentialsFile = "configs/credentials.json"
)

// Subscription represents a YouTube subscription
type Subscription struct {
	ChannelID string `json:"channelId"`
	Title     string `json:"title"`
}

// TokenData represents the OAuth token stored in token.json
type TokenData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       string `json:"expiry"`
}

func getRefreshToken(ctx context.Context) (*oauth2.Token, error) {
	config, err := service.LoadCredentials(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	token, err := service.GetTokenFromWeb(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %w", err)
	}

	fmt.Println(fmt.Sprintln("REFRESH TOKEN: %s", token.RefreshToken))
	return token, nil
}

func CreateSecretCMD(ctx context.Context, c *cli.Command) error {
	token, err := getRefreshToken(ctx)
	if err != nil {
		return nil
	}

	fmt.Printf("Refresh token: %s\n", token.RefreshToken)
	return nil
}
