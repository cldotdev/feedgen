package site

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

// HackernewsParser is a parser for Hacker News (https://news.ycombinator.com/).
type HackernewsParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser HackernewsParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	category := query.Get("category")
	switch category {
	case "best":
		feed, err = parser.GetBestFeed()
	case "":
		err = &feedgen.ParameterNotFoundError{"category"}
	default:
		err = &feedgen.ParameterValueInvalidError{"category"}
		return
	}

	return
}

// GetBestFeed returns generated feed with best category.
func (parser HackernewsParser) GetBestFeed() (feed *feeds.Feed, err error) {
	now := time.Now()
	title := "Top Links | Hacker News"
	url := "https://news.ycombinator.com/best"
	feed = &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: url},
		Description: "",
		Author:      nil,
		Created:     now,
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	re := regexp.MustCompile(`(?s)<td class="title"><span class="titleline"><a href="(.+?)".*?>(.+?)</a>.+?<span class="score" id=".+?">(\d+?) points</span>.+?by <a href=".+?" class="hnuser">(.+?)</a>.+?(\d+?) (days?|hours?|minutes?) ago.+?<a href="(.+?)">(\d+?)&nbsp;comments?</a>`)
	matchGroup := re.FindAllSubmatch(body, -1)
	if len(matchGroup) == 0 {
		err = &feedgen.ItemFetchError{url}
		return
	}

	for _, m := range matchGroup {
		var itemPointDesc string
		var itemCommentDesc string

		itemLink := string(m[1])
		itemTitle := string(m[2])
		itemPoint := string(m[3])
		itemAuthor := string(m[4])
		itemBeforeTime := string(m[5])
		itemBeforeTimeUnit := string(m[6])
		itemCommentPath := string(m[7])
		itemCommentCount := string(m[8])
		itemCommentLink := fmt.Sprintf("https://news.ycombinator.com/%s", itemCommentPath)

		if itemPoint == "1" {
			itemPointDesc = fmt.Sprintf("%s point", itemPoint)
		} else {
			itemPointDesc = fmt.Sprintf("%s points", itemPoint)
		}

		if itemCommentCount == "1" {
			itemCommentDesc = fmt.Sprintf("%s comment", itemCommentCount)
		} else {
			itemCommentDesc = fmt.Sprintf("%s comments", itemCommentCount)
		}
		itemDescription := fmt.Sprintf("%s. <a href=\"%s\" >%s</a>", itemPointDesc, itemCommentLink, itemCommentDesc)
		created := now

		if strings.Contains(itemBeforeTimeUnit, "day") {
			day, _ := strconv.Atoi(itemBeforeTime)
			created = now.AddDate(0, 0, -day)
		} else if strings.Contains(itemBeforeTimeUnit, "hour") {
			duration, _ := time.ParseDuration(fmt.Sprintf("-%sh", itemBeforeTime))
			created = now.Add(duration)
		} else if strings.Contains(itemBeforeTimeUnit, "minute") {
			duration, _ := time.ParseDuration(fmt.Sprintf("-%sm", itemBeforeTime))
			created = now.Add(duration)
		}

		feed.Add(&feeds.Item{
			Id:          itemCommentLink,
			Title:       string(itemTitle),
			Link:        &feeds.Link{Href: itemLink},
			Description: itemDescription,
			Author:      &feeds.Author{Name: itemAuthor},
			Created:     created,
		})
	}

	return
}
