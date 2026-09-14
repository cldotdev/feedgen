## Purpose

Turns the 大管家房屋網 rental listings into an Atom feed, so readers can subscribe to new rental properties instead of reloading the search page.

## Requirements

### Requirement: Feed generation for the rental listing page

The system SHALL expose a `GET /chrb` endpoint that returns an Atom feed of the rental properties on the first page of `https://chrb.com.tw/rental_properties`.

#### Scenario: Successful feed without parameters

- **WHEN** a request is made to `GET /chrb`
- **THEN** the response is an Atom feed titled "大管家房屋網"
- **THEN** the feed link points at `https://chrb.com.tw/rental_properties`
- **THEN** the feed contains the properties listed in the site's default ordering

#### Scenario: Sidebar properties are excluded

- **WHEN** the listing page renders its 精選房源 sidebar alongside the search results
- **THEN** the feed contains only the properties inside the page's `search_results` region
- **AND** no sidebar property appears as an item

#### Scenario: No properties found on the page

- **WHEN** the listing page returns no property matching the expected markup
- **THEN** the parser returns an `ItemFetchError` for the listing URL

#### Scenario: Listing page is unavailable

- **WHEN** the listing page answers with a status outside the 2XX range
- **THEN** the parser returns an error naming that status and the URL it requested

### Requirement: Sort parameter

The parser SHALL accept an optional `sort` query parameter, map it to the ordering the site accepts, and reject any other value.

#### Scenario: Newest properties first

- **WHEN** a request is made to `GET /chrb?sort=newest`
- **THEN** the parser requests the listing ordered by `published_at desc`
- **THEN** the feed lists the most recently published properties first

#### Scenario: Recently updated properties first

- **WHEN** a request is made to `GET /chrb?sort=updated`
- **THEN** the parser requests the listing ordered by `updated_at desc`

#### Scenario: Ordering by rent and by size

- **WHEN** a request is made with `sort` set to `rent_asc`, `rent_desc`, `size_asc`, or `size_desc`
- **THEN** the parser requests the listing ordered by the site's corresponding `rent` or `area_ping` ordering

#### Scenario: Cheapest properties first

- **WHEN** a request is made to `GET /chrb?sort=rent_asc`
- **THEN** the feed lists the properties by rent, from lowest to highest

#### Scenario: Omitted sort parameter

- **WHEN** a request is made to `GET /chrb` without a `sort` parameter
- **THEN** the parser requests the listing without an ordering parameter
- **AND** no error is returned

#### Scenario: Invalid sort parameter

- **WHEN** a request is made with an unsupported `sort` value, such as `sort=cheapest`
- **THEN** the parser returns a `ParameterValueInvalidError` for "sort"

### Requirement: Item content

Each feed item SHALL identify one property by its absolute URL and carry the fields the listing card shows.

#### Scenario: Item identity and link

- **WHEN** the listing card links to `/rental_properties/0102415`
- **THEN** the item id and link are both `https://chrb.com.tw/rental_properties/0102415`

#### Scenario: Item title and description

- **WHEN** a property is converted to a feed item
- **THEN** the item title is the property title shown on the card
- **AND** the description carries the address, the property type and layout, the size in 坪, the floor, the rent, and the tags shown on the card

#### Scenario: Fields the card omits

- **WHEN** a card shows no layout, no 屋齡, or no tag
- **THEN** the description omits that line and carries the remaining fields
- **AND** the omitted field is not filled in from the card that follows

#### Scenario: Property age

- **WHEN** a card shows a 屋齡 such as `3年`
- **THEN** the description carries it

#### Scenario: Item publication date

- **WHEN** the listing card shows a date such as `2026-09-14`
- **THEN** the item's created time is that date in UTC+8
