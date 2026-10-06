## 1. Site Parser

- [x] 1.1 Create `site/thinkingtaiwan.go` with `ThinkingtaiwanParser` implementing `feedgen.Parser`, using the shared `client` and `zone` from `site/util.go`
- [x] 1.2 Fetch `https://www.thinkingtaiwan.net/articles`, returning an error naming the URL on a non-2xx response
- [x] 1.3 Extract each listing row's href (`/article/<id>` or `/content/<id>`), title, and teaser, bounding the field regexes to one row and unescaping HTML character references; return `ItemFetchError` when no row matches
- [x] 1.4 Fetch each article page with at most two requests in flight, reading the `post-date` date as midnight in `zone` and the `field--name-field-writer` name as the author
- [x] 1.5 Skip an article whose page fails to load, returns a non-2xx status, or has no parseable date; keep an article with no writer, with an empty author
- [x] 1.6 Build items in listing order with the absolute article URL as both `Id` and `Link`, the feed titled "最新文章 | 想想論壇" and linked to `/articles`; return `ItemFetchError` when every article was skipped

## 2. Tests

- [x] 2.1 Create `site/thinkingtaiwan_test.go` with a test of the feed title, feed link, and a non-empty item list
- [x] 2.2 Add a test asserting every item has a non-empty title, a non-empty author, an absolute link under `https://www.thinkingtaiwan.net/` equal to its id, and a non-zero created timestamp
- [x] 2.3 Add a test asserting no item has a later created timestamp than the item before it
- [x] 2.4 Add a test that the first article's link returns an HTTP 2xx status
- [x] 2.5 Run `go test -count=1 ./site/...` and confirm the new tests pass

## 3. Wiring and Documentation

- [x] 3.1 Register `GET /thinkingtaiwan` in `web/main.go`
- [x] 3.2 Add a 想想論壇 entry with the example URL `https://feedgen.org/thinkingtaiwan` to `README.md`
- [x] 3.3 Run `go vet ./...` and `make install` to confirm the project still builds clean

## 4. Review Fixes

- [x] 4.1 Bound the date pattern to the `post-date` element, so a page without a date is skipped rather than matched against a later date on the page
- [x] 4.2 Run the 想想論壇 tests with `-race` to confirm the bounded concurrent fetch is race-free
