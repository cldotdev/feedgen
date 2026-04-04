package site

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

// GamerForumParser is a parser for Gamer Forum (https://forum.gamer.com.tw/).
type GamerForumParser struct{}

type GamerForumRespBody struct {
	Data GamerForumRespData `json:"data"`
}

type GamerForumRespData struct {
	OtherInfo GamerForumRespDataOtherInfo `json:"otherInfo"`
	List      []GamerForumRespDataList    `json:"list"`
}

type GamerForumRespDataOtherInfo struct {
	Subscribed   bool   `json:"subscribed"`
	BoardName    string `json:"boardName"`
	BoardImage   string `json:"board_image"`
	BoardSummary string `json:"board_summary"`
}

type GamerForumRespDataList struct {
	Bsn            int    `json:"bsn"`
	SnA            int    `json:"snA"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Author         string `json:"author"`
	Nickname       string `json:"nickname"`
	Ctime          string `json:"ctime"`
	ReplyTimestamp int64  `json:"reply_timestamp"`
	SubboardTitle  string `json:"subboard_title"`
	Gp             int64  `json:"gp"`
	Locked         bool   `json:"locked"`
	Del            bool   `json:"del"`
}

// GetFeed returns generated feed with the given query parameters.
func (parser GamerForumParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	bsn := query.Get("bsn")
	if bsn == "" {
		err = &feedgen.ParameterNotFoundError{"bsn"}
		return
	}

	if matched, _ := regexp.MatchString(`^\d+$`, bsn); !matched {
		err = &feedgen.ParameterValueInvalidError{"bsn"}
		return
	}

	gp := query.Get("gp")
	if gp != "" {
		if matched, _ := regexp.MatchString(`^(5|20|50|100|200)$`, gp); !matched {
			err = &feedgen.ParameterValueInvalidError{"gp"}
			return
		}
	}

	now := time.Now()

	// The parameters for https://api.gamer.com.tw:
	//   - keyword
	//   - gpnum (with type=3)
	url := fmt.Sprintf("https://api.gamer.com.tw/mobile_app/forum/v3/B.php?bsn=%s&order=post&page=1", bsn)

	if gp != "" {
		url = fmt.Sprintf("%s&type=3&gpnum=%s", url, gp)
	} else {
		url = fmt.Sprintf("%s&type=1", url)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:106.0) Gecko/20100101 Firefox/106.0")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	respBody := new(GamerForumRespBody)
	err = json.Unmarshal(bodyBytes, respBody)
	if err != nil {
		return
	}

	respData := respBody.Data
	title := fmt.Sprintf("%s 哈啦板 - 巴哈姆特", respData.OtherInfo.BoardName)
	description := respData.OtherInfo.BoardSummary

	feed = &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: url},
		Description: description,
		Author:      nil,
		Created:     now,
	}

	if len(respData.List) == 0 {
		err = &feedgen.ItemFetchError{url}
		return
	}

	for _, list := range respData.List {
		itemID := fmt.Sprintf(
			"https://forum.gamer.com.tw/C.php?bsn=%d&snA=%d",
			list.Bsn,
			list.SnA,
		)
		itemTitle := list.Title
		itemLink := itemID
		itemContent := list.Summary
		itemAuthor := fmt.Sprintf("%s (%s)", list.Author, list.Nickname)

		feed.Add(&feeds.Item{
			Id:          itemID,
			Title:       itemTitle,
			Link:        &feeds.Link{Href: itemLink},
			Description: itemContent,
			Author:      &feeds.Author{Name: itemAuthor},
			Created:     now,
		})
	}

	return
}
