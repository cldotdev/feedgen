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

## Image Build

```bash
docker build --no-cache --output type=cacheonly .
```

Run this after changing the Dockerfile, `go.mod`, or `go.sum`. The Dockerfile pins exact Alpine package versions, so the build starts failing once the upstream index drops one, whatever the commit itself changed. `ci.yml` keeps the image off the pull request gate for that reason, which leaves the release workflow as the first thing to build it. A tag is a late place to find a broken build.

`--no-cache` is what exercises the pins: it makes `apk add` resolve against the live index instead of a cached layer. `--output type=cacheonly` builds every layer without exporting an image, leaving nothing behind in `docker images`. The BuildKit cache still grows; `docker builder prune` clears it.
