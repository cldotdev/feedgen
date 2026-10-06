## Purpose

Turns 想想論壇's latest-articles listing into an Atom feed, so readers can subscribe to the site's commentary instead of polling it by hand.

## Requirements

### Requirement: Feed generation for latest articles

The system SHALL expose a `GET /thinkingtaiwan` endpoint that takes no query parameters and returns an Atom feed of the articles on the first page of 想想論壇's latest-articles listing.

#### Scenario: Successful feed

- **WHEN** a request is made to `GET /thinkingtaiwan`
- **THEN** the response is an Atom feed with title "最新文章 | 想想論壇"
- **THEN** the feed link points at `https://www.thinkingtaiwan.net/articles`
- **THEN** the feed contains the articles listed on that page

#### Scenario: Both article URL forms included

- **WHEN** the listing mixes articles under `/article/<id>` and `/content/<id>`
- **THEN** the feed includes articles of both forms

### Requirement: Feed item content

Each feed item SHALL carry the article's title, absolute link, and summary as shown on the listing page, and its author and publication date as shown on the article's own page.

#### Scenario: Item fields are populated

- **WHEN** a feed is generated
- **THEN** every item has a non-empty title and a non-zero created timestamp
- **THEN** every item link is an absolute URL under `https://www.thinkingtaiwan.net/`
- **THEN** every item id equals its link

#### Scenario: Author taken from the article page

- **WHEN** an article page names a writer
- **THEN** the item's author is that writer's name

#### Scenario: Article without a named writer

- **WHEN** an article page names no writer
- **THEN** the item is still included in the feed with an empty author

#### Scenario: Title and summary are decoded

- **WHEN** a listing entry contains HTML character references such as `&amp;` or `&#039;`
- **THEN** the item title and description contain the decoded characters, not the references

### Requirement: Publication date

The parser SHALL take each item's created timestamp from the publication date on the article's page, which has day granularity, as midnight of that date in UTC+8, without depending on the system time zone database.

#### Scenario: Date converted to a UTC+8 timestamp

- **WHEN** an article page shows `發佈於｜2026-10-02`
- **THEN** the item's created timestamp is 2026-10-02 00:00 in UTC+8

### Requirement: Listing order

The feed SHALL keep the order in which the listing presents its articles, which is newest first.

#### Scenario: Items ordered newest first

- **WHEN** a feed is generated
- **THEN** no item has a later created timestamp than the item before it

#### Scenario: Same-day articles keep the listing order

- **WHEN** several listed articles share a publication date
- **THEN** they appear in the feed in the order the listing presents them

### Requirement: Per-article failure tolerance

The parser SHALL omit an article whose page cannot be retrieved or shows no recognizable publication date, and still return the remaining articles.

#### Scenario: One article page fails

- **WHEN** one listed article's page returns a non-2xx status, fails to load, or has no recognizable publication date
- **THEN** that article is left out of the feed
- **THEN** the other articles are still returned

### Requirement: Upstream failure reporting

The parser SHALL report an error instead of returning an empty feed when the listing cannot be retrieved or no article can be turned into an item.

#### Scenario: Listing page returns a non-2xx status

- **WHEN** the upstream listing request returns a status outside the 200-299 range
- **THEN** the parser returns an error naming the source URL

#### Scenario: No entries match the expected listing structure

- **WHEN** the listing page contains no recognizable article entries
- **THEN** the parser returns an `ItemFetchError` for the source URL

#### Scenario: Every article page fails

- **WHEN** the listing has entries but none of their article pages yields a publication date
- **THEN** the parser returns an `ItemFetchError` for the source URL

### Requirement: Article link accessibility

Each article link in the generated feed SHALL resolve to an accessible page.

#### Scenario: First article link returns HTTP 2xx

- **WHEN** the feed is generated
- **THEN** an HTTP GET to the first article's link returns a status code in the 200-299 range
