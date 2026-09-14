## Why

大管家房屋網 rebuilt its site. `https://www.chrb.com.tw/tenement_ajax.php`, the endpoint `ChrbParser` posts to, now redirects to `chrb.com.tw` and returns 404, so `GET /chrb` produces an `ItemFetchError` and `TestChrbParser_GetFeed` fails. The listings themselves are intact: the new `/rental_properties` page is server-rendered and carries the same fields the feed has always shown, so the parser can be repaired rather than retired.

## What Changes

- Rewrite `site/chrb.go` against `https://chrb.com.tw/rental_properties`, replacing the `http.PostForm` call with a GET.
- Scope entry matching to the `<turbo-frame id="search_results">` region. The 精選房源 sidebar reuses the same `property-link` class, and matching on that class alone would mix five sidebar cards into every feed.
- Accept an optional `sort` query parameter mapping to the site's own `q[s]` values, and reject values outside that set. The parameterless `GET /chrb` in the README keeps working and returns the site's default ordering.
- Replace the description fields that the new listing no longer carries. 更新日期 is gone and the tag pills (有電梯, 租金補貼, and similar) take its place. 屋齡 remains on the cards that show one, so the description keeps it as an optional field.
- Rewrite `site/chrb_test.go` for the new markup, without `t.Parallel()`, following `e8f7cc5`.
- Update the 大管家房屋網 entry in `README.md` with an example for the new parameter.

## Capabilities

### New Capabilities

- `chrb`: Site parser for 大管家房屋網 rental listings, exposing `GET /chrb` with an optional `sort` parameter as an Atom feed.

### Modified Capabilities

(none)

`GET /chrb` predates this project's adoption of OpenSpec and has no spec under `openspec/specs/`, so its requirements are written as a new capability rather than as a delta against one that was never recorded.

## Impact

- **Code**: `site/chrb.go` and `site/chrb_test.go` are rewritten. The route registration in `web/main.go` is unchanged.
- **API**: `GET /chrb` keeps its path and its parameterless form. It gains an optional `sort` parameter and, with it, the first validation error this endpoint can return.
- **Feed content**: **BREAKING** for subscribers. Item ids move from `https://www.chrb.com.tw/<id>.html` to `https://chrb.com.tw/rental_properties/<id>`, so every existing item reappears as new once. The description gains tag pills and loses 更新日期.
- **Dependencies**: None. The page is server-rendered HTML, parsed with `regexp` like every other parser here.
- **Tests**: `site/chrb_test.go` makes real HTTP requests to `chrb.com.tw`, as before.
