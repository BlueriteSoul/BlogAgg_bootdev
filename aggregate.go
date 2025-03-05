package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/BlueriteSoul/BlogAgg_bootdev/internal/database"
	"github.com/google/uuid"
)

func parsePubDate(pubDateStr string) (time.Time, error) {
	// Try parsing with common formats
	formats := []string{
		time.RFC1123,
		time.RFC3339,
		time.RFC822,
		"2006-01-02 15:04:05", // Add any other formats you encounter
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, pubDateStr)
		if err == nil {
			return t, nil
		}
	}

	// If no format matches, return an error
	return time.Time{}, fmt.Errorf("failed to parse pubDate: %s", pubDateStr)
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var result RSSFeed
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't complete the GET request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Couldn't fetch the request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading the response: %w", err)
	}
	err = xml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failiure: %w", err)
	}
	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)
	for i := range result.Channel.Item {
		result.Channel.Item[i].Title = html.UnescapeString(result.Channel.Item[i].Title)
		result.Channel.Item[i].Description = html.UnescapeString(result.Channel.Item[i].Description)
	}
	return &result, nil
}

func scrapeFeeds(s *state) error {
	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get next feed to fetch, error: %w", err)
	}
	err = s.db.MarkFeedFetched(context.Background(), feedToFetch.ID)
	if err != nil {
		return fmt.Errorf("Failed to mark feed as fetched, error: %w", err)
	}
	fetchedFeed, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		return fmt.Errorf("Failed fetch feed, error: %w", err)
	}
	/*for _, item := range fetchedFeed.Channel.Item {
		fmt.Println(item.Title)
	}*/
	for _, item := range fetchedFeed.Channel.Item {
		var nullDescription sql.NullString

		if item.Description != "" {
			nullDescription = sql.NullString{
				String: item.Description,
				Valid:  true,
			}
		} else {
			nullDescription = sql.NullString{
				String: "",
				Valid:  false,
			}
		}
		pubdate, err := parsePubDate(item.PubDate)
		if err != nil {
			return fmt.Errorf("Failed parse time, error: %w", err)
		}
		post := database.CreatePostParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Title: item.Title, Url: item.Link, Description: nullDescription, PublishedAt: pubdate, FeedID: feedToFetch.ID}
		s.db.CreatePost(context.Background(), post)
		fmt.Println(item.Title)
	}
	return nil
}
