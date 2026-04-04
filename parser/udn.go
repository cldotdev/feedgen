package parser

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

// Article represents a parsed UDN article with a resolved timestamp.
type Article struct {
	Title     string
	Author    string
	Paragraph string
	URL       string
	Created   time.Time
}

type udnResponse struct {
	Articles []udnArticle `json:"lists"`
}

type udnArticle struct {
	Title     string          `json:"title"`
	Author    udnArticleField `json:"author"`
	Paragraph string          `json:"paragraph"`
	URL       string          `json:"url"`
	Time      udnArticleTime  `json:"time"`
}

type udnArticleField struct {
	Title string `json:"title"`
}

type udnArticleTime struct {
	DateTime  string `json:"dateTime"`
	Date      string `json:"date"`
	Timestamp int64  `json:"timestamp"`
}

// FetchArticles fetches articles from a UDN JSON API endpoint and returns
// them with timestamps converted to time.Time.
func FetchArticles(rawLink string) ([]Article, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", rawLink, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, rawLink)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data udnResponse
	if err := sonic.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	articles := make([]Article, 0, len(data.Articles))
	for _, a := range data.Articles {
		if a.Time.Timestamp == 0 {
			continue
		}

		var created time.Time
		switch feedgen.CountDigits(a.Time.Timestamp) {
		case 10:
			created = time.Unix(a.Time.Timestamp, 0)
		case 13:
			created = time.UnixMilli(a.Time.Timestamp)
		default:
			return nil, &feedgen.ItemFetchError{SourceURL: rawLink}
		}

		articles = append(articles, Article{
			Title:     a.Title,
			Author:    a.Author.Title,
			Paragraph: a.Paragraph,
			URL:       a.URL,
			Created:   created,
		})
	}

	return articles, nil
}

// AddArticlesToFeed appends articles as feed items to the given feed.
func AddArticlesToFeed(feed *feeds.Feed, articles []Article) {
	for _, a := range articles {
		feed.Add(&feeds.Item{
			Id:          a.URL,
			Title:       a.Title,
			Link:        &feeds.Link{Href: a.URL},
			Description: a.Paragraph,
			Author:      &feeds.Author{Name: a.Author},
			Created:     a.Created,
		})
	}
}
