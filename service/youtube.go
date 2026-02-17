package service

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var Scopes = []string{"https://www.googleapis.com/auth/youtube.readonly"}

func LoadCredentials(credentialsFile string) (*oauth2.Config, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %w", err)
	}

	return config, nil
}

func GetTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n", authURL)
	fmt.Print("Enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return token, nil
}

func GetYouTubeService(clientID, clientSecret, refreshToken string) (*youtube.Service, error) {
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, fmt.Errorf("YouTube credentials are not set in environment variables")
	}

	googleConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       Scopes,
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	ctx := context.Background()
	client := googleConfig.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube service: %w", err)
	}

	return service, nil
}

func LoadSubscribedChannelIDs(service *youtube.Service) ([]string, error) {
	var channelIDs []string
	pageToken := ""

	for {
		call := service.Subscriptions.List([]string{"snippet"}).
			Mine(true).
			MaxResults(50)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("error fetching subscriptions: %w", err)
		}

		for _, item := range response.Items {
			channelID := item.Snippet.ResourceId.ChannelId
			channelIDs = append(channelIDs, channelID)
		}

		pageToken = response.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return channelIDs, nil
}
