package feedgen

import (
	"errors"
	"log"
)

// SkipReason tells why a parser left an item out of a feed.
type SkipReason string

// The fixed set of skip reasons.
const (
	SkipReasonNetwork    SkipReason = "network"
	SkipReasonHTTPStatus SkipReason = "http_status"
	SkipReasonNotFound   SkipReason = "not_found"
	SkipReasonNoDate     SkipReason = "no_date"
)

// SkippedItem describes one item a parser left out of a feed.
type SkippedItem struct {
	Parser    string
	SourceURL string
	ItemURL   string
	Reason    SkipReason
	Message   string
}

// SkipRecorder persists skipped items.
type SkipRecorder interface {
	RecordSkip(SkippedItem) error
}

type logSkipRecorder struct{}

func (logSkipRecorder) RecordSkip(item SkippedItem) error {
	log.Printf("skipped item: parser=%s reason=%s item=%s source=%s message=%s",
		item.Parser, item.Reason, item.ItemURL, item.SourceURL, item.Message)
	return nil
}

var skipRecorder SkipRecorder = logSkipRecorder{}

// SetSkipRecorder installs the recorder that RecordSkip forwards to.
// It is not safe for concurrent use; call it before serving requests.
func SetSkipRecorder(r SkipRecorder) {
	skipRecorder = r
}

// RecordSkip reports a skipped item to the installed recorder.
// A recorder failure is logged and never returned to the caller.
func RecordSkip(item SkippedItem) {
	if err := skipRecorder.RecordSkip(item); err != nil {
		log.Printf("record skipped item %s: %v", item.ItemURL, err)
	}
}

// SkipReasonOf maps a fetch error to a skip reason.
func SkipReasonOf(err error) SkipReason {
	var notFound PageContentNotFoundError
	var notFoundPtr *PageContentNotFoundError
	var status UnexpectedStatusError
	var statusPtr *UnexpectedStatusError

	switch {
	case errors.As(err, &notFound), errors.As(err, &notFoundPtr):
		return SkipReasonNotFound
	case errors.As(err, &status), errors.As(err, &statusPtr):
		return SkipReasonHTTPStatus
	default:
		return SkipReasonNetwork
	}
}
