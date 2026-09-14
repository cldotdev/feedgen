package site

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

const chrbBaseURL = "https://chrb.com.tw"

// chrbSortOrders maps the feed's own sort names onto the Ransack orderings the
// listing's own <select> offers, so a feed URL never carries the site's
// internal parameter names.
var chrbSortOrders = map[string]string{
	"newest":    "published_at desc",
	"updated":   "updated_at desc",
	"rent_asc":  "rent asc",
	"rent_desc": "rent desc",
	"size_asc":  "area_ping asc",
	"size_desc": "area_ping desc",
}

// Capturing one anchor at a time bounds each field regex below to a single
// entry, so a card that omits an optional field such as the layout or the tags
// cannot absorb the next card's.
var (
	chrbEntryRe   = regexp.MustCompile(`(?s)<a class="property-link[^>]*href="/rental_properties/([^"]+)"[^>]*>(.*?)</a>`)
	chrbTitleRe   = regexp.MustCompile(`(?s)<h2[^>]*>(.*?)</h2>`)
	chrbTagAreaRe = regexp.MustCompile(`(?s)</h2>(.*?)<img src="/images/icons/`)
	chrbTagRe     = regexp.MustCompile(`(?s)<span[^>]*>([^<]*)</span>`)
	chrbKindRe    = regexp.MustCompile(`(?s)icons/house\.svg[^>]*>\s*<span>([^<]*)</span>(?:\s*<span[^>]*>｜</span>\s*<span>([^<]*)</span>)?`)
	chrbSizeRe    = regexp.MustCompile(`(?s)icons/fullscreen\.svg[^>]*>\s*([^<]*?)\s*</span>`)
	chrbAgeRe     = regexp.MustCompile(`(?s)icons/age\.svg[^>]*>\s*([^<]*?)\s*</span>`)
	chrbFloorRe   = regexp.MustCompile(`(?s)icons/layers\.svg[^>]*>\s*([^<]*?)\s*</span>`)
	chrbAddressRe = regexp.MustCompile(`(?s)icons/map-pin\.svg[^>]*>\s*([^<]*?)\s*</div>`)
	chrbRentRe    = regexp.MustCompile(`(?s)<span[^>]*>([^<]*)</span>\s*<span[^>]*>/月</span>`)
	chrbCreatedRe = regexp.MustCompile(`(?s)>\s*(\d{4}-\d{2}-\d{2})\s*</div>`)
)

// ChrbParser is a parser for 大管家房屋網 (https://chrb.com.tw/rental_properties).
type ChrbParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser ChrbParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	link := chrbBaseURL + "/rental_properties"

	requestURL := link
	if sort := query.Get("sort"); sort != "" {
		order, ok := chrbSortOrders[sort]
		if !ok {
			err = &feedgen.ParameterValueInvalidError{Parameter: "sort"}
			return
		}

		requestURL = fmt.Sprintf("%s?%s", link, url.Values{"q[s]": {order}}.Encode())
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status %d from %s", resp.StatusCode, requestURL)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	entries := chrbEntryRe.FindAllStringSubmatch(chrbSearchResults(string(body)), -1)
	if len(entries) == 0 {
		err = &feedgen.ItemFetchError{SourceURL: link}
		return
	}

	feed = &feeds.Feed{
		Title:   "大管家房屋網",
		Link:    &feeds.Link{Href: link},
		Items:   make([]*feeds.Item, 0, len(entries)),
		Created: time.Now(),
	}

	for _, entry := range entries {
		created, dated := chrbCreated(entry[2])
		if !dated {
			continue
		}

		itemLink := fmt.Sprintf("%s/%s", link, entry[1])

		feed.Add(&feeds.Item{
			Id:          itemLink,
			Title:       field(chrbTitleRe, entry[2]),
			Link:        &feeds.Link{Href: itemLink},
			Description: chrbDescription(entry[2]),
			Created:     created,
		})
	}

	return
}

// chrbSearchResults narrows the page to the frame holding the search results,
// leaving the 精選房源 sidebar outside. An absent frame yields an empty string,
// which the caller reports as a fetch failure.
func chrbSearchResults(body string) string {
	start := strings.Index(body, `id="search_results"`)
	if start < 0 {
		return ""
	}

	end := strings.Index(body[start:], "</turbo-frame>")
	if end < 0 {
		return ""
	}

	return body[start : start+end]
}

func chrbCreated(entry string) (created time.Time, ok bool) {
	match := chrbCreatedRe.FindStringSubmatch(entry)
	if match == nil {
		return
	}

	created, err := time.ParseInLocation("2006-01-02", match[1], zone)
	if err != nil {
		return time.Time{}, false
	}

	return created, true
}

// chrbTags reads the pills between the title and the first field icon. Their
// only class is a Tailwind utility, so the region they sit in identifies them.
func chrbTags(entry string) []string {
	area := chrbTagAreaRe.FindStringSubmatch(entry)
	if area == nil {
		return nil
	}

	var tags []string
	for _, match := range chrbTagRe.FindAllStringSubmatch(area[1], -1) {
		if tag := text(match[1]); tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}

func chrbDescription(entry string) string {
	kind := ""
	if match := chrbKindRe.FindStringSubmatch(entry); match != nil {
		kind = text(match[1])
		if layout := text(match[2]); layout != "" {
			kind = fmt.Sprintf("%s %s", kind, layout)
		}
	}

	fields := []struct {
		label string
		value string
	}{
		{"地址", field(chrbAddressRe, entry)},
		{"類型", kind},
		{"坪數", field(chrbSizeRe, entry)},
		{"樓層", field(chrbFloorRe, entry)},
		{"屋齡", field(chrbAgeRe, entry)},
		{"租金", field(chrbRentRe, entry)},
		{"標籤", strings.Join(chrbTags(entry), "、")},
	}

	var lines []string
	for _, field := range fields {
		if field.value == "" {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s: %s", field.label, field.value))
	}

	return strings.Join(lines, "<br>")
}
