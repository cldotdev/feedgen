## Context

The PTT parser (`site/ptt.go`) uses a default `http.Client{}` with no TLS configuration. Go 1.26 introduced post-quantum MLKEM key exchange groups as defaults in the TLS ClientHello (X25519MLKEM768, SecP256r1MLKEM768, SecP384r1MLKEM1024). Each MLKEM key share adds ~1.2 KB, causing the ClientHello to exceed what certain network middleboxes can buffer. On Railway.com (all tested regions: Amsterdam, Singapore, California), this results in TCP connection resets when connecting to ptt.cc. Go 1.25.3 (v1.0.1) works because it sends fewer or no post-quantum key shares.

Both `GetFeed` and `GetFeedItem` originally created separate bare `http.Client{}` instances with no TLS configuration or timeout.

## Goals / Non-Goals

**Goals:**

- Restore PTT connectivity on hosting platforms with middleboxes that reject oversized TLS ClientHello messages.
- Minimal, targeted change limited to the PTT parser.

**Non-Goals:**

- Applying the same fix to other site parsers (can be done separately if needed).
- Creating a shared HTTP client abstraction across all parsers.
- Disabling post-quantum TLS globally via GODEBUG.

## Decisions

### Use explicit CurvePreferences over GODEBUG environment variable

**Choice**: Set `CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384}` on a shared `http.Transport` in `site/ptt.go`.

**Alternatives considered**:

- `GODEBUG=tlsmlkem=0` in Dockerfile: Disables post-quantum globally, affecting all outbound connections. Less visible, environment-dependent, and overly broad.
- Per-method inline Transport: Duplicates the Transport config in two methods. Go docs recommend reusing Transports.

**Rationale**: A package-level `http.Transport` variable is explicit, self-documenting, reusable across both methods, and scoped to PTT only. A shared `pttClient` wraps the transport with a 30-second timeout and is used directly by both methods.

### Keep classic curves only (X25519, P-256, P-384)

These three curves provide strong security and universal middlebox compatibility. P-521 is omitted as it is rarely negotiated and adds no practical value here.

### Set explicit MinVersion and Timeout

`MinVersion: tls.VersionTLS12` makes the Go default explicit, satisfying static analysis (semgrep). `Timeout: 30 * time.Second` prevents indefinite hangs on upstream connections, consistent with the UDN parser.

## Risks / Trade-offs

- **[No post-quantum protection for PTT connections]** Classic curves remain secure against current threats. Post-quantum protection can be re-enabled when middlebox compatibility improves. This is a pragmatic trade-off.
- **[Other parsers may face the same issue]** If other sites also fail due to post-quantum ClientHello, each parser would need a similar fix. Accepted for now to keep the change minimal.
