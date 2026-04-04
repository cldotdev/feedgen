## Context

feedgen generates Atom feeds by scraping/fetching various websites. Each site has a parser in `site/` that implements the `feedgen.Parser` interface. The existing `site/udn_game.go` fetches articles from UDN's JSON API, unmarshals them into local structs, and converts them to feed items. The new 轉角國際 site uses the same UDN JSON API format, making shared extraction beneficial.

Current state of `site/udn_game.go`:

- Defines its own JSON structs (`UdnGameData`, `UdnGameArticles`, etc.)
- Handles HTTP fetch and JSON unmarshal inline
- Handles 10-digit/13-digit Unix timestamp conversion

## Goals / Non-Goals

**Goals:**

- Extract shared UDN JSON types and fetch logic into `parser/udn.go`
- Refactor `site/udn_game.go` to use the shared parser package
- Add `site/udn_global_vision.go` for 轉角國際 深度專欄
- Maintain backward compatibility for the `/udn_game` endpoint

**Non-Goals:**

- Supporting additional 轉角國際 tags beyond `in-depth-column` in this change
- Using the `cate` (category) field from the API response
- Pagination support (first page of 20 articles is sufficient)
- Refactoring non-UDN parsers

## Decisions

### 1. Shared package location: `parser/udn.go`

Place shared UDN logic in a top-level `parser/` package rather than `site/udn/` sub-package.

**Rationale**: Keeps `site/` flat with one file per site parser (existing convention). The `parser/` package serves as a reusable layer that `site/` depends on. Dependency direction is clean: `site` -> `parser` -> `feedgen` (root).

**Alternatives considered**:

- `site/udn/` sub-package: Breaks the flat `site/` convention and changes import paths for existing parsers.
- Shared code in `site/udn_common.go`: Same package, but pollutes `site` namespace with UDN-specific helpers.

### 2. Exported `FetchArticles` function

Expose `parser.FetchArticles(rawLink string) ([]parser.Article, error)` that handles HTTP GET, JSON unmarshal, and timestamp conversion into `time.Time`.

**Rationale**: Both `udn_game` and `udn_global_vision` share the exact same fetch-unmarshal-convert flow. Returning articles with parsed `time.Time` (instead of raw timestamps) keeps conversion logic centralized.

**Alternatives considered**:

- Return raw `[]Article` with unparsed timestamps: Each site parser would duplicate timestamp logic.

### 3. Tag parameter mapping via `map[string]tagConfig`

Map English tag slugs to Chinese tag names and feed metadata using a static map:

```go
var globalVisionTags = map[string]tagConfig{
    "in-depth-column": {name: "深度專欄", feedTitle: "深度專欄 | 轉角國際"},
}
```

**Rationale**: Provides user-friendly English query parameter values while mapping to the Chinese tag names required by the UDN API. Easy to extend with more tags later.

**Alternatives considered**:

- Accept raw Chinese tag names as query params: Requires URL encoding, poor UX.
- Hardcode a single tag with no parameter: Less flexible, harder to extend.

### 4. Article timestamp returned as `time.Time` in shared struct

The shared `Article` struct stores `Created time.Time` (already parsed) rather than the raw `Time` sub-struct.

**Rationale**: Every consumer needs `time.Time` for feed items. Centralizing the 10-digit/13-digit detection avoids duplication and keeps site parsers focused on feed construction.

### 5. `AddArticlesToFeed` shared helper

Extract the article-to-feed-item conversion loop into `parser.AddArticlesToFeed(feed, articles)` to eliminate identical loops in both `udn_game.go` and `udn_global_vision.go`.

**Rationale**: Both site parsers had a character-for-character identical loop mapping `parser.Article` fields to `feeds.Item` fields. Centralizing this removes duplication and ensures any future UDN parser gets the same mapping for free.

### 6. Defensive HTTP handling in `FetchArticles`

Added a 30-second HTTP client timeout and a status code check (non-2xx returns an error) to `FetchArticles`.

**Rationale**: Prevents indefinite blocking on slow upstreams and provides clear error messages instead of confusing JSON parse failures on error responses.

### 7. Zero-timestamp article skipping

Articles with `timestamp == 0` are silently skipped rather than causing the entire feed to fail.

**Rationale**: A single article with a missing timestamp (zero value) should not poison the entire feed. Skipping preserves the remaining valid articles.

## Risks / Trade-offs

- **[UDN API change]** Both parsers depend on the same undocumented UDN JSON API. If the API changes, both break simultaneously. -> Mitigation: Integration tests catch this; shared code means a single fix point.
- **[Tag mapping maintenance]** Adding new tags requires code changes to the static map. -> Mitigation: Acceptable for now; the map is trivial to extend.
- **[Refactoring risk]** Changing `udn_game.go` could break existing behavior. -> Mitigation: Existing tests validate the refactored parser produces valid feeds.
