## Context

See proposal.md - Why.

What the new site looks like, all verified against the live pages:

- `https://www.chrb.com.tw/tenement.php` returns 301 to `https://chrb.com.tw/rental_properties`, and `tenement_ajax.php` is gone from both hosts.
- The listing is a Rails application served through Turbo. The search results sit inside `<turbo-frame id="search_results">`; the 精選房源 sidebar sits outside it and reuses the same `property-link` anchor class.
- A page holds 12 properties, and the listing reports 1089 properties over 91 pages.
- Sorting is a `q[s]` query parameter. The page's own `<select>` offers `published_at desc`, `updated_at desc`, `rent asc`, `rent desc`, `area_ping desc`, and `area_ping asc`.
- Each card carries a `/rental_properties/<id>` link, an `<h2>` title, tag pills, property type and layout, size in 坪, floor as `6F/6F`, an address, a rent, and a date such as `2026-09-14`. 屋齡 appears as `3年` on the cards that have one, a 店面 card can omit the layout, and a card can carry no tag pill.
- Omitting the ordering parameter returns the same listing as `published_at desc`, so the site's default ordering is newest first.
- The site rejects some clients by user agent: `Python-urllib/3.14` gets 403 where `Go-http-client/1.1`, an empty agent, and a browser agent all get 200. Go's default client needs no header of its own.

## Goals / Non-Goals

**Goals:**

- Restore `GET /chrb` with the parameterless form the README documents.
- Keep the feed free of anything the site did not put in its search results.

**Non-Goals:**

- Paginating. The parser reads page one, as every other parser here reads one listing.
- Fetching each property's detail page. Everything the description needs is on the card.
- Preserving the old item ids. The old URLs no longer resolve, so there is nothing to preserve them against.

## Decisions

### Isolate the `search_results` frame before matching entries

The parser cuts the response down to the region between `id="search_results"` and the closing `</turbo-frame>`, then matches property cards inside it.

Matching `property-link` across the whole document was the simpler alternative and is wrong: the 精選房源 sidebar uses the same class, so five sidebar cards would join every feed, and they are neither new nor ordered by the requested sort. Those cards also use a different inner layout, `<h3>` with no date, so they would parse into items missing the fields everything else has.

This repeats the decision `site/scitechvista.go` made when it anchored on the listing's own class to keep the 推薦文章 sidebar out of the feed.

### Anchor the regex on semantic attributes, not Tailwind classes

Entry matching keys on `id="search_results"`, `class="property-link`, `href="/rental_properties/<id>"`, and the `/images/icons/<name>.svg` markers that label the type, size, age, and floor fields. It does not key on utility classes such as `text-[#DC4548]` or `line-clamp-2`.

Utility classes encode presentation and change whenever the design does; a color literal in a regex is a parser that breaks on a restyle. Two fields carry no marker of their own. The rent is matched by its neighbouring `/月` text rather than by its color class. The tag pills have no class but a utility one, so they are matched positionally instead, in the region between `</h2>` and the card's first `/images/icons/` image, where nothing else sits.

### `sort` maps names onto the site's `q[s]` values

The parser accepts `newest`, `updated`, `rent_asc`, `rent_desc`, `size_asc`, and `size_desc`, mapping them through a static map to `published_at desc`, `updated_at desc`, `rent asc`, `rent desc`, `area_ping asc`, and `area_ping desc`. Any other value is a `ParameterValueInvalidError`. Omitting the parameter sends no `q[s]` and takes the site's default ordering.

Passing `q[s]` values through verbatim was the alternative. It would put a Rails Ransack sort expression in a feed URL, tie the public API to the site's internal parameter names, and leave the endpoint accepting values the site may reject. `site/scitechvista.go` already maps a `section` name onto a site-specific path this way, and `site/udn_game.go` does the same for `by`.

`sort` is optional rather than required because the README documents a bare `/chrb` and existing subscribers use it.

### GET replaces `PostForm`, and the query is no longer passed through

The old parser handed the entire `url.Values` to `http.PostForm`, so any parameter a caller invented reached the site untouched and no validation existed. The new parser builds its own query string from the mapped `sort` value alone.

This is why validation appears in a parser that never had any: the endpoint now has a parameter of its own to get wrong.

### Request `chrb.com.tw` directly

The parser uses the apex host rather than `www`, which 301s to it. One less round trip, and the feed link matches the URL a reader lands on.

### Absolute URL as both id and link

An item's id and link are both `https://chrb.com.tw/rental_properties/<id>`, matching what `site/scitechvista.go` does.

The property id is stable across the rebuild: `0102415` appears in both the old `<id>.html` and the new path. The full URL still changes, so every currently subscribed item resurfaces once as new. That is unavoidable while keeping the id equal to a URL that resolves.

### Fields the description carries

In: address, property type and layout, size, floor, 屋齡, rent, and tag pills. Out: 更新日期, which the new card does not show.

The layout, 屋齡, and the tag pills are each absent from some cards, so the description omits the line rather than emitting an empty one. Matching each field against a single captured card, rather than against the whole page, is what keeps an absent field from absorbing the next card's value.

The regex requires the date, so a card that ever renders without one is skipped rather than emitted with a zero time. No such card was observed across 57 cards on five pages, so this is a consequence of the match rather than a designed behavior; `site/scitechvista.go` skips undated entries for the same reason.

### `rent_asc` is the ordering the tests can check

The listing's default ordering is already newest first, so a feed requested with `sort=newest` is indistinguishable from one requested without a parameter, and a date-ordering assertion over it passes even if `q[s]` never leaves the parser. Rent discriminates: the cheapest properties never head the default page, so asserting non-decreasing rent over `sort=rent_asc` fails the moment the parameter stops arriving.

The remaining four orderings are covered by asserting the map itself rather than by four more requests, because they exercise the mapping, not the site.

### Tests run serially

`site/chrb_test.go` drops the `t.Parallel()` the current file has. Commit `e8f7cc5` removed it from the PTT tests because real HTTP requests fired in parallel invite rate limiting and flaky results, and the reasoning carries over unchanged.

## Risks / Trade-offs

- Item ids change, so every subscriber sees the current listings once as new → unavoidable; the old URLs no longer resolve, and the alternative is an id that points nowhere.
- The site was rebuilt once and can be restyled again, breaking the regex → mitigated by keying on semantic attributes rather than presentation classes, and detected by the daily integration workflow rather than by a reader noticing an empty feed.
- A restyle could also drop a card's date, which the parser requires → the card is skipped rather than dated zero, so the feed shrinks instead of carrying a wrong date.
- The site returns 403 to at least one common client by user agent → Go's default agent is accepted today, and adding a browser agent pre-emptively would be guessing at a rule that is not published. If it starts returning 403, that is the moment to set one.
- Reading only page one caps the feed at 12 properties → matches every other parser here, and a reader wanting more is asking for a different feature than a change feed.
