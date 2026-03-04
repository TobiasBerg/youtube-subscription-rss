package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/TobiasBerg/youtube-subscription-rss/config"
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

var scopes = []string{"https://www.googleapis.com/auth/youtube.readonly"}

// AtomFeed represents the root Atom feed structure
type AtomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomEntry represents an Atom feed entry
type AtomEntry struct {
	XMLName   xml.Name   `xml:"entry"`
	ID        string     `xml:"id"`
	VideoID   string     `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
	ChannelID string     `xml:"http://www.youtube.com/xml/schemas/2015 channelId"`
	Title     string     `xml:"title"`
	Link      AtomLink   `xml:"link"`
	Author    AtomAuthor `xml:"author"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
}

// AtomLink represents an Atom link element
type AtomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// AtomAuthor represents an Atom author element
type AtomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

// FeedResult holds the result of fetching a feed
type FeedResult struct {
	ChannelID string
	Feed      string
	Error     error
}

// FeedItem holds parsed feed entry with metadata
type FeedItem struct {
	ChannelID string
	Entry     AtomEntry
	Published time.Time
}

func getChannelFeedURL(channelID string) string {
	return fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID)
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func fetchFeed(ctx context.Context, channelID string) FeedResult {
	url := getChannelFeedURL(channelID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FeedResult{ChannelID: channelID, Error: err}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return FeedResult{ChannelID: channelID, Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FeedResult{
			ChannelID: channelID,
			Error:     fmt.Errorf("HTTP status: %d", resp.StatusCode),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FeedResult{ChannelID: channelID, Error: err}
	}

	return FeedResult{ChannelID: channelID, Feed: string(body)}
}

func extractFeedItems(feedXML, channelID string) []FeedItem {
	var feed AtomFeed

	err := xml.Unmarshal([]byte(feedXML), &feed)
	if err != nil {
		return nil
	}

	items := make([]FeedItem, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		published, err := time.Parse(time.RFC3339, entry.Published)
		if err != nil {
			published = time.Time{}
		}

		items = append(items, FeedItem{
			ChannelID: channelID,
			Entry:     entry,
			Published: published,
		})
	}

	return items
}

func GenerateFeed(ctx context.Context, cfg config.AppConfig) ([]byte, error) {
	youtubeService, err := GetYouTubeService(
		cfg.YoutubeClientID,
		cfg.YoutubeClientSecret,
		cfg.YoutubeRefreshToken,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube service: %w", err)
	}

	log.Println("Loading subscribed channels...")
	channelIDs, err := LoadSubscribedChannelIDs(youtubeService)
	if err != nil {
		return nil, fmt.Errorf("error loading subscriptions: %w", err)
	}
	log.Printf("Found %d subscribed channels\n", len(channelIDs))

	log.Println("Fetching feeds...")
	feeds := make(map[string]FeedResult, len(channelIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Buffered channel to limit concurrent workers
	semaphore := make(chan struct{}, 25)

	for _, channelID := range channelIDs {
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			result := fetchFeed(ctx, cid)
			mu.Lock()
			feeds[cid] = result
			mu.Unlock()
		}(channelID)
	}

	wg.Wait()
	close(semaphore)
	log.Printf("Fetched %d feeds\n", len(feeds))

	var allItems []FeedItem
	failedCount := 0
	for cid, result := range feeds {
		if result.Error != nil {
			log.Printf("Failed to fetch feed for %s: %v\n", cid, result.Error)
			failedCount++
			continue
		}
		items := extractFeedItems(result.Feed, cid)
		for _, item := range items {
			if !cfg.IncludeShorts && strings.Contains(item.Entry.Link.Href, "/shorts/") {
				continue
			}
			allItems = append(allItems, item)
		}
	}

	log.Printf("Extracted %d total items from %d successful feeds (%d failed)\n",
		len(allItems), len(feeds)-failedCount, failedCount)

	if len(allItems) == 0 {
		log.Println("Warning: No items found in any feeds")
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Published.After(allItems[j].Published)
	})

	top100 := allItems
	if len(allItems) > 100 {
		top100 = allItems[:100]
	}

	log.Printf("Keeping top %d items\n", len(top100))

	outputFeed := AtomFeed{
		Title:   "Top 100 Recent YouTube Channel Videos",
		ID:      "merged:youtube:subscriptions",
		Updated: time.Now().UTC().Format(time.RFC3339),
		Entries: make([]AtomEntry, 0, len(top100)),
	}

	for _, item := range top100 {
		entry := item.Entry
		channelName := entry.Author.Name
		entry.Title = fmt.Sprintf("[%s] %s", channelName, entry.Title)
		outputFeed.Entries = append(outputFeed.Entries, entry)
	}

	output, err := xml.MarshalIndent(outputFeed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling XML: %w", err)
	}

	return []byte(xml.Header + string(output)), nil
}
