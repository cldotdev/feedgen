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

// HackmdData represents the response from HackMD API
type HackmdData struct {
	Notes []HackmdNote `json:"notes"`
}

// HackmdNote represents a single note/article from HackMD
type HackmdNote struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	ShortID     string    `json:"shortId"`
	Username    string    `json:"username"`
	Userpath    string    `json:"userpath"`
	PublishedAt time.Time `json:"publishedAt"`
}

// HackmdParser is a parser for HackMD (https://hackmd.io/).
type HackmdParser struct{}

// GetFeed returns generated feed with the given query parameters.
func (parser HackmdParser) GetFeed(query feedgen.QueryValues) (feed *feeds.Feed, err error) {
	now := time.Now()

	// Get username from query parameter 'u'
	username := query.Get("u")
	if username == "" {
		err = &feedgen.ParameterNotFoundError{"u"}
		return
	}

	// Construct API URL
	apiURL := fmt.Sprintf("https://hackmd.io/api/%s/notes?limit=12&page=1&keyword=&sortBy=publishedAt&order=desc", username)
	sourceLink := fmt.Sprintf("https://hackmd.io/%s", username)

	feed = &feeds.Feed{
		Title:       fmt.Sprintf("%s | HackMD", username),
		Link:        &feeds.Link{Href: sourceLink},
		Description: fmt.Sprintf("Latest published notes from %s on HackMD", username),
		Author:      &feeds.Author{Name: username},
		Created:     now,
	}

	// Fetch data from HackMD API
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Parse JSON response
	data := HackmdData{}
	err = sonic.Unmarshal(body, &data)
	if err != nil {
		return
	}

	// Convert notes to feed items
	for _, note := range data.Notes {
		// Construct note URL
		noteURL := fmt.Sprintf("https://hackmd.io/%s/%s", username, note.ShortID)

		feedItem := &feeds.Item{
			Id:          noteURL,
			Title:       note.Title,
			Link:        &feeds.Link{Href: noteURL},
			Description: note.Content,
			Author:      &feeds.Author{Name: note.Username},
			Created:     note.PublishedAt,
		}

		feed.Add(feedItem)
	}

	return
}
