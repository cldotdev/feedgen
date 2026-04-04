package site

import (
	"fmt"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
	"github.com/cldotdev/feedgen/parser"
)

type globalVisionTag struct {
	name      string
	feedTitle string
}

var globalVisionTags = map[string]globalVisionTag{
	"in-depth-column": {
		name:      "深度專欄",
		feedTitle: "深度專欄 | 轉角國際",
	},
}

// UdnGlobalVisionParser is a parser for 轉角國際 (https://global.udn.com/global_vision/).
type UdnGlobalVisionParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (p UdnGlobalVisionParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	tagSlug := query.Get("tag")
	if tagSlug == "" {
		err = &feedgen.ParameterNotFoundError{Parameter: "tag"}
		return
	}

	tag, ok := globalVisionTags[tagSlug]
	if !ok {
		err = &feedgen.ParameterValueInvalidError{Parameter: "tag"}
		return
	}

	rawLink := fmt.Sprintf("https://global.udn.com/global_vision/load/article/newest/tag:%s", tag.name)

	articles, err := parser.FetchArticles(rawLink)
	if err != nil {
		return
	}

	feed = &feeds.Feed{
		Title:   tag.feedTitle,
		Link:    &feeds.Link{Href: fmt.Sprintf("https://global.udn.com/global_vision/newest/tag/%s", tag.name)},
		Created: time.Now(),
	}

	parser.AddArticlesToFeed(feed, articles)

	return
}
