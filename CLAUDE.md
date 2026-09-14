# Development

## Build

```bash
make install
```

## Test

```bash
go test -count=1 ./site/...
```

Tests are integration tests that make real HTTP requests to external sites. `-count=1` bypasses the test result cache, which otherwise reports a pass without reaching any site.
