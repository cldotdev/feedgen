package site

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

type UdnGameData struct {
	Articles []UdnGameArticles `json:"lists"`
}

type UdnGameArticles struct {
	Title     string               `json:"title"`
	Author    UdnGameArticleAuthor `json:"author"`
	Paragraph string               `json:"paragraph"`
	Url       string               `json:"url"`
	Time      UdnGameArticleTime   `json:"time"`
}

type UdnGameArticleAuthor struct {
	Title string `json:"title"`
}

type UdnGameArticleTime struct {
	DateTime  string `json:"dateTime"`
	Date      string `json:"date"`
	Timestamp int64  `json:"timestamp"`
}

// UdnGameParser is a parser for 遊戲角落 (https://game.udn.com/rank/newest/2003).
type UdnGameParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser UdnGameParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	now := time.Now()

	var sourceLink, rawLink string
	var title, subTitle string

	section := query.Get("section")
	switch section {
	case "rank":
		by := query.Get("by")
		switch by {
		case "newest":
			sourceLink = "https://game.udn.com/rank/newest/2003"
			rawLink = "https://game.udn.com/game/load/article/newest/?limit=20&time=&fl=author,view,photo,cate,hash"
			subTitle = "最新文章"
		case "pv":
			sourceLink = "https://game.udn.com/rank/pv/2003"
			rawLink = "https://game.udn.com/game/load/article/trend/?limit=20&time=7in30&fl=author,view,photo,cate,hash"
			subTitle = "最多瀏覽"
		default:
			err = &feedgen.ParameterValueInvalidError{"by"}
		}
	default:
		err = &feedgen.ParameterValueInvalidError{"section"}
	}

	title = fmt.Sprintf("%s | 遊戲角落", subTitle)

	feed = &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: sourceLink},
		Description: "",
		Author:      nil,
		Created:     now,
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", rawLink, nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	data := UdnGameData{}
	err = sonic.Unmarshal(body, &data)
	if err != nil {
		return
	}

	for _, article := range data.Articles {
		var created time.Time

		switch feedgen.CountDigits(article.Time.Timestamp) {
		case 10:
			created = time.Unix(article.Time.Timestamp, 0)
		case 13:
			created = time.UnixMilli(article.Time.Timestamp)
		default:
			err = &feedgen.ItemFetchError{rawLink}
			return
		}

		feedItem := &feeds.Item{
			Id:          article.Url,
			Title:       article.Title,
			Link:        &feeds.Link{Href: article.Url},
			Description: article.Paragraph,
			Author:      &feeds.Author{Name: article.Author.Title},
			Created:     created,
		}

		feed.Add(feedItem)
	}

	return
}
