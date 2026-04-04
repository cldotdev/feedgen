package site

import (
	"net/http"
	"net/url"
	"testing"
)

func TestUdnGlobalVisionParser_GetFeed_InDepthColumn(t *testing.T) {
	t.Parallel()
	p := UdnGlobalVisionParser{}

	query := url.Values{}
	query.Set("tag", "in-depth-column")

	feed, err := p.GetFeed(query)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if feed == nil {
		t.Fatal("Feed is nil")
	}

	if feed.Title != "深度專欄 | 轉角國際" {
		t.Errorf("Unexpected feed title: %s", feed.Title)
	}

	if len(feed.Items) == 0 {
		t.Error("Feed has no items")
	}

	t.Logf("Successfully generated feed with %d items", len(feed.Items))
}

func TestUdnGlobalVisionParser_MissingTag(t *testing.T) {
	t.Parallel()
	p := UdnGlobalVisionParser{}

	_, err := p.GetFeed(url.Values{})
	if err == nil {
		t.Error("Expected error for missing tag parameter, got nil")
	}
}

func TestUdnGlobalVisionParser_InvalidTag(t *testing.T) {
	t.Parallel()
	p := UdnGlobalVisionParser{}

	query := url.Values{}
	query.Set("tag", "nonexistent")

	_, err := p.GetFeed(query)
	if err == nil {
		t.Error("Expected error for invalid tag parameter, got nil")
	}
}

func TestUdnGlobalVisionParser_ArticleLinks(t *testing.T) {
	t.Parallel()
	p := UdnGlobalVisionParser{}

	query := url.Values{}
	query.Set("tag", "in-depth-column")

	feed, err := p.GetFeed(query)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if len(feed.Items) == 0 {
		t.Skip("No items to test")
	}

	item := feed.Items[0]
	if item.Link == nil || item.Link.Href == "" {
		t.Error("First item has no link")
		return
	}

	client := &http.Client{}
	resp, err := client.Get(item.Link.Href)
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
