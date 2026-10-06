## Purpose

Keeps one unreachable PTT article from deciding the outcome of a whole board feed, and keeps an unusable listing from passing as an empty board.

## ADDED Requirements

### Requirement: Per-article failure tolerance

The PTT parser SHALL leave out an article whose page cannot be fetched or is not found, and still return the remaining articles, wherever the failed article sits in the listing.

#### Scenario: Last listed article fails

- **WHEN** the last article in a PTT listing returns a not-found page and the others succeed
- **THEN** the parser returns a feed of the other articles and no error

#### Scenario: A middle article fails

- **WHEN** an article other than the last one fails and the others succeed
- **THEN** the parser returns a feed of the other articles and no error

### Requirement: Upstream failure reporting

The PTT parser SHALL report an `ItemFetchError` for the listing URL instead of returning an empty feed when the listing page is not a recognizable PTT list or none of its articles can be turned into an item. A recognizable list with no articles, such as a search without results, SHALL yield an empty feed.

#### Scenario: Search without results

- **WHEN** a PTT search page renders the article list container with no article in it
- **THEN** the parser returns an empty feed and no error

#### Scenario: Listing page is not a recognizable list

- **WHEN** the PTT page has no article list container, for example because the board does not exist or the markup changed
- **THEN** the parser returns an `ItemFetchError` for the listing URL

#### Scenario: Every article fails

- **WHEN** the listing has article links but none of their pages can be turned into an item
- **THEN** the parser returns an `ItemFetchError` for the listing URL
