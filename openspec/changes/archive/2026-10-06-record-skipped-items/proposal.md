## Why

Five parsers drop an article silently when its page cannot be fetched or its date cannot be read, so a reader sees a shorter feed and the operator has no trace of what failed. The PTT parser also gets this wrong in the other direction: a failure of the last listed article turns the whole feed into an HTTP 400, while a failure of any other article goes unnoticed. Separately, the 想想論壇 parser fetches two article pages at a time on the assumption that this halves latency, and measurement shows it does not.

## What Changes

- Record every skipped item in a SQLite file, `log/fetch_error.db`, keyed by parser, item URL, and failure reason. A repeat of the same key updates its last-seen time, count, and latest message instead of adding a row.
- Add a recorder hook to the root `feedgen` package that the parsers in `site/` and `parser/` report to, and a separate package holding the SQLite implementation, wired up in `web/main.go`.
- Report skips from `site/ptt.go`, `site/thinkingtaiwan.go`, `site/scitechvista.go`, `site/chrb.go`, and `parser/udn.go`.
- Fix `site/ptt.go` so that a failed article is skipped wherever it sits in the listing, and return `ItemFetchError` when no article succeeds or the page is not a recognizable list.
- Fetch 想想論壇 article pages one at a time, and correct the archived `add-thinkingtaiwan` design, which claims the two-at-a-time fetch halves latency.
- Add the dependency `modernc.org/sqlite` v1.60.1, a pure-Go SQLite driver.

## Capabilities

### New Capabilities

- `skipped-item-recording`: Persisting each item a parser skips, with its failure reason, to a deduplicated SQLite table for later debugging.
- `ptt-item-failure-handling`: How the PTT parser treats per-article failures and an empty listing.

### Modified Capabilities

(none)

## Impact

- **Code**: New recorder hook in the root package and a new SQLite recorder package. Skip reporting added to five parsers. `site/ptt.go` error handling and `site/thinkingtaiwan.go` fetch loop rewritten. `web/main.go` opens the database at startup.
- **API**: `GET /ptt` stops answering 400 when only the last listed article fails, and starts answering 400 with `ItemFetchError` when the page is not a recognizable list or every article fails. A search without results still returns an empty feed. No other endpoint changes behavior.
- **Dependencies**: `modernc.org/sqlite` v1.60.1 in `go.mod` and `go.sum`. The Dockerfile builds with `CGO_ENABLED=0`, which rules out cgo-based drivers. The Dockerfile's `ca-certificates` pin moves to `20260909-r0` because the old one left the Alpine index.
- **Deployment**: The database lands in the `log` directory that `compose.yaml` already mounts from the host. The runtime image has no `sqlite3` client, so the file is read from the host.
- **Docs**: The archived `openspec/changes/archive/2026-10-06-add-thinkingtaiwan/` artifacts are corrected where they describe the concurrent fetch.
