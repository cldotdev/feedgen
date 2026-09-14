package site

import (
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 30 * time.Second}

// time.LoadLocation is avoided here because the runtime image may ship
// without tzdata.
var zone = time.FixedZone("CST", 8*60*60)

func field(re *regexp.Regexp, entry string) string {
	match := re.FindStringSubmatch(entry)
	if match == nil {
		return ""
	}

	return text(match[1])
}

func text(raw string) string {
	return strings.TrimSpace(html.UnescapeString(raw))
}
