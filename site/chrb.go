package site

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/feeds"

	"github.com/cldotdev/feedgen"
)

// ChrbParser is a parser for 大管家房屋網 (https://www.chrb.com.tw/tenement.php).
type ChrbParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser ChrbParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	now := time.Now()

	var link = "https://www.chrb.com.tw/tenement_ajax.php"

	resp, err := http.PostForm(link, query)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	title := "大管家房屋網"

	feed = &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: link},
		Description: "",
		Author:      nil,
		Created:     now,
	}

	baseLink := "https://www.chrb.com.tw"
	re := regexp.MustCompile(`(?s)<a href="(\d+.html)".+?<span class="tenement_04">(.+?)</span>.+?<span class="tenement_06">(.+?)</span>.+?<span class="tenement_07">(用途：.+?)\&nbsp;.+?(總樓層：.+?)</span>.+?<!--<span class="tenement_07">(.+?)</span><br /> edit by.+?<td class="price" align="center" noWrap>(.+?)</td>.+?<td class="tenement_07" align="center" noWrap>(.+?)</td>.+?<td class="tenement_07" align="center" noWrap>(.+?)</td>.+?<td class="tenement_07" align="center" noWrap>(.+?)</td>`)
	matchGroup := re.FindAllStringSubmatch(string(body), -1)
	if len(matchGroup) == 0 {
		err = &feedgen.ItemFetchError{link}
		return
	}

	for _, m := range matchGroup {
		itemLink := fmt.Sprintf("%s/%s", baseLink, m[1])
		itemTitle := m[2]
		itemAuthor := ""
		address := m[3]    // 地址
		purpose := m[4]    // 用途
		floor := m[5]      // 總樓層
		editedDate := m[6] // 更新日期
		price := m[7]      // 租金
		size := m[8]       // 坪數
		layout := m[9]     // 格局
		age := m[10]       // 屋齡
		itemDescription := fmt.Sprintf(
			"地址: %s<br>用途: %s<br>總樓層: %s<br>租金: %s<br>坪數: %s<br>格局: %s<br>屋齡: %s<br>更新日期: %s",
			address, purpose, floor, price, size, layout, age, editedDate,
		)

		created, _ := time.Parse("2006-01-02", editedDate)

		feed.Add(&feeds.Item{
			Id:          itemLink,
			Title:       itemTitle,
			Link:        &feeds.Link{Href: itemLink},
			Description: itemDescription,
			Author:      &feeds.Author{Name: itemAuthor},
			Created:     created,
		})
	}

	return
}
