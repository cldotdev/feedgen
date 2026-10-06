// Package fetcherror stores the items that parsers skip in a SQLite database.
package fetcherror

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "modernc.org/sqlite"

	"github.com/cldotdev/feedgen"
)

const schema = `CREATE TABLE IF NOT EXISTS skipped_items (
	parser     TEXT    NOT NULL,
	item_url   TEXT    NOT NULL,
	reason     TEXT    NOT NULL,
	source_url TEXT    NOT NULL,
	message    TEXT    NOT NULL,
	first_seen TEXT    NOT NULL,
	last_seen  TEXT    NOT NULL,
	count      INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY (parser, item_url, reason)
)`

const upsert = `INSERT INTO skipped_items
	(parser, item_url, reason, source_url, message, first_seen, last_seen)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (parser, item_url, reason) DO UPDATE SET
	source_url = excluded.source_url,
	message    = excluded.message,
	last_seen  = excluded.last_seen,
	count      = count + 1`

// Recorder is a feedgen.SkipRecorder backed by a SQLite database.
type Recorder struct {
	db  *sql.DB
	now func() time.Time
}

// Open opens the database at path, creating the file and table when missing.
func Open(path string) (*Recorder, error) {
	q := url.Values{}
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "busy_timeout(5000)")

	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, err
	}

	// SQLite admits one writer at a time; one connection serializes the
	// concurrent requests inside this process.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create skipped_items: %w", err)
	}

	return &Recorder{db: db, now: time.Now}, nil
}

// RecordSkip inserts the item, or updates the existing record of the same
// parser, item URL, and reason.
func (r *Recorder) RecordSkip(item feedgen.SkippedItem) error {
	ts := r.now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(upsert, item.Parser, item.ItemURL, string(item.Reason),
		item.SourceURL, item.Message, ts, ts)
	return err
}

// Close closes the database.
func (r *Recorder) Close() error {
	return r.db.Close()
}
