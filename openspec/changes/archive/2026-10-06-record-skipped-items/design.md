## Context

See proposal.md - Why. Logging today is set up once in `web/main.go` `setLogger`: the standard `log` package and gin both write to `log/production.log` (or `log/development.log` outside release mode), and `compose.yaml` mounts `./.srv/log` onto `/app/log`. The `site` and `parser` packages write no logs. Both import the root `feedgen` package, which holds the shared error types.

Skip points today:

| Location | Skips when |
|---|---|
| `site/ptt.go` `GetFeed` loop | `GetFeedItem` fails: network or read error, or `PageContentNotFoundError` for a `404 - Not Found` page |
| `site/thinkingtaiwan.go` | article fetch fails, returns non-2xx, or shows no `post-date` |
| `site/scitechvista.go` | listing entry has no parseable ROC date |
| `site/chrb.go` | listing card has no parseable date |
| `parser/udn.go` `FetchArticles` | API article has timestamp 0 (shared by `udn_game` and `udn_global_vision`) |

`site/ptt.go` `GetFeed` assigns `GetFeedItem`'s error to the named return `err` inside the loop, so the value left after the loop is the last article's; and when the listing matches no link it returns the half-built feed with a nil error.

For 想想論壇, a timing probe fetched the same nine article pages serially and two at a time: totals were 3.15 s and 3.34 s serial against 3.33 s and 3.32 s concurrent, and each request took about 670 ms with two in flight against about 330 ms alone. The server appears to queue requests from one client; that is inferred from timing, not from its configuration.

The Dockerfile builds with `CGO_ENABLED=0`, and `go.mod` pins Go 1.26.1.

## Goals / Non-Goals

**Goals:**

- Keep the parsers free of any database import; a parser only reports a skip.
- Make recording best-effort: no recorder failure reaches a feed response.

**Non-Goals:**

- Changing `production.log`, its format, or its rotation.
- Storing page bodies.
- Pruning or rotating `fetch_error.db`. One row per key bounds its growth by the number of distinct failing articles.
- Recording listing-level failures (non-2xx listing, no entries matched). Those already surface as an error response and a line in `production.log`.

## Decisions

### 1. Recorder hook in the root package

Add to the root `feedgen` package:

- `SkipReason`, a string type with the constants `SkipReasonNetwork`, `SkipReasonHTTPStatus`, `SkipReasonNotFound`, and `SkipReasonNoDate` for `network`, `http_status`, `not_found`, and `no_date`.
- `SkippedItem` with `Parser`, `SourceURL`, `ItemURL`, `Reason`, and `Message`.
- `SkipRecorder`, an interface with one method, `RecordSkip(SkippedItem) error`.
- `SetSkipRecorder(SkipRecorder)` and a package-level `RecordSkip(SkippedItem)` that forwards to the installed recorder and writes any error it returns through the standard `log` package.
- A default recorder that writes one line per skip through the standard `log` package.

`SetSkipRecorder` is called once in `main` before the router starts serving, so the package-level variable needs no synchronization.

**Rationale**: `site` and `parser` already import the root package, so neither gains an import, and only the binary links the SQLite driver. The default recorder makes the hook safe to call from tests and from a server whose database failed to open, and it leaves the skip visible in test output and in `production.log`.

**Alternatives considered**:

- Return skipped items from `GetFeed` and record them in `route`: keeps the parsers free of global state, but changes the `Parser` interface and every implementation for a diagnostic.
- Pass a recorder into each parser struct: the parsers are registered as zero-value structs in `web/main.go` and constructed directly in tests, so every construction site would change.

### 2. Reason classification from error types

Add `UnexpectedStatusError{SourceURL string, StatusCode int}` to `error.go`, with the message `unexpected status <code> from <url>` that the parsers already produce with `fmt.Errorf`. Add `SkipReasonOf(err)` to the root package to map an error to a `SkipReason`: `PageContentNotFoundError` to `not_found`, `UnexpectedStatusError` to `http_status`, anything else to `network`. It matches both the value and the pointer form of each type, because the error methods have value receivers and the parsers return pointers such as `&feedgen.PageContentNotFoundError{...}`. Parsers that skip for a missing date pass `no_date` directly.

Only `thinkingtaiwanGet` switches to `UnexpectedStatusError` in this change, because it is the only article-page fetch that checks the status code. The listing-level `fmt.Errorf` calls in other parsers keep their current form.

**Rationale**: The raw error text cannot be the deduplication key: a network error carries the local ephemeral port (`read tcp 172.18.0.3:54321->...`), so every occurrence would be a new row. Classifying with `errors.As` keeps the reasons stable while the raw text still goes into `message`.

### 3. SQLite recorder in its own package

Add a package `fetcherror` with `Open(path string) (*Recorder, error)`, `(*Recorder).RecordSkip(feedgen.SkippedItem) error`, and `Close() error`, using `modernc.org/sqlite` v1.60.1 through `database/sql`.

Schema, created with `CREATE TABLE IF NOT EXISTS` on open:

