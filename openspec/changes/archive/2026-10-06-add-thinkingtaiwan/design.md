## Context

See proposal.md - Why. 想想論壇 is a Drupal site that serves only server-rendered HTML. The closest precedent is `site/scitechvista.go`: regex scraping with the standard library, one anchor-bounded chunk per entry, the shared `client` and `zone` from `site/util.go`, and no re-sorting. `site/ptt.go` is the precedent for fetching each item's own page after the listing.

Properties of the site, verified against the live pages, that shape the approach:

- `/articles` (Drupal view `list`, display `page_2`) lists 9 articles per page, newest first. Each row is a `<div class="entity-row-inner views-row-inner">` holding an `<h2><a href="/article/<id>">` or `<h2><a href="/content/<id>">` title and a `<div class="field-body">` teaser. Both URL forms appear in the same listing. Rows carry no date and no author.
- Crawling every page of `/articles` and of each `/topics/*` listing showed that the six top-level topics partition `/articles` exactly (210 articles, no duplicates, none outside a topic), and the numeric sub-topics partition 思想政策. English Edition articles (`/en` URLs) are separate content and do not appear in `/articles`.
- An article page shows `<div class="post-date"><span>發佈於<span class="divider">｜</span>YYYY-MM-DD` with no time of day, and the writer in `<div class="field field--name-field-writer ... field__item">NAME</div>`. It has no JSON-LD and no `article:published_time` meta tag.
- The site publishes about three articles together once a week, so one listing page covers roughly three weeks and same-day ties are the norm.
- Fifteen listing requests fired at once returned several pages with no entries; the same requests made one at a time all succeeded. The cause, throttling or a transient error, was not established.

## Goals / Non-Goals

**Goals:**

- Give every item a real publication date, so the Atom output stays valid.
- Keep upstream load low enough not to trip whatever made the parallel crawl fail, by fetching one request at a time.

**Non-Goals:**

- A `topic` parameter. `/articles` is the union of every topic listing, so a filter adds no articles; it can be added later as a slug whitelist without changing the default URL.
- `/archive` (the old 想想 1.0 site, a different view) and English Edition.
- Pagination beyond the first page, matching the other parsers.
- Article body text. The listing teaser is the description.

## Decisions

### 1. Date and author come from each article page

Fetch the listing once, then each listed article's page, and read the date and writer there.

**Rationale**: No other source carries the publication date.

**Alternatives considered**:

- Sitemap `lastmod`: it is the node's modification time, not its publication time (`content/100471` shows an edit at 08:36:44 among scheduled on-the-minute times, and older nodes carry 2026 dates), and the sitemap is about 1 MB across four pages.
- Leave the date unset: `gorilla/feeds` formats a zero `Created` and `Updated` as an empty string, which emits an empty `<updated>` element and makes the Atom output invalid.

### 2. Serial article fetches in listing order

Fetch the article pages one at a time in listing order and add each usable item to the feed as it completes.

**Rationale**: A timing probe fetched the same nine article pages serially and two at a time: totals were 3.15 s and 3.34 s serial against 3.33 s and 3.32 s with two in flight, and each request took about 670 ms with two in flight against about 330 ms alone. The server appears to queue requests from one client (inferred from timing, not from its configuration), so concurrency adds code without cutting latency. Serial fetching also matches `site/ptt.go` and keeps the listing order without a sort.

**Alternatives considered**:

- Two requests in flight with per-index result slots: measured no faster than serial, and adds a semaphore, a wait group, and result slots.
- Unbounded fan-out: nine simultaneous requests is close to the burst that already failed.

### 3. Regex scraping, one row at a time

Split the listing into rows on the `entity-row-inner views-row-inner` wrapper, then extract the href, title, and teaser per row. Accept `/(article|content)/<digits>` hrefs. On the article page, match `post-date` and `field--name-field-writer` independently. The date pattern matches the first `YYYY-MM-DD` inside the `post-date` element, after its label and divider, without matching the `發佈於` label text itself.

**Rationale**: Bounding the field regexes to one row keeps a row with a missing teaser from absorbing the next row's, the same reasoning as `site/scitechvista.go`. Bounding the date pattern to the `post-date` element keeps a page without a date from matching a later date elsewhere on the page, so Decision 6 skips it as intended; leaving the label out of the pattern tolerates a reworded label. No HTML parser is in `go.mod`.

**Alternatives considered**:

- Match each row's `<h2>` anchor and the teaser that follows: works, but splitting on the wrapper bounds every field with one mechanism.
- Let the date pattern scan forward from `post-date` with a lazy wildcard: simpler, but on a page that lost its date it picks up any later date instead of skipping the article.

### 4. Midnight UTC+8 timestamps from the shared fixed zone

Build `Created` with `time.Date(year, month, day, 0, 0, 0, 0, zone)`, using `zone` from `site/util.go`.

**Rationale**: The page shows only a date. `zone` avoids `time.LoadLocation`, which fails without tzdata in the runtime image.

### 5. Preserve listing order rather than re-sorting

**Rationale**: The listing is already newest first, and same-day items share a timestamp, so a sort could only permute ties. Keeping the order also keeps the newest-first test meaningful, as in `site/scitechvista.go`.

### 6. Skip failed articles; error only when nothing is left

Skip an article whose page fetch fails, returns a non-2xx status, or has no parseable date. Return `ItemFetchError` when the listing has no entries or when every article was skipped, and an error naming the URL when the listing itself returns a non-2xx status. A missing writer does not skip the article; the author is left empty.

**Rationale**: One bad article should not empty the feed, and a feed with zero items looks to a reader like the site stopped publishing.

### 7. Absolute URL as link and id

Prefix the href with `https://www.thinkingtaiwan.net` and use it for both `Item.Link` and `Item.Id`. The `/article/` or `/content/` form is kept as the listing gives it; both resolve and the page's own `og:url` uses the same form.

## Risks / Trade-offs

- **[Request amplification]** One feed request makes up to ten upstream requests when the Redis cache is off. → Mitigation: the site updates weekly and feed readers poll a few times a day, so the absolute volume stays small; requests are sent one at a time.
- **[Markup drift]** The regexes depend on Drupal theme class names (`entity-row-inner`, `field-body`, `post-date`, `field--name-field-writer`). → Mitigation: drift in the listing yields `ItemFetchError`, drift in the date yields the same once every article is skipped, and the integration tests catch both.
- **[Midnight timestamps]** Same-day articles share a timestamp. → Mitigation: listing order is preserved, so readers that sort stably still show them in the site's order.
- **[Partial markup drift]** If only some article pages change markup, those items silently drop from the feed. → Mitigation: accepted; the integration tests catch the drift only once every article drops.
