## 1. Shared UDN Parser Package

- [x] 1.1 Create `parser/udn.go` with shared types (`Article`, internal JSON structs) and exported `FetchArticles` function
- [x] 1.2 Refactor `site/udn_game.go` to use `parser.FetchArticles` and remove inline types/fetch logic
- [x] 1.3 Run existing `udn_game` tests to verify refactor

## 2. Global Vision Parser

- [x] 2.1 Create `site/udn_global_vision.go` with `UdnGlobalVisionParser` implementing `feedgen.Parser`
- [x] 2.2 Create `site/udn_global_vision_test.go` with integration tests for feed generation, parameter validation, and article link checks

## 3. Wiring and Documentation

- [x] 3.1 Register `GET /udn_global_vision` route in `web/main.go`
- [x] 3.2 Add 轉角國際 entry to `README.md`
- [x] 3.3 Run all tests (`go test ./site/...`) to verify everything works

## 4. Audit Fixes

- [x] 4.1 Upgrade `quic-go` (v0.55→v0.59) and `x/crypto` (v0.43→v0.49) to resolve CVEs
- [x] 4.2 Add HTTP client timeout (30s) to `FetchArticles`
- [x] 4.3 Add HTTP status code check (non-2xx returns error) to `FetchArticles`
- [x] 4.4 Skip articles with zero timestamp instead of failing the entire feed

## 5. Simplify

- [x] 5.1 Extract `parser.AddArticlesToFeed` helper to eliminate duplicate loops in `udn_game.go` and `udn_global_vision.go`