```sql
CREATE TABLE skipped_items (
    parser     TEXT    NOT NULL,
    item_url   TEXT    NOT NULL,
    reason     TEXT    NOT NULL,
    source_url TEXT    NOT NULL,
    message    TEXT    NOT NULL,
    first_seen TEXT    NOT NULL,
    last_seen  TEXT    NOT NULL,
    count      INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (parser, item_url, reason)
);
```

A write is one `INSERT ... ON CONFLICT (parser, item_url, reason) DO UPDATE SET source_url = excluded.source_url, message = excluded.message, last_seen = excluded.last_seen, count = count + 1`. Timestamps are RFC 3339 strings in UTC, so they sort as text and read correctly in the `sqlite3` client. The recorder takes the current time from a `now func() time.Time` field that defaults to `time.Now`, so a test can move the clock.

Open the database with WAL journaling and a busy timeout, and limit the pool to one open connection.

**Rationale**: `CGO_ENABLED=0` rules out `github.com/mattn/go-sqlite3`. Of the two pure-Go drivers, `modernc.org/sqlite` is past 1.0, while `github.com/ncruces/go-sqlite3` (v0.35.6) is not. gin serves requests concurrently and SQLite admits one writer at a time; one connection serializes writes inside the process, and the busy timeout covers an external reader such as an operator's `sqlite3` session holding a lock. The upsert keeps one row per key, as the user asked, and the count and last-seen time show whether a failure is still happening.

**Alternatives considered**:

- `INSERT ... ON CONFLICT DO NOTHING`: as cheap, but loses whether the failure persists.
- A text log with an in-memory set of seen keys: no dependency, but the set resets on restart and the file cannot answer "how often" without post-processing.

### 4. Wiring in `web/main.go`

After `setLogger`, open `log/fetch_error.db`, built from the same `log` directory `setLogger` creates, and install it with `feedgen.SetSkipRecorder`. If opening fails, write the error with `log.Println` and keep the default recorder. Development mode records to the same file name.

**Rationale**: The `log` directory is already created at startup and mounted from the host, and `.gitignore` already ignores it.

### 5. Parser names and item URLs

Report the route name as `Parser`: `ptt`, `thinkingtaiwan`, `scitechvista`, `chrb`. `parser/udn.go` reports `udn`, because `FetchArticles` does not know which site called it; `SourceURL` tells the two UDN feeds apart. `ItemURL` is the absolute article URL the item would have carried, so `site/scitechvista.go` and `site/chrb.go` build the link before the date check instead of after it.

### 6. PTT error handling

In the `GetFeed` loop, hold `GetFeedItem`'s error in a loop-local variable, record the skip, and continue. After the loop, return `ItemFetchError` for the listing URL when no item succeeded. When the listing matches no article link, return an empty feed if the page still contains the `r-list-container` element, and `ItemFetchError` otherwise.

**Rationale**: Matches `site/thinkingtaiwan.go` and the specs of the other parsers: one bad article never decides the response. A PTT search without results answers 200 with the list container and no `r-ent` row (verified with `/bbs/Steam/search?q=zzqxjkvnoresult`), which is a legitimate empty result for a strict filter such as `q=recommend:99` on a quiet board; a page without the container, such as a missing board's 404 page or changed markup, is the failure the error is for.

**Alternatives considered**:

- Return `ItemFetchError` for any listing without links: turns an empty search into an HTTP 400 that a feed reader shows as a broken feed.
- Return an empty feed only when `q` is set: needs no extra check, but cannot tell an empty search from changed markup.

### 7. Serial fetching for 想想論壇

Replace the semaphore, `sync.WaitGroup`, per-index slots, and `thinkingtaiwanMaxInFlight` with a plain loop that fetches each article page in listing order.

**Rationale**: Two in flight measured no faster than serial, so the concurrency only added code. Serial fetching also matches `site/ptt.go`. Correct Decision 2, its matching Goal, and the matching tasks in `openspec/changes/archive/2026-10-06-add-thinkingtaiwan/`, which the user approved; `openspec/specs/thinkingtaiwan/spec.md` states no concurrency and stays as is.

## Risks / Trade-offs

- **[Binary size]** `modernc.org/sqlite` embeds a transpiled SQLite, adding several MB to the binary (an estimate, not measured). → Mitigation: accepted with the dependency.
- **[Image build]** `go.mod` and `go.sum` change. → Mitigation: run the local image build check that `CLAUDE.md` defines. That check failed on the Dockerfile's `ca-certificates=20260611-r0` pin, which the Alpine 3.23 index no longer carries; the pin moves to `20260909-r0`, the version the index now serves.
- **[Write latency on the request path]** Each skip adds one synchronous SQLite write to the request that found it. → Mitigation: skips are rare, and with the Redis cache a failing URL is fetched at most once per cache period.
- **[Global recorder]** A package-level recorder is shared state. → Mitigation: it is set once before serving; tests that do not set it get the default, which writes no file.
- **[Unbounded rows]** Each distinct failing article adds a row forever. → Mitigation: accepted; a feed lists a few dozen articles at most, so growth tracks the number of distinct failures, not requests.

## Migration Plan

Deploying creates `log/fetch_error.db` on first start. Rolling back removes the recorder; the file stays on the host and can be deleted by hand.
