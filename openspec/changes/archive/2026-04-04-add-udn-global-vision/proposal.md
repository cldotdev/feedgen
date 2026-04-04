## Why

feedgen currently lacks support for 轉角國際 (global.udn.com), a popular international news site under the UDN family. Adding it as a feed source enables users to subscribe to its "深度專欄" (in-depth column) articles via Atom feed. Additionally, the existing `udn_game` parser shares the same UDN JSON API structure, creating an opportunity to extract shared parsing logic into a reusable `parser` package.

## What Changes

- Create a new `parser/` package with shared UDN types and article-fetching logic extracted from `site/udn_game.go`.
- Refactor `site/udn_game.go` to use the shared `parser` package instead of inline types and fetch logic.
- Add `site/udn_global_vision.go` implementing a new parser for 轉角國際 深度專欄.
- Add `site/udn_global_vision_test.go` with integration tests.
- Register a new route `GET /udn_global_vision?tag=in-depth-column` in `web/main.go`.
- Update `README.md` with the new site and example URL.

## Capabilities

### New Capabilities

- `udn-shared-parser`: Shared UDN JSON API types and article-fetching logic in `parser/udn.go`, reusable across UDN family site parsers.
- `udn-global-vision`: Site parser for 轉角國際 深度專欄 with tag-based article fetching via `GET /udn_global_vision?tag=in-depth-column`.

### Modified Capabilities

(none)

## Impact

- **Code**: New `parser/` package; `site/udn_game.go` refactored to depend on it; new `site/udn_global_vision.go`.
- **API**: New endpoint `GET /udn_global_vision?tag=in-depth-column`. No changes to existing endpoints.
- **Dependencies**: No new external dependencies. Uses existing `bytedance/sonic` for JSON and `gorilla/feeds` for feed generation. Upgraded `quic-go` and `x/crypto` transitive dependencies to resolve CVEs.
- **Tests**: New integration tests that make real HTTP requests to `global.udn.com`.
