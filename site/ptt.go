package site

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

// pttTransport uses classic key exchange curves only.
// Go 1.26+ sends post-quantum MLKEM key shares by default, which enlarges
// the TLS ClientHello beyond what some network middleboxes can handle,
// causing "connection reset by peer" errors on certain hosting platforms.
var pttTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
	},
}

var pttClient = &http.Client{Transport: pttTransport, Timeout: 30 * time.Second}

// PttParser is a parser for PTT Web (https://www.ptt.cc/bbs/index.html).
type PttParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser PttParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	now := time.Now()
	boardName := query.Get("b")
	if boardName == "" {
		err = &feedgen.ParameterNotFoundError{"b"}
		return
	}

	q := query.Get("q")
	title := fmt.Sprintf("批踢踢實業坊 %s 板", boardName)
	var url string
	if q == "" {
		url = "https://www.ptt.cc/bbs/" + boardName + "/index.html"
	} else {
		url = "https://www.ptt.cc/bbs/" + boardName + "/search?q=" + neturl.QueryEscape(q)
	}

	feed = &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: url},
		Description: "",
		Author:      nil,
		Created:     now,
	}

	client := pttClient
	cookie := http.Cookie{Name: "over18", Value: "1"}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.AddCookie(&cookie)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// 排除置底文章
	re := regexp.MustCompile(`(?s)(.+?)<div class="r-list-sep"></div>`)
	match := re.Find(body)
	if match == nil {
		match = body
	}

	var feedItems []*feeds.Item

	re = regexp.MustCompile(fmt.Sprintf(`<a href="(/bbs/%s/M\..+?\.html)">`, boardName))
	matchGroup := re.FindAllSubmatch(match, -1)
	feedItemsCount := len(matchGroup)
	if feedItemsCount == 0 {
		return
	}

	for _, m := range matchGroup {
		url := "https://www.ptt.cc" + string(m[1])

		var feedItem *feeds.Item
		feedItem, err = parser.GetFeedItem(url)
		if err != nil {
			continue
		}

		feedItems = append(feedItems, feedItem)
	}

	feedgen.SortFeedItemsLatestFirst(feedItems)

	for _, feedItem := range feedItems {
		feed.Add(feedItem)
	}

	return
}

// GetFeedItem returns feed item generated from item URL.
func (parser PttParser) GetFeedItem(url string) (feedItem *feeds.Item, err error) {
	re := regexp.MustCompile(`(?s)<div id="main-content" class="bbs-screen bbs-content"><div class="article-metaline"><span class="article-meta-tag">作者</span><span class="article-meta-value">(.+?)</span></div>(<div class="article-metaline-right"><span class="article-meta-tag">看板</span><span class="article-meta-value">(.+?)</span></div>)?<div class="article-metaline"><span class="article-meta-tag">標題</span><span class="article-meta-value">(.+?)</span></div>(<div class="article-metaline"><span class="article-meta-tag">時間</span><span class="article-meta-value">(.+?)</span></div>)?(.+?)<span class="f2">※ (發信站|編輯)`)
	re2 := regexp.MustCompile(`(?s)class="bbs-screen bbs-content">(.+?)<span class="f2">※ (發信站|編輯)`)
	re3 := regexp.MustCompile(`404 - Not Found`)
	client := pttClient
	cookie := http.Cookie{Name: "over18", Value: "1"}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.AddCookie(&cookie)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	author := "null"
	board := "null"
	title := "null"
	created := time.Now()
	description := "null"

	match := re.FindSubmatch(body)
	if match == nil {
		match = re2.FindSubmatch(body)
		if match == nil {
			match = re3.FindSubmatch(body)
			if match == nil {
				// err = &feedgen.PageContentFetchError{url}
				description = "<pre>" + string(body) + "</pre>"
			} else {
				err = &feedgen.PageContentNotFoundError{url}
				return
			}

		} else {
			description = "<pre>" + string(match[1]) + "</pre>"
		}
	} else {
		author = string(match[1])
		board = string(match[3])
		title = string(match[4])
		const timeForm = "Mon Jan 2 15:04:05 2006"
		date := string(match[6])
		created, _ = time.Parse(timeForm, date)
		loc, _ := time.LoadLocation("Asia/Taipei")
		created = created.In(loc)
		content := string(match[7])
		content = regexp.MustCompile(`(?s)<div class="richcontent"><blockquote.+?</script></div>`).ReplaceAllString(content, "")
		content = regexp.MustCompile(`(?s)<div class="richcontent"><div class="resize-container"><div class="resize-content"><iframe.+</iframe></div></div></div>`).ReplaceAllString(content, "")
		content = regexp.MustCompile(`(?s)<div class="richcontent"><img src=".+?" alt="" /></div>`).ReplaceAllString(content, "")
		content = regexp.MustCompile(`(?s)<div.+?>(.+?)</div>`).ReplaceAllString(content, "$1")
		content = regexp.MustCompile(`(?s)<span.+?>(.+?)</span>`).ReplaceAllString(content, "$1")
		content = regexp.MustCompile(`(?s)<a.+?>(.+?)</a>`).ReplaceAllString(content, "$1")
		description = "<pre>"
		if board != "" {
			description += "看板：" + board + "\n"
		}
		description += "作者：" + author + "\n" + "標題：" + title + "\n"
		if date != "" {
			description += "時間：" + date + "\n"
		}
		description += content + "</pre>"
	}

	feedItem = &feeds.Item{
		Id:          url,
		Title:       title,
		Link:        &feeds.Link{Href: url},
		Description: description,
		Author:      &feeds.Author{Name: author},
		Created:     created,
	}

	return
}
