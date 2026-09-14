## Why

科技大觀園 (scitechvista.nat.gov.tw) is the National Science and Technology Council's popular-science publication. It publishes new articles regularly but offers no Atom or RSS feed, so readers have to poll the site by hand. Adding it to feedgen lets them subscribe instead.

## What Changes

- Add `site/scitechvista.go` implementing a `ScitechvistaParser` that scrapes the article listing pages of the 文章 channel (`C000003`).
- Support a required `section` query parameter accepting `new` (最新文章), `hot` (熱門文章), and `featured` (精選文章); the three listings share identical markup.
- Convert the listing's ROC-calendar dates (for example `115/08/31`) to `time.Time` in UTC+8.
- Add `site/scitechvista_test.go` with integration tests.
- Register `GET /scitechvista` in `web/main.go`.
- Add a 科技大觀園 entry to `README.md`.

## Capabilities

### New Capabilities

- `scitechvista`: Site parser for 科技大觀園 article listings, exposing `GET /scitechvista?section=new|hot|featured` as an Atom feed.

### Modified Capabilities

(none)

## Impact

- **Code**: New `site/scitechvista.go` and `site/scitechvista_test.go`. One route registration in `web/main.go`. No existing parser is touched.
- **API**: New endpoint `GET /scitechvista?section=<section>`. No changes to existing endpoints.
- **Dependencies**: No new external dependencies. The listing is server-rendered HTML, parsed with the standard library `regexp` and `html` packages, matching `site/chrb.go` and `site/ptt.go`.
- **Tests**: New integration tests that make real HTTP requests to `scitechvista.nat.gov.tw`.
