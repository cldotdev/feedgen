package feedgen

import (
	"errors"
	"fmt"
	"testing"
)

func TestSkipReasonOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want SkipReason
	}{
		{"not found", PageContentNotFoundError{SourceURL: "u"}, SkipReasonNotFound},
		{"not found pointer", &PageContentNotFoundError{SourceURL: "u"}, SkipReasonNotFound},
		{"status", UnexpectedStatusError{SourceURL: "u", StatusCode: 500}, SkipReasonHTTPStatus},
		{"status pointer", &UnexpectedStatusError{SourceURL: "u", StatusCode: 500}, SkipReasonHTTPStatus},
		{"wrapped status", fmt.Errorf("wrap: %w", UnexpectedStatusError{SourceURL: "u", StatusCode: 404}), SkipReasonHTTPStatus},
		{"plain", errors.New("connection reset"), SkipReasonNetwork},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SkipReasonOf(tt.err); got != tt.want {
				t.Errorf("SkipReasonOf() = %q, want %q", got, tt.want)
			}
		})
	}
}
