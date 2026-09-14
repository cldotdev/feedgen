package site

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestScitechvistaParser_GetFeed_Sections(t *testing.T) {
	t.Parallel()

	for slug, feedTitle := range scitechvistaFeedTitles {
		t.Run(slug, func(t *testing.T) {
			t.Parallel()
			p := ScitechvistaParser{}

			query := url.Values{}
			query.Set("section", slug)

			feed, err := p.GetFeed(query)
			if err != nil {
				t.Fatalf("Failed to get feed: %v", err)
			}

			if feed == nil {
				t.Fatal("Feed is nil")
			}

			if feed.Title != feedTitle {
				t.Errorf("Unexpected feed title: %s", feed.Title)
			}

			if feed.Link == nil || feed.Link.Href != scitechvistaBaseURL+"/Article/C000003/"+slug {
				t.Errorf("Unexpected feed link: %v", feed.Link)
			}

			if len(feed.Items) == 0 {
				t.Error("Feed has no items")
			}

			t.Logf("Successfully generated feed with %d items", len(feed.Items))
		})
	}
}

func TestScitechvistaParser_MissingSection(t *testing.T) {
	t.Parallel()
	p := ScitechvistaParser{}

	_, err := p.GetFeed(url.Values{})
	if err == nil {
		t.Error("Expected error for missing section parameter, got nil")
	}
}

func TestScitechvistaParser_InvalidSection(t *testing.T) {
	t.Parallel()
	p := ScitechvistaParser{}

	query := url.Values{}
	query.Set("section", "nonexistent")

	_, err := p.GetFeed(query)
	if err == nil {
		t.Error("Expected error for invalid section parameter, got nil")
	}
}

func TestScitechvistaParser_ItemFields(t *testing.T) {
	t.Parallel()
	p := ScitechvistaParser{}

	query := url.Values{}
	query.Set("section", "new")

	feed, err := p.GetFeed(query)
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

		if item.Link == nil || !strings.HasPrefix(item.Link.Href, scitechvistaBaseURL+"/") {
			t.Errorf("Item %q has no absolute link: %v", item.Title, item.Link)
			continue
		}

		if item.Id != item.Link.Href {
			t.Errorf("Item %q id %q does not match link %q", item.Title, item.Id, item.Link.Href)
		}

		if item.Created.IsZero() {
			t.Errorf("Item %q has zero created timestamp", item.Title)
		}

		if item.Created.Year() <= 2000 {
			t.Errorf("Item %q created in year %d, ROC year was likely not converted", item.Title, item.Created.Year())
		}
	}
}

// The listing page renders a 推薦文章 sidebar of older articles alongside the
// section's own list. Leaking those entries breaks the newest-first order the
// site publishes, so this ordering check doubles as a sidebar leak check.
func TestScitechvistaParser_ItemsOrderedNewestFirst(t *testing.T) {
	t.Parallel()
	p := ScitechvistaParser{}

	query := url.Values{}
	query.Set("section", "new")

	feed, err := p.GetFeed(query)
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

func TestScitechvistaParser_ArticleLinks(t *testing.T) {
	t.Parallel()
	p := ScitechvistaParser{}

	query := url.Values{}
	query.Set("section", "new")

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
