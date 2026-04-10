## Why

Go 1.26 sends post-quantum MLKEM key shares (X25519MLKEM768, SecP256r1MLKEM768, SecP384r1MLKEM1024) by default in the TLS ClientHello, making the handshake message significantly larger (~1.2 KB per key share). Network middleboxes on hosting platforms such as Railway.com cannot handle the oversized ClientHello and reset the TCP connection, causing "read: connection reset by peer" errors when connecting to PTT (ptt.cc). Rolling back to v1.0.1 (Go 1.25.3, which sends fewer/no post-quantum key shares) resolves the issue across all tested regions.

## What Changes

- Configure the PTT parser's HTTP client to use classic-only TLS key exchange curves (X25519, P-256, P-384), excluding post-quantum MLKEM groups that inflate the ClientHello.
- No functional change to feed parsing logic; only the TLS handshake parameters are affected.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none - this is an implementation-level fix; no spec-level behavior changes)

## Impact

- **Code**: `site/ptt.go` - a shared `pttClient` (backed by a custom `http.Transport` with explicit `CurvePreferences`, `MinVersion`, and `Timeout`) replaces the bare `http.Client{}` in both `GetFeed` and `GetFeedItem`.
- **Dependencies**: No new dependencies; `crypto/tls` is in the standard library.
- **Compatibility**: The fix restores compatibility with middleboxes that cannot handle post-quantum TLS. Classic curves remain secure and widely supported.
