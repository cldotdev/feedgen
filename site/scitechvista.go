package site

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

const scitechvistaBaseURL = "https://scitechvista.nat.gov.tw"

var scitechvistaFeedTitles = map[string]string{
	"new":      "最新文章 | 科技大觀園",
	"hot":      "熱門文章 | 科技大觀園",
	"featured": "精選文章 | 科技大觀園",
}

var scitechvistaClient = &http.Client{Timeout: 30 * time.Second}

// The same page also renders a 推薦文章 sidebar whose anchors carry
// "kf-item py-3", so anchoring on "kf-item align-items-center" keeps the
// sidebar out of the feed. Capturing one anchor at a time also bounds each
// field regex below to a single entry, so an entry that omits an optional
// field cannot absorb the next entry's.
var (
	scitechvistaEntryRe  = regexp.MustCompile(`(?s)<a href="(/Article/[^"]+)"[^>]*class="kf-item align-items-center">(.*?)</a>`)
	scitechvistaDateRe   = regexp.MustCompile(`(?s)kf-date[^>]*>\s*<span>\s*(\d{2,3})/(\d{2})/(\d{2})\s*</span>`)
	scitechvistaTitleRe  = regexp.MustCompile(`(?s)<div class="kf-title[^"]*">(.*?)</div>`)
	scitechvistaTextRe   = regexp.MustCompile(`(?s)<div class="kf-txt[^"]*">(.*?)</div>`)
	scitechvistaAuthorRe = regexp.MustCompile(`(?s)<span class="text-truncate Author">(.*?)</span>`)
)

// time.LoadLocation is avoided here because the runtime image may ship
// without tzdata.
var scitechvistaZone = time.FixedZone("CST", 8*60*60)

// ScitechvistaParser is a parser for 科技大觀園 (https://scitechvista.nat.gov.tw/).
type ScitechvistaParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser ScitechvistaParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	sectionSlug := query.Get("section")
	if sectionSlug == "" {
		err = &feedgen.ParameterNotFoundError{Parameter: "section"}
		return
	}

	feedTitle, ok := scitechvistaFeedTitles[sectionSlug]
	if !ok {
		err = &feedgen.ParameterValueInvalidError{Parameter: "section"}
		return
	}

	link := fmt.Sprintf("%s/Article/C000003/%s", scitechvistaBaseURL, sectionSlug)

	resp, err := scitechvistaClient.Get(link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status %d from %s", resp.StatusCode, link)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	entries := scitechvistaEntryRe.FindAllStringSubmatch(string(body), -1)
	if len(entries) == 0 {
		err = &feedgen.ItemFetchError{SourceURL: link}
		return
	}

	feed = &feeds.Feed{
		Title:   feedTitle,
		Link:    &feeds.Link{Href: link},
		Items:   make([]*feeds.Item, 0, len(entries)),
		Created: time.Now(),
	}

	for _, entry := range entries {
		created, dated := scitechvistaCreated(entry[2])
		if !dated {
			continue
		}

		itemLink := scitechvistaBaseURL + html.UnescapeString(entry[1])

		feed.Add(&feeds.Item{
			Id:          itemLink,
			Title:       scitechvistaField(scitechvistaTitleRe, entry[2]),
			Link:        &feeds.Link{Href: itemLink},
			Description: scitechvistaField(scitechvistaTextRe, entry[2]),
			Author:      &feeds.Author{Name: scitechvistaField(scitechvistaAuthorRe, entry[2])},
			Created:     created,
		})
	}

	return
}

func scitechvistaField(re *regexp.Regexp, entry string) string {
	match := re.FindStringSubmatch(entry)
	if match == nil {
		return ""
	}

	return strings.TrimSpace(html.UnescapeString(match[1]))
}

func scitechvistaCreated(entry string) (created time.Time, ok bool) {
	match := scitechvistaDateRe.FindStringSubmatch(entry)
	if match == nil {
		return
	}

	rocYear, _ := strconv.Atoi(match[1])
	month, _ := strconv.Atoi(match[2])
	day, _ := strconv.Atoi(match[3])

	return time.Date(rocYear+1911, time.Month(month), day, 0, 0, 0, 0, scitechvistaZone), true
}
