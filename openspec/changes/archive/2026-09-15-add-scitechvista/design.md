## Context

See proposal.md - Why. 科技大觀園 serves its article listings as server-rendered HTML with no JSON API behind them, which rules out the `parser/udn.go` fetch-and-unmarshal path that `site/udn_game.go` and `site/udn_global_vision.go` share. The closest precedents are `site/chrb.go` and `site/ptt.go`, which both scrape HTML with `regexp` from the standard library; the project has no HTML parsing dependency such as goquery in `go.mod`, and `go.mod` pins Go 1.26.1.

Two properties of the listing page shape the approach:

- `https://scitechvista.nat.gov.tw/Article/C000003/new` renders two article lists inside `div.kf-diagramtext-col` wrappers: the section's own 5 entries, whose anchors carry `class="kf-item align-items-center"`, and a 4-entry 推薦文章 sidebar, whose anchors carry `class="kf-item py-3"`. The sidebar entries are older than the section entries, so leaking them produces an out-of-order feed.
- Dates are printed in the ROC calendar with no time component (`115/08/31`).

## Goals / Non-Goals

**Goals:**

- Scrape the 文章 channel (`C000003`) listings with the standard library only.
- Keep the section set open to extension through a static map, the way `globalVisionTags` works in `site/udn_global_vision.go`.
- Fail loudly when the markup stops matching, rather than serving an empty feed.

**Non-Goals:**

- Pagination. The listing exposes `PageIndex`, but each feed serves the first page only, matching the 轉角國際 precedent.
- The video channels `C000008` (科普影片) and `C000009` (TechTalk), and the per-category listings under `C000003/category/<uuid>`. Their markup is identical, so adding them later stays cheap.
- Fetching each article's detail page to enrich the item body. The listing's excerpt is the description.

## Decisions

### 1. Regex scraping in `site/scitechvista.go`

Parse the listing with `regexp` against the raw response body, as `site/chrb.go` and `site/ptt.go` do.

**Rationale**: No HTML parser is in `go.mod`, and the two existing HTML-scraping parsers set the precedent. Adding goquery for one site would be the only dependency of its kind in the project.

**Alternatives considered**:

- Add `github.com/PuerkitoBio/goquery`: More robust against markup drift and nicer to read, but it is a new dependency for a single parser and diverges from how the other scrapers are written.

### 2. Two-stage matching: isolate each entry, then extract fields

Match the section's anchors one at a time with `<a href="(/Article/[^"]+)"[^>]*class="kf-item align-items-center">(.*?)</a>`, where the non-greedy tail stops at the entry's own closing tag because the listing nests no anchors inside an entry. Then run small per-field regexes over each captured chunk.

**Rationale**: The `Article-AuthorRow` block is conditional markup. A single `(?s)` regex spanning the anchor through the author would, on an entry that has no author, run past the entry boundary and swallow the next entry's author, silently dropping articles. Bounding each chunk to one anchor first keeps every field match inside one entry and makes the author optional without risk. Anchoring on `align-items-center` also excludes the 推薦文章 sidebar, whose anchors carry `kf-item py-3`, by construction rather than by a negative lookahead, which Go's RE2 engine does not support anyway.

**Alternatives considered**:

- One monolithic regex per entry: shorter, but couples the fields and loses entries whenever an optional block is absent.
- Splitting on the `kf-diagramtext-col` wrappers: those wrap the sidebar entries too, so the section still has to be told apart by the anchor class; matching the anchor directly does both jobs at once.
- Truncating the body at the 推薦文章 heading: handles the sidebar but not the optional-author problem.

### 3. Section mapping via a static map

Map each accepted section slug to its feed title:

```go
var scitechvistaFeedTitles = map[string]string{
    "new":      "最新文章 | 科技大觀園",
    "hot":      "熱門文章 | 科技大觀園",
    "featured": "精選文章 | 科技大觀園",
}
```

The slug doubles as the upstream path segment, so the URL is built from the slug directly rather than from a second field that would always hold the same string.

**Rationale**: Mirrors `globalVisionTags`, keeps the endpoint's accepted values explicit, and turns an unknown section into a `ParameterValueInvalidError` instead of a request to a non-existent upstream page. The three sections share identical markup, verified against all three pages.

**Alternatives considered**:

