# udn-global-vision

Site parser for 轉角國際 深度專欄 with tag-based article fetching via
`GET /udn_global_vision?tag=in-depth-column`.

## Requirements

### Requirement: Feed generation for in-depth-column tag

The system SHALL expose a `GET /udn_global_vision` endpoint that accepts a
`tag` query parameter and returns an Atom feed of articles from 轉角國際.

#### Scenario: Successful feed with tag=in-depth-column

- **WHEN** a request is made to `GET /udn_global_vision?tag=in-depth-column`
- **THEN** the response is an Atom feed with title "深度專欄 | 轉角國際"
- **THEN** the feed contains articles fetched from the UDN Global Vision
  API for the "深度專欄" tag
- **THEN** each feed item has a non-empty title, link, description, author,
  and created timestamp

### Requirement: Tag parameter validation

The parser SHALL validate the `tag` query parameter and return an error for
missing or unsupported values.

#### Scenario: Missing tag parameter

- **WHEN** a request is made to `GET /udn_global_vision` without a `tag`
  parameter
- **THEN** the parser returns a `ParameterNotFoundError` for "tag"

#### Scenario: Invalid tag parameter

- **WHEN** a request is made with an unsupported tag value
  (e.g., `tag=nonexistent`)
- **THEN** the parser returns a `ParameterValueInvalidError` for "tag"

### Requirement: Article link accessibility

Each article link in the generated feed SHALL resolve to an accessible page.

#### Scenario: First article link returns HTTP 2xx

- **WHEN** the feed is generated with `tag=in-depth-column`
- **THEN** an HTTP GET to the first article's link returns a status code in
  the 200-299 range
