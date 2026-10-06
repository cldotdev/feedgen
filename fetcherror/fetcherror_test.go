package fetcherror

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cldotdev/feedgen"
)

type row struct {
	sourceURL, message, firstSeen, lastSeen string
	count                                   int
}

func openTest(t *testing.T, path string) *Recorder {
	t.Helper()
	r, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	return r
}

func getRow(t *testing.T, r *Recorder, parser, itemURL string, reason feedgen.SkipReason) row {
	t.Helper()
	var got row
	err := r.db.QueryRow(`SELECT source_url, message, first_seen, last_seen, count
		FROM skipped_items WHERE parser = ? AND item_url = ? AND reason = ?`,
		parser, itemURL, string(reason)).
		Scan(&got.sourceURL, &got.message, &got.firstSeen, &got.lastSeen, &got.count)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func countRows(t *testing.T, r *Recorder) int {
	t.Helper()
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM skipped_items`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRecordSkip(t *testing.T) {
	r := openTest(t, filepath.Join(t.TempDir(), "fetch_error.db"))
	t1 := time.Date(2026, 10, 6, 1, 2, 3, 0, time.UTC)
	t2 := t1.Add(time.Hour)
	item := feedgen.SkippedItem{
		Parser: "ptt", SourceURL: "https://example.com/list", ItemURL: "https://example.com/a",
		Reason: feedgen.SkipReasonNetwork, Message: "first",
	}

	r.now = func() time.Time { return t1 }
	if err := r.RecordSkip(item); err != nil {
		t.Fatal(err)
	}
	got := getRow(t, r, item.Parser, item.ItemURL, item.Reason)
	if got.count != 1 || got.firstSeen != "2026-10-06T01:02:03Z" || got.lastSeen != got.firstSeen {
		t.Errorf("first write = %+v", got)
	}

	r.now = func() time.Time { return t2 }
	item.Message = "second"
	item.SourceURL = "https://example.com/list2"
	if err := r.RecordSkip(item); err != nil {
		t.Fatal(err)
	}
	got = getRow(t, r, item.Parser, item.ItemURL, item.Reason)
	want := row{sourceURL: item.SourceURL, message: "second", firstSeen: "2026-10-06T01:02:03Z", lastSeen: "2026-10-06T02:02:03Z", count: 2}
	if got != want {
		t.Errorf("repeat write = %+v, want %+v", got, want)
	}

	item.Reason = feedgen.SkipReasonHTTPStatus
	if err := r.RecordSkip(item); err != nil {
		t.Fatal(err)
	}
	if n := countRows(t, r); n != 2 {
		t.Errorf("rows = %d, want 2", n)
	}
}

func TestOpenKeepsRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fetch_error.db")
	r, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	item := feedgen.SkippedItem{Parser: "chrb", ItemURL: "https://example.com/a", Reason: feedgen.SkipReasonNoDate}
	if err := r.RecordSkip(item); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	r = openTest(t, path)
	if n := countRows(t, r); n != 1 {
		t.Errorf("rows after reopen = %d, want 1", n)
	}
}
