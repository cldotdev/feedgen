# udn-shared-parser

Shared UDN JSON API types and article-fetching logic in `parser/udn.go`,
reusable across UDN family site parsers.

## Requirements

### Requirement: Shared UDN article types

The `parser` package SHALL define shared types for UDN JSON API responses,
including `Article` with fields for title, author name, paragraph, URL, and
created time (`time.Time`).

#### Scenario: Types are importable by site parsers

- **WHEN** a site parser imports `github.com/cldotdev/feedgen/parser`
- **THEN** it can use `parser.Article` and access all fields needed to
  construct a `feeds.Item`

### Requirement: Fetch and parse UDN articles

The `parser` package SHALL export a `FetchArticles(rawLink string)
([]Article, error)` function that performs an HTTP GET request with a
30-second timeout, validates the HTTP response status, unmarshals the JSON
response, and returns a slice of `Article` with timestamps converted to
`time.Time`.

#### Scenario: Successful fetch from a valid UDN API endpoint

- **WHEN** `FetchArticles` is called with a valid UDN JSON API URL
- **THEN** it returns a non-empty slice of `Article` with no error
- **THEN** each article has a non-empty `Title`, `URL`, and non-zero
  `Created`

#### Scenario: Non-2xx HTTP response

- **WHEN** the upstream API returns a non-2xx status code
- **THEN** `FetchArticles` returns an error describing the status code

#### Scenario: Timestamp conversion for 10-digit Unix timestamps

- **WHEN** the API returns an article with a 10-digit timestamp
- **THEN** the `Created` field is parsed as seconds since epoch

#### Scenario: Timestamp conversion for 13-digit Unix timestamps

- **WHEN** the API returns an article with a 13-digit timestamp
- **THEN** the `Created` field is parsed as milliseconds since epoch

#### Scenario: Zero timestamp is skipped

- **WHEN** the API returns an article with a timestamp of 0
- **THEN** that article is excluded from the result slice

#### Scenario: Invalid timestamp digit count

- **WHEN** the API returns an article with a non-zero timestamp that is
  neither 10 nor 13 digits
- **THEN** `FetchArticles` returns an `ItemFetchError`

### Requirement: Add articles to feed

The `parser` package SHALL export an `AddArticlesToFeed(feed, articles)`
function that appends a slice of `Article` as feed items to a given feed.

#### Scenario: Articles are added as feed items

- **WHEN** `AddArticlesToFeed` is called with a feed and a slice of articles
- **THEN** each article is added to the feed with `Id`, `Title`, `Link`,
  `Description`, `Author`, and `Created` fields mapped from the article
