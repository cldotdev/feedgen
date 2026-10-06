## 1. Recorder Hook

- [x] 1.1 Add `UnexpectedStatusError{SourceURL, StatusCode}` to `error.go` with the message `unexpected status <code> from <url>`
- [x] 1.2 Add `SkipReason` and its four constants, `SkippedItem`, `SkipRecorder`, `SetSkipRecorder`, `RecordSkip`, a default recorder that writes one line through the standard `log` package, and the error-to-reason helper to the root package
- [x] 1.3 Add a unit test of the error-to-reason helper covering `PageContentNotFoundError`, `UnexpectedStatusError` (also when wrapped), and a plain error

## 2. SQLite Recorder

- [x] 2.1 Add `modernc.org/sqlite` v1.60.1 to `go.mod` and `go.sum`
- [x] 2.2 Create package `fetcherror` with `Open`, `RecordSkip`, and `Close`, creating the `skipped_items` table on open, with WAL journaling, a busy timeout, one open connection, UTC RFC 3339 timestamps, and an injectable `now`
- [x] 2.3 Implement the write as an upsert keyed on `(parser, item_url, reason)` that updates `source_url`, `message`, and `last_seen` and increments `count`
- [x] 2.4 Add unit tests using `t.TempDir()`: a first write creates a row with count 1; a repeat increments count, moves `last_seen`, and keeps `first_seen`; a different reason for the same article adds a second row; opening an existing database keeps its rows
- [x] 2.5 Open `log/fetch_error.db` in `web/main.go` after `setLogger` and install it; on failure, log the error and keep the default recorder

## 3. Parsers

- [x] 3.1 In `site/ptt.go`, use a loop-local error for `GetFeedItem`, record each skip with the classified reason, and return `ItemFetchError` when the listing matches no link or no item succeeds
- [x] 3.2 In `site/thinkingtaiwan.go`, fetch article pages serially in listing order, removing the semaphore, `sync.WaitGroup`, per-index slots, and `thinkingtaiwanMaxInFlight`; return `UnexpectedStatusError` from `thinkingtaiwanGet`; record each skip with the classified reason or `no_date`
- [x] 3.3 In `site/scitechvista.go` and `site/chrb.go`, build the item link before the date check and record each undated entry with `no_date`
- [x] 3.4 In `parser/udn.go`, record each zero-timestamp article as parser `udn` with `no_date`

## 4. Archived Change Correction

- [x] 4.1 In `openspec/changes/archive/2026-10-06-add-thinkingtaiwan/`, correct Decision 2, the concurrency Goal, the request-amplification risk, and tasks 1.4 and 4.2 in `design.md` and `tasks.md` to describe serial fetching and the timing evidence

## 5. Verification

- [x] 5.1 Run `go vet ./...` and `make install`
- [x] 5.2 Run `go test -count=1 ./...` and confirm the new unit tests and the existing integration tests pass
- [x] 5.3 Run `docker build --no-cache --output type=cacheonly .` and confirm the image builds with the new dependency

## 6. Review Fixes

- [x] 6.1 In `site/ptt.go`, return an empty feed instead of `ItemFetchError` when the listing has no article link but still renders the `r-list-container` element, so a search without results is not an error
- [x] 6.2 Bump the Dockerfile pin to `ca-certificates=20260909-r0`, since the Alpine 3.23 index no longer carries `20260611-r0`, and rerun the image build check
