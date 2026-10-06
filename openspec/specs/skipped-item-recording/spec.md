## Purpose

Leaves a durable, deduplicated trace of every item a parser drops from a feed, so an operator can tell which articles fail, why, and how often, without the failure changing what the feed returns.

## Requirements

### Requirement: Skipped items are recorded

The system SHALL record each item that a parser leaves out of a feed because the item's page could not be fetched or its publication date could not be read. This covers the PTT, 想想論壇, 科技大觀園, 大管家房屋網, and UDN parsers.

#### Scenario: Article page fetch fails

- **WHEN** a parser that fetches each article's page skips an article because the request failed, returned a non-2xx status, or returned a not-found page
- **THEN** a record exists for that parser and article URL with the matching failure reason

#### Scenario: Listing entry has no readable date

- **WHEN** a parser skips a listing entry because it has no parseable publication date or a zero timestamp
- **THEN** a record exists for that parser and the entry's article URL with the reason `no_date`

### Requirement: Record content

Each record SHALL carry the parser name, the article URL, a failure reason from a fixed set, the URL of the listing being turned into a feed, the latest error message, the first and last time the failure was seen, and how many times it was seen. Records SHALL NOT contain page bodies.

#### Scenario: Failure reason categories

- **WHEN** an item is recorded
- **THEN** its reason is one of `network`, `http_status`, `not_found`, or `no_date`

### Requirement: Deduplicated records

The system SHALL keep one record per parser, article URL, and reason. A repeat of an existing key SHALL update that record's last-seen time, latest message, and listing URL, and increment its count, while keeping its first-seen time.

#### Scenario: Same failure seen twice

- **WHEN** the same parser skips the same article for the same reason on two requests
- **THEN** one record exists for that key
- **THEN** its count is 2, its first-seen time is the first occurrence, and its last-seen time is the second occurrence

#### Scenario: Same article fails for a different reason

- **WHEN** an article recorded with reason `network` is later skipped with reason `http_status`
- **THEN** two records exist for that article, one per reason

### Requirement: Storage location

The server SHALL store the records in the SQLite database `log/fetch_error.db`, relative to its working directory, creating the file and its table when they do not exist.

#### Scenario: First start without a database

- **WHEN** the server starts and `log/fetch_error.db` does not exist
- **THEN** the file is created with an empty table of records

### Requirement: Recording never affects feed output

Failing to open the database or to write a record SHALL NOT change the response to a feed request and SHALL NOT stop the server. The failure SHALL be written to the server's text log instead.

#### Scenario: Database cannot be opened at startup

- **WHEN** the server cannot open `log/fetch_error.db`
- **THEN** the server still starts and serves feeds
- **THEN** skipped items are written to the server's text log instead of the database

#### Scenario: A record cannot be written

- **WHEN** writing a record fails
- **THEN** the feed response is the same as if the write had succeeded
- **THEN** the write failure is written to the server's text log
