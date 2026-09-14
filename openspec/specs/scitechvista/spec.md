## Purpose

Turns the 科技大觀園 article listings into Atom feeds, so readers can subscribe to the site's popular-science articles instead of polling the pages by hand.

## Requirements

### Requirement: Feed generation for article sections

The system SHALL expose a `GET /scitechvista` endpoint that accepts a `section` query parameter and returns an Atom feed of the articles listed in that section of 科技大觀園.

#### Scenario: Successful feed with section=new

- **WHEN** a request is made to `GET /scitechvista?section=new`
- **THEN** the response is an Atom feed with title "最新文章 | 科技大觀園"
- **THEN** the feed link points at `https://scitechvista.nat.gov.tw/Article/C000003/new`
- **THEN** the feed contains the articles listed on that page

#### Scenario: Successful feed with section=hot

- **WHEN** a request is made to `GET /scitechvista?section=hot`
- **THEN** the response is an Atom feed with title "熱門文章 | 科技大觀園"

#### Scenario: Successful feed with section=featured

- **WHEN** a request is made to `GET /scitechvista?section=featured`
- **THEN** the response is an Atom feed with title "精選文章 | 科技大觀園"

### Requirement: Section parameter validation

The parser SHALL validate the `section` query parameter and return an error for missing or unsupported values.

#### Scenario: Missing section parameter

- **WHEN** a request is made to `GET /scitechvista` without a `section` parameter
- **THEN** the parser returns a `ParameterNotFoundError` for "section"

#### Scenario: Invalid section parameter

- **WHEN** a request is made with an unsupported section value (e.g., `section=nonexistent`)
- **THEN** the parser returns a `ParameterValueInvalidError` for "section"

### Requirement: Feed item content

Each feed item SHALL carry the article's title, absolute link, summary, author, and publication date as shown on the listing page.

#### Scenario: Item fields are populated

- **WHEN** a feed is generated for any supported section
- **THEN** every item has a non-empty title, a non-empty link, and a non-zero created timestamp
- **THEN** every item link is an absolute URL under `https://scitechvista.nat.gov.tw/`
- **THEN** every item id equals its link

#### Scenario: Title and summary are decoded

- **WHEN** a listing entry contains HTML character references such as `&amp;` or `&#39;`
- **THEN** the item title and description contain the decoded characters, not the references

#### Scenario: Article without a listed author

- **WHEN** a listing entry shows no author
- **THEN** the item is still included in the feed with an empty author

### Requirement: Republic of China calendar date conversion

The parser SHALL convert the listing's ROC-calendar dates into Gregorian timestamps in UTC+8, without depending on the system time zone database.

#### Scenario: ROC year converted to Gregorian year

- **WHEN** a listing entry shows the date `115/08/31`
- **THEN** the item's created timestamp is 2026-08-31 in UTC+8

### Requirement: Listing scope

The feed SHALL contain only the articles of the requested section, excluding the recommended-articles sidebar that the same page renders.

#### Scenario: Sidebar articles excluded

- **WHEN** a feed is generated for any supported section
- **THEN** the feed contains only entries from the section's own list
- **THEN** items are ordered from newest to oldest by created timestamp

### Requirement: Upstream failure reporting

The parser SHALL report an error instead of returning an empty or partial feed when the listing cannot be retrieved or contains no recognizable entries.

#### Scenario: Listing page returns a non-2xx status

- **WHEN** the upstream listing request returns a status outside the 200-299 range
- **THEN** the parser returns an error naming the source URL

#### Scenario: No entries match the expected listing structure

- **WHEN** the listing page contains no recognizable article entries
- **THEN** the parser returns an `ItemFetchError` for the source URL

### Requirement: Article link accessibility

Each article link in the generated feed SHALL resolve to an accessible page.

#### Scenario: First article link returns HTTP 2xx

- **WHEN** the feed is generated with `section=new`
- **THEN** an HTTP GET to the first article's link returns a status code in the 200-299 range
