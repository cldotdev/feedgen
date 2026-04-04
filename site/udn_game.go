package site

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
	"github.com/cldotdev/feedgen/parser"
)

// UdnGameParser is a parser for 遊戲角落 (https://game.udn.com/rank/newest/2003).
type UdnGameParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (p UdnGameParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	var sourceLink, rawLink string
	var subTitle string

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
			err = &feedgen.ParameterValueInvalidError{Parameter: "by"}
			return
		}
	default:
		err = &feedgen.ParameterValueInvalidError{Parameter: "section"}
		return
	}

	articles, err := parser.FetchArticles(rawLink)
	if err != nil {
		return
	}

	feed = &feeds.Feed{
		Title:   fmt.Sprintf("%s | 遊戲角落", subTitle),
		Link:    &feeds.Link{Href: sourceLink},
		Created: time.Now(),
	}

	parser.AddArticlesToFeed(feed, articles)

	return
}
