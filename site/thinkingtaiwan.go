package site

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

const (
	thinkingtaiwanBaseURL = "https://www.thinkingtaiwan.net"

	thinkingtaiwanFeedTitle = "最新文章 | 想想論壇"

	// Bounds the load placed on the site: each feed request fetches one page
	// per listing row.
	thinkingtaiwanMaxInFlight = 2
)

// Rows are split on the row marker before any field is matched, so a row that
// omits an optional field cannot absorb the next row's.
const thinkingtaiwanRowMarker = `<div class="entity-row-inner views-row-inner">`

// The listing mixes /article/N and /content/N hrefs. The date pattern stays
// inside the post-date element, so a page without one is skipped instead of
// matching a later date elsewhere on the page.
var (
	thinkingtaiwanTitleRe  = regexp.MustCompile(`(?s)<h2><a href="(/(?:article|content)/\d+)"[^>]*>(.*?)</a></h2>`)
	thinkingtaiwanTextRe   = regexp.MustCompile(`(?s)<div class="field-body"><a [^>]*><p>(.*?)</p>`)
	thinkingtaiwanDateRe   = regexp.MustCompile(`class="post-date">\s*<span>[^<]*(?:<span[^>]*>[^<]*</span>)?\s*(\d{4})-(\d{2})-(\d{2})`)
	thinkingtaiwanWriterRe = regexp.MustCompile(`(?s)field--name-field-writer[^"]*">(.*?)</div>`)
)

// ThinkingtaiwanParser is a parser for 想想論壇 (https://www.thinkingtaiwan.net/).
type ThinkingtaiwanParser struct{}

// GetFeed returns generated feed.
func (parser ThinkingtaiwanParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	link := thinkingtaiwanBaseURL + "/articles"

	body, err := thinkingtaiwanGet(link)
	if err != nil {
		return
	}

	rows := strings.Split(body, thinkingtaiwanRowMarker)[1:]

	items := make([]*feeds.Item, 0, len(rows))
	for _, row := range rows {
		match := thinkingtaiwanTitleRe.FindStringSubmatch(row)
		if match == nil {
			continue
		}

		itemLink := thinkingtaiwanBaseURL + match[1]
		items = append(items, &feeds.Item{
			Id:          itemLink,
			Title:       text(match[2]),
			Link:        &feeds.Link{Href: itemLink},
			Description: field(thinkingtaiwanTextRe, row),
		})
	}

	if len(items) == 0 {
		err = &feedgen.ItemFetchError{SourceURL: link}
		return
	}

	// Per-index slots keep the listing order without sorting.
	complete := make([]bool, len(items))
	sem := make(chan struct{}, thinkingtaiwanMaxInFlight)

	var wg sync.WaitGroup
	for i, item := range items {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			complete[i] = thinkingtaiwanFillArticle(item)
		}()
	}
	wg.Wait()

	feed = &feeds.Feed{
		Title:   thinkingtaiwanFeedTitle,
		Link:    &feeds.Link{Href: link},
		Items:   make([]*feeds.Item, 0, len(items)),
		Created: time.Now(),
	}

	for i, item := range items {
		if complete[i] {
			feed.Add(item)
		}
	}

	if len(feed.Items) == 0 {
		feed = nil
		err = &feedgen.ItemFetchError{SourceURL: link}
	}

	return
}

// thinkingtaiwanFillArticle sets the date and author from the article page and
// reports whether the article has a usable date.
func thinkingtaiwanFillArticle(item *feeds.Item) bool {
	body, err := thinkingtaiwanGet(item.Link.Href)
	if err != nil {
		return false
	}

	match := thinkingtaiwanDateRe.FindStringSubmatch(body)
	if match == nil {
		return false
	}

	year, _ := strconv.Atoi(match[1])
	month, _ := strconv.Atoi(match[2])
	day, _ := strconv.Atoi(match[3])

	item.Created = time.Date(year, time.Month(month), day, 0, 0, 0, 0, zone)
	item.Author = &feeds.Author{Name: field(thinkingtaiwanWriterRe, body)}

	return true
}

func thinkingtaiwanGet(url string) (body string, err error) {
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
		return
	}

	raw, err := io.ReadAll(resp.Body)
	return string(raw), err
}
