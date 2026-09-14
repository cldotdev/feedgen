## 1. Parser Rewrite

- [x] 1.1 Replace the `http.PostForm` call in `site/chrb.go` with a GET against `https://chrb.com.tw/rental_properties`, building the query from the parser's own values rather than forwarding the caller's
- [x] 1.2 Add the `sort` map from `newest`, `updated`, `rent_asc`, `rent_desc`, `size_asc`, and `size_desc` onto the site's `q[s]` values, returning a `ParameterValueInvalidError` for anything else and sending no `q[s]` when the parameter is absent
- [x] 1.3 Cut the response down to the region between `id="search_results"` and the closing `</turbo-frame>` before matching entries, so the 精選房源 sidebar cannot reach the feed
- [x] 1.4 Write the entry regex against the card markup, keying on `class="property-link`, the `/rental_properties/<id>` href, the `<h2>` title, and the `/images/icons/` markers rather than on Tailwind utility classes
- [x] 1.5 Extract the tag pills, address, type and layout, size, floor, 屋齡, rent, and date, and build the description from them, omitting the line for any field the card does not carry
- [x] 1.6 Match the tag pills positionally, between the title and the card's first field icon, since their only class is a Tailwind utility
- [x] 1.7 Request through a client with a 30 second timeout and return an error for a status outside the 2XX range
- [x] 1.8 Parse the card date as `2006-01-02` in a fixed UTC+8 zone, following `site/scitechvista.go` rather than `time.LoadLocation`, which returns nil where the image ships without tzdata
- [x] 1.9 Set each item's id and link to `https://chrb.com.tw/rental_properties/<id>`
- [x] 1.10 Return an `ItemFetchError` for the listing URL when no entry matches
- [x] 1.11 Update the doc comment on `ChrbParser` to name the new listing URL

## 2. Tests

- [x] 2.1 Rewrite `site/chrb_test.go` for the parameterless feed, without `t.Parallel()`, asserting a non-empty title and at least one item
- [x] 2.2 Add a case asserting that `sort=newest` yields items in non-increasing date order
- [x] 2.3 Add a case asserting that `sort=rent_asc` yields items in non-decreasing rent order, which the `sort=newest` case cannot show because the site already orders the listing newest first by default
- [x] 2.4 Add a case asserting the `sort` map's six entries without a request, covering the orderings the ordering cases do not exercise
- [x] 2.5 Add a case asserting that an unsupported `sort` value returns `ParameterValueInvalidError`
- [x] 2.6 Add a case asserting that every item link is under `https://chrb.com.tw/rental_properties/` and that the first one returns a 2XX status
- [x] 2.7 Add a case asserting the feed holds no more items than the listing page shows, which fails if the sidebar leaks in
- [x] 2.8 Run `go test -count=1 -run TestChrb ./site/...` and confirm every case passes

## 3. Documentation

- [x] 3.1 Update the 大管家房屋網 entry in `README.md` with the parameterless example and one using `sort`, matching the shape of the neighbouring entries

## 4. Local Verification

- [x] 4.1 Run `make install`, `go vet ./...`, and `gofmt -l .` and confirm all three are clean
- [x] 4.2 Run `go test -count=1 ./site/...` and confirm the chrb cases pass, treating a failure in another parser as unrelated to this change
