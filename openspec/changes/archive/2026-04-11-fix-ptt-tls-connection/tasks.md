## 1. Implementation

- [x] 1.1 Add `crypto/tls` import and a package-level `pttTransport` variable with classic-only `CurvePreferences` in `site/ptt.go`
- [x] 1.2 Add a package-level `pttClient` variable wrapping `pttTransport` with a 30-second timeout
- [x] 1.3 Update `GetFeed` to use `pttClient` instead of `&http.Client{}`
- [x] 1.4 Update `GetFeedItem` to use `pttClient` instead of `&http.Client{}`

## 2. Hardening (from audit)

- [x] 2.1 Add `MinVersion: tls.VersionTLS12` to the TLS config to satisfy semgrep
- [x] 2.2 Add `Timeout: 30 * time.Second` to the HTTP client to prevent indefinite hangs

## 3. Verification

- [x] 3.1 Run `go build ./...` to confirm compilation
- [x] 3.2 Run `go test ./site/... -run TestPtt` to confirm PTT integration tests pass
