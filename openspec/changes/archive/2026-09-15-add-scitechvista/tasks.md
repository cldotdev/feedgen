## 1. Site Parser

- [x] 1.1 Create `site/scitechvista.go` with `ScitechvistaParser` implementing `feedgen.Parser`, a dedicated `http.Client` with a 30-second timeout, and the `scitechvistaSections` map for `new`, `hot`, and `featured`
- [x] 1.2 Validate the `section` parameter, returning `ParameterNotFoundError` when missing and `ParameterValueInvalidError` when unsupported
- [x] 1.3 Match the section's listing entries one anchor at a time, keeping the 推薦文章 sidebar out of the feed
- [x] 1.4 Extract title, relative link, excerpt, author, and date per entry, treating the author as optional and unescaping HTML character references in the text fields
- [x] 1.5 Convert ROC-calendar dates to `time.Time` in a fixed UTC+8 zone
- [x] 1.6 Build feed items with the absolute article URL as both `Id` and `Link`, preserving the listing's newest-first order
- [x] 1.7 Return an error naming the source URL on a non-2xx response, and `ItemFetchError` when no entry matches

## 2. Tests

- [x] 2.1 Create `site/scitechvista_test.go` covering feed generation for each supported section and its expected feed title
- [x] 2.2 Add tests for a missing `section` parameter and an unsupported `section` value
- [x] 2.3 Add a test asserting every item has a non-empty title, an absolute link equal to its id, and a created timestamp whose year is later than 2000
- [x] 2.4 Add a test asserting items are ordered newest first, which catches sidebar entries leaking into the feed
- [x] 2.5 Add a test that the first article's link returns an HTTP 2xx status
- [x] 2.6 Run `go test ./site/...` and confirm the new tests pass

## 3. Wiring and Documentation

- [x] 3.1 Register `GET /scitechvista` in `web/main.go`
- [x] 3.2 Add a 科技大觀園 entry with example URLs to `README.md`
- [x] 3.3 Run `go vet ./...` and `go build ./...` to confirm the project still builds clean

## 4. Simplify

- [x] 4.1 Drop the redundant `path` field, reducing the section map to a slug-to-feed-title `map[string]string` and building the URL from the slug
- [x] 4.2 Remove the test helper that existed only to flatten that map, ranging the map directly instead
- [x] 4.3 Preallocate `feed.Items` to the matched entry count, as `parser/udn.go` does
- [x] 4.4 Remove the doc comment on `scitechvistaCreated` that restated the function name
