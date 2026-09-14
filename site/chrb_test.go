package site

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cldotdev/feedgen"
)

// These tests make real HTTP requests, which invite rate limiting and flaky
// results when fired in parallel, so none of them calls t.Parallel.

func TestChrbParser_GetFeed(t *testing.T) {
	p := ChrbParser{}

	feed, err := p.GetFeed(url.Values{})
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	if feed == nil {
		t.Fatal("Feed is nil")
	}

	if feed.Title == "" {
		t.Error("Feed title is empty")
	}

	if feed.Link == nil || feed.Link.Href != chrbBaseURL+"/rental_properties" {
		t.Errorf("Unexpected feed link: %v", feed.Link)
	}

	if len(feed.Items) == 0 {
		t.Error("Feed has no items")
	}

	t.Logf("Successfully generated feed with %d items", len(feed.Items))
}

func TestChrbParser_ItemFields(t *testing.T) {
	p := ChrbParser{}

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

		if item.Link == nil || !strings.HasPrefix(item.Link.Href, chrbBaseURL+"/rental_properties/") {
			t.Errorf("Item %q has no property link: %v", item.Title, item.Link)
			continue
		}

		if item.Id != item.Link.Href {
			t.Errorf("Item %q id %q does not match link %q", item.Title, item.Id, item.Link.Href)
		}

		if item.Created.IsZero() {
			t.Errorf("Item %q has zero created timestamp", item.Title)
		}

		for _, field := range []string{"地址", "類型", "坪數", "樓層", "租金"} {
			if !strings.Contains(item.Description, field+": ") {
				t.Errorf("Item %q description is missing %s: %s", item.Title, field, item.Description)
			}
		}
	}
}

func TestChrbParser_SortNewestOrdersByDate(t *testing.T) {
	p := ChrbParser{}

	query := url.Values{}
	query.Set("sort", "newest")

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

func TestChrbSortOrders(t *testing.T) {
	expected := map[string]string{
		"newest":    "published_at desc",
		"updated":   "updated_at desc",
		"rent_asc":  "rent asc",
		"rent_desc": "rent desc",
		"size_asc":  "area_ping asc",
		"size_desc": "area_ping desc",
	}

	if len(chrbSortOrders) != len(expected) {
		t.Errorf("Expected %d sort values, got %d", len(expected), len(chrbSortOrders))
	}

	for sort, order := range expected {
		if chrbSortOrders[sort] != order {
			t.Errorf("Sort %s maps to %q, expected %q", sort, chrbSortOrders[sort], order)
		}
	}
}

// The listing's default ordering is already newest first, so sort=newest alone
// cannot show that the ordering reached the site. Rent can: the cheapest
// properties never head the default page.
func TestChrbParser_SortRentAscOrdersByRent(t *testing.T) {
	p := ChrbParser{}

	query := url.Values{}
	query.Set("sort", "rent_asc")

	feed, err := p.GetFeed(query)
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	previous := -1
	for _, item := range feed.Items {
		rent, ok := chrbItemRent(item.Description)
		if !ok {
			t.Errorf("Item %q description has no rent: %s", item.Title, item.Description)
			continue
		}

		if rent < previous {
			t.Errorf("Item %q rent %d follows a higher rent %d", item.Title, rent, previous)
		}

		previous = rent
	}
}

func TestChrbParser_InvalidSort(t *testing.T) {
	p := ChrbParser{}

	query := url.Values{}
	query.Set("sort", "cheapest")

	_, err := p.GetFeed(query)

	var invalid *feedgen.ParameterValueInvalidError
	if !errors.As(err, &invalid) {
		t.Fatalf("Expected ParameterValueInvalidError, got %v", err)
	}

	if invalid.Parameter != "sort" {
		t.Errorf("Expected error for parameter sort, got %s", invalid.Parameter)
	}
}

// The 精選房源 sidebar reuses the property card's anchor class but titles its
// cards with <h3>, so counting the listing's own <h2> titles bounds how many
// items a leak-free feed can hold.
func TestChrbParser_ExcludesSidebarProperties(t *testing.T) {
	p := ChrbParser{}

	feed, err := p.GetFeed(url.Values{})
	if err != nil {
		t.Fatalf("Failed to get feed: %v", err)
	}

	listed := strings.Count(chrbListingPage(t), "<h2")
	if listed == 0 {
		t.Fatal("Listing page shows no property title")
	}

	if len(feed.Items) > listed {
		t.Errorf("Feed holds %d items but the listing shows %d properties", len(feed.Items), listed)
	}
}

func TestChrbParser_PropertyLinks(t *testing.T) {
	p := ChrbParser{}

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

	client := &http.Client{}
	resp, err := client.Get(item.Link.Href)
	if err != nil {
		t.Errorf("Failed to access property link %s: %v", item.Link.Href, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Errorf("Property link %s returned non-2XX status: %d", item.Link.Href, resp.StatusCode)
	} else {
		t.Logf("Property link %s returned status %d", item.Link.Href, resp.StatusCode)
	}
}

func chrbListingPage(t *testing.T) string {
	t.Helper()

	link := chrbBaseURL + "/rental_properties"

	resp, err := client.Get(link)
	if err != nil {
		t.Fatalf("Failed to fetch %s: %v", link, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", link, err)
	}

	return chrbSearchResults(string(body))
}

var chrbItemRentRe = regexp.MustCompile(`租金: ([\d,]+)`)

func chrbItemRent(description string) (rent int, ok bool) {
	match := chrbItemRentRe.FindStringSubmatch(description)
	if match == nil {
		return
	}

	rent, err := strconv.Atoi(strings.ReplaceAll(match[1], ",", ""))
	if err != nil {
		return 0, false
	}

	return rent, true
}
