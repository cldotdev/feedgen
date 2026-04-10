### Requirement: PTT parser uses classic TLS key exchange

The PTT parser SHALL configure its HTTP client to use classic-only TLS key exchange curves (X25519, P-256, P-384), excluding post-quantum MLKEM groups that cause oversized ClientHello messages.

#### Scenario: Successful connection without post-quantum key shares

- **WHEN** the PTT parser sends an HTTPS request to ptt.cc
- **THEN** the TLS ClientHello SHALL contain only classic key exchange groups (X25519, P-256, P-384)
- **AND** the connection SHALL succeed on platforms with middleboxes that reject post-quantum ClientHello
