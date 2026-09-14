## Why

The repository has no continuous integration. Nothing verifies that a push compiles, is formatted, or passes `go vet`, and master currently fails both `gofmt -l` and `go vet`. Five version tags exist (`1.0.0` through `1.0.4`), but the repository has no GitHub Releases and publishes no container image, so every deployment has to build from source on the target host.

The site parsers also break silently: when a target site changes its markup, the failure surfaces only when a reader notices an empty feed. The `site/` integration tests already detect this, but nothing runs them on a schedule.

## What Changes

- Fix the pre-existing lint failures that would make CI red on its first run: 11 `go vet` unkeyed-field findings across `site/chrb.go`, `site/gamer_forum.go`, `site/hackernews.go`, `site/hackmd.go`, and `site/ptt.go`, plus the `gofmt` violation in `web/main.go`.
- Add `.github/workflows/ci.yml`, triggered on push to master and on pull requests, running `make install`, `gofmt -l .`, and `go vet ./...`. This is the gate for pull requests and the source of the primary README badge.
- Add `.github/workflows/integration.yml`, triggered on a daily schedule and on manual dispatch, running `go test -count=1 ./site/...`. The `site/` tests make real HTTP requests to external sites, so a failure means a target site changed or was unreachable, not that the code regressed. Keeping them off the pull request path stops external outages from blocking merges while preserving them as a canary.
- Add `.github/workflows/release.yml`, triggered on pushing a tag matching `[0-9]+.[0-9]+.[0-9]+`, building and pushing a `linux/amd64` image to `ghcr.io/cldotdev/feedgen` and creating a GitHub Release with generated notes.
- Add two status badges to `README.md`, one per CI workflow.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none)

This change sets `skip_specs: true`. It adds build tooling and documentation without altering what the service does: no endpoint, query parameter, feed output, or parser behavior changes. The `go vet` and `gofmt` fixes are pure refactors, since naming struct fields and reformatting leave the compiled behavior identical.

## Impact

- **Code**: Keyed struct literals in five files under `site/`, and formatting in `web/main.go`. No logic changes.
- **Build**: Three new workflow files under `.github/workflows/`. `Dockerfile` and `compose.yaml` are untouched; the release workflow builds `Dockerfile` as it stands.
- **API**: None.
- **Dependencies**: None added to `go.mod`. The workflows depend on published GitHub Actions, pinned to their major tags.
- **Distribution**: Introduces a published container image at `ghcr.io/cldotdev/feedgen` and GitHub Releases, neither of which the project produces today.
- **Tests**: No test files change. `go test ./site/...` moves from being run by hand to running daily, with `-count=1` so the result cache cannot stand in for reaching the sites.
- **Documentation**: `CLAUDE.md` gains `-count=1` on its Test command for the same reason.
