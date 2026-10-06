## Why

想想論壇 (`www.thinkingtaiwan.net`) publishes policy and current-affairs commentary every week but offers no Atom or RSS feed: `/feed` and `/rss.xml` return 404 and the Drupal JSON:API is disabled. Adding it to feedgen lets readers subscribe instead of polling the site by hand.

## What Changes

- Add `site/thinkingtaiwan.go` implementing a `ThinkingtaiwanParser` that scrapes the first page of the site-wide latest-articles listing at `/articles`.
- Fetch each listed article's page for its publication date and author, because the listing carries neither.
- Take no query parameters. `/articles` already contains every article of every `/topics/*` listing, so a topic filter would only narrow it.
- Add `site/thinkingtaiwan_test.go` with integration tests.
- Register `GET /thinkingtaiwan` in `web/main.go`.
- Add a 想想論壇 entry to `README.md`.

## Capabilities

### New Capabilities

- `thinkingtaiwan`: Site parser for 想想論壇's latest articles, exposing `GET /thinkingtaiwan` as an Atom feed.

### Modified Capabilities

(none)

## Impact

- **Code**: New `site/thinkingtaiwan.go` and `site/thinkingtaiwan_test.go`. One route registration in `web/main.go`. No existing parser is touched.
- **API**: New endpoint `GET /thinkingtaiwan`. No changes to existing endpoints.
- **Dependencies**: None. The pages are server-rendered HTML, parsed with the standard library `regexp` and `html` packages, matching `site/scitechvista.go` and `site/chrb.go`.
- **Upstream load**: Each uncached feed request makes one listing request plus one request per listed article (nine today). The Redis response cache in `web/main.go` absorbs repeats only when `FG_REDIS_HOST` is set.
- **Tests**: New integration tests that make real HTTP requests to `www.thinkingtaiwan.net`.
