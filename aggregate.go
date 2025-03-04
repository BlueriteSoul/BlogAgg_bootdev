package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

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