- Pass the section through to the URL unchecked: one less map, but any typo becomes an upstream 404 surfaced as an opaque error, and the feed title cannot be localized.
- Hardcode `new` with no parameter: the user's request names only `new`, but `hot` and `featured` cost one map entry each.
- Carry an explicit path field per section: needed only if a slug ever has to differ from its upstream segment, which none does today; a slug that must differ can grow the map value back into a struct then.

### 4. ROC date conversion with a fixed UTC+8 zone

Parse `<ROC year>/<month>/<day>`, add 1911 to the year, and build the timestamp with `time.Date(..., time.FixedZone("CST", 8*60*60))`.

**Rationale**: `time.LoadLocation` returns a nil location when the container image ships without tzdata, and passing that nil to `Time.In` panics. That is the bug commit 3a5f001 fixed in `site/ptt.go`, and this parser uses the same fixed zone. The listing prints no time of day, so items land at midnight UTC+8; this only affects ordering within a single day, and the listing is already ordered newest first.

**Alternatives considered**:

- `time.LoadLocation("Asia/Taipei")`: the failure the ptt fix exists to avoid.
- Import `time/tzdata` to embed the database: fixes the panic but adds about 450 KB to the binary for one arithmetic conversion.

### 5. Absolute URLs as both link and id

Prefix the relative `href` with `https://scitechvista.nat.gov.tw` and use the result for both `Item.Link` and `Item.Id`.

**Rationale**: Atom ids must be stable and globally unique; the article UUID is embedded in the query string, so the absolute URL satisfies both. Every other parser in `site/` already uses the article URL as the id.

### 6. Empty match set is an error

Return `feedgen.ItemFetchError` when no entry matches, and an error naming the source URL on a non-2xx response.

**Rationale**: `site/chrb.go` does the same. A silently empty feed looks to a feed reader like the site stopped publishing, which hides the breakage; an error surfaces as HTTP 400 with a message.

### 7. Dedicated HTTP client with a timeout

Give the parser its own `http.Client` with a 30-second timeout, as `parser/udn.go` and `site/ptt.go` do.

**Rationale**: `http.DefaultClient` has no timeout, so a stalled government host would hold the request open indefinitely.

### 8. Preserve the listing's order rather than re-sorting

Emit items in the order the listing renders them, instead of calling `feedgen.SortFeedItemsLatestFirst`.

**Rationale**: All three listings already render newest first, so a sort would be a no-op on correct output. It would not be a no-op on incorrect output: sorting makes the newest-first assertion in the tests hold no matter what the parser scraped, which throws away the cheapest signal that sidebar entries leaked in. Leaving the order alone keeps that assertion meaningful. The listing also prints no time of day, so `sort.Slice` in the shared helper could permute same-day items, which the listing's own order gets right.

**Alternatives considered**:

- Call `feedgen.SortFeedItemsLatestFirst`: consistent with `site/gamer_forum.go`, but that parser merges several sources and genuinely needs it; here it only hides breakage.

### 9. Skip entries with no parseable date

Drop an entry whose date cannot be read, rather than emitting it with a zero timestamp.

**Rationale**: `parser/udn.go` skips articles with a zero timestamp for the same reason: one malformed entry should not poison the rest of the feed, and a zero timestamp would sort the item to the end of every reader's list and break the newest-first order the rest of the feed keeps. The empty-match error in Decision 6 still catches the case where the date markup changes for every entry at once.

## Risks / Trade-offs

- **[Markup drift]** The regexes depend on Bootstrap-style class names (`kf-item align-items-center`, `kf-date`, `kf-title`, `kf-txt`, `text-truncate Author`) that a site redesign would change. → Mitigation: the empty-match error turns drift into a visible failure, and the integration tests catch it on the next run.
- **[Thin feed]** Five items per page means a reader that polls less often than the site publishes will miss articles. → Mitigation: accepted for now, matching the 轉角國際 precedent; pagination is a small follow-up if it proves a problem.
- **[Post-quantum TLS ClientHello]** Go 1.26 sends MLKEM key shares by default, which some middleboxes reject with a connection reset; this is the bug fixed for `www.ptt.cc` in commit 9990be1. `scitechvista.nat.gov.tw` may sit behind similar equipment. → Mitigation: the integration tests connect to it without a reset today, so the parser uses a plain client; if a reset appears later, reuse the classic-curves `http.Transport` from `site/ptt.go`.
- **[Midnight timestamps]** Without a time of day, all items published on the same date share a timestamp. → Mitigation: the parser preserves the listing's own order, so same-day items stay in the order the site presents them.
