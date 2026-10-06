package site

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestThinkingtaiwanParser_GetFeed(t *testing.T) {
	t.Parallel()
	p := ThinkingtaiwanParser{}

	feed, err := p.GetFeed(url.Values{})
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if feed == nil {
		t.Fatal("Feed is nil")
	}

	if feed.Title != thinkingtaiwanFeedTitle {
		t.Errorf("Unexpected feed title: %s", feed.Title)
	}

	if feed.Link == nil || feed.Link.Href != thinkingtaiwanBaseURL+"/articles" {
		t.Errorf("Unexpected feed link: %v", feed.Link)
	}

	if len(feed.Items) == 0 {
		t.Error("Feed has no items")
	}

	t.Logf("Successfully generated feed with %d items", len(feed.Items))
}

func TestThinkingtaiwanParser_ItemFields(t *testing.T) {
	t.Parallel()
	p := ThinkingtaiwanParser{}

	feed, err := p.GetFeed(url.Values{})
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if len(feed.Items) == 0 {
		t.Fatal("Feed has no items")
	}

	for _, item := range feed.Items {
		if item.Title == "" {
			t.Error("Item has empty title")
		}

		if item.Author == nil || item.Author.Name == "" {
			t.Errorf("Item %q has no author", item.Title)
		}

		if item.Link == nil || !strings.HasPrefix(item.Link.Href, thinkingtaiwanBaseURL+"/") {
			t.Errorf("Item %q has no absolute link: %v", item.Title, item.Link)
			continue
		}

		if item.Id != item.Link.Href {
			t.Errorf("Item %q id %q does not match link %q", item.Title, item.Id, item.Link.Href)
		}

		if item.Created.IsZero() {
			t.Errorf("Item %q has zero created timestamp", item.Title)
		}
	}
}

func TestThinkingtaiwanParser_ItemsOrderedNewestFirst(t *testing.T) {
	t.Parallel()
	p := ThinkingtaiwanParser{}

	feed, err := p.GetFeed(url.Values{})
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	for i := 1; i < len(feed.Items); i++ {
		if feed.Items[i].Created.After(feed.Items[i-1].Created) {
			t.Errorf("Item %d (%s) is newer than item %d (%s)",
				i, feed.Items[i].Created, i-1, feed.Items[i-1].Created)
		}
	}
}

func TestThinkingtaiwanParser_ArticleLinks(t *testing.T) {
	t.Parallel()
	p := ThinkingtaiwanParser{}

	feed, err := p.GetFeed(url.Values{})
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

	resp, err := http.Get(item.Link.Href)
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
