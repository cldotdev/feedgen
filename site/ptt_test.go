package site

import (
	"net/http"
	"net/url"
	"testing"
)

// TestPttParser_GetFeed_Steam tests feed generation for PTT Steam board
func TestPttParser_GetFeed_Steam(t *testing.T) {
	parser := PttParser{}

	// Test with b=Steam
	query := url.Values{}
	query.Set("b", "Steam")

	feed, err := parser.GetFeed(query)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if feed == nil {
		t.Fatal("Feed is nil")
	}

	if feed.Title == "" {
		t.Error("Feed title is empty")
	}

	if len(feed.Items) == 0 {
		t.Error("Feed has no items")
	}

	t.Logf("Successfully generated feed with %d items", len(feed.Items))
}

// TestPttParser_GetFeed_Movie tests feed generation for PTT movie board with recommend filter
func TestPttParser_GetFeed_Movie(t *testing.T) {
	parser := PttParser{}

	// Test with b=movie&q=recommend:30
	query := url.Values{}
	query.Set("b", "movie")
	query.Set("q", "recommend:30")

	feed, err := parser.GetFeed(query)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if feed == nil {
		t.Fatal("Feed is nil")
	}

	if feed.Title == "" {
		t.Error("Feed title is empty")
	}

	// Note: With recommend:30 filter, there might be 0 items
	t.Logf("Successfully generated feed with %d items", len(feed.Items))
}

// TestPttParser_ArticleLinks tests HTTP response for article links
func TestPttParser_ArticleLinks(t *testing.T) {
	parser := PttParser{}
	query := url.Values{}
	query.Set("b", "Steam")

	feed, err := parser.GetFeed(query)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if len(feed.Items) == 0 {
		t.Skip("No items to test")
	}

	// Test only the first article link
	item := feed.Items[0]
	if item.Link == nil || item.Link.Href == "" {
		t.Error("First item has no link")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", item.Link.Href, nil)
	if err != nil {
		t.Errorf("Failed to create request for %s: %v", item.Link.Href, err)
		return
	}

	// Add over18 cookie for PTT
	cookie := http.Cookie{Name: "over18", Value: "1"}
	req.AddCookie(&cookie)

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Failed to access article link %s: %v", item.Link.Href, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Errorf("Article link %s returned non-2XX status: %d", item.Link.Href, resp.StatusCode)
	} else {
		t.Logf("Article link %s returned status %d", item.Link.Href, resp.StatusCode)
	}
}

// TestPttParser_MissingParameter tests error handling for missing b parameter
func TestPttParser_MissingParameter(t *testing.T) {
	parser := PttParser{}
	query := url.Values{}

	_, err := parser.GetFeed(query)
	if err == nil {
		t.Error("Expected error for missing b parameter, got nil")
	}
}
