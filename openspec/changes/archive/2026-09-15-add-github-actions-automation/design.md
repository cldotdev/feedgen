## Context

See proposal.md - Why.

Three properties of the repository shape every decision below:

- Every test file lives under `site/` and makes real HTTP requests to the site it parses. There are no unit tests and no fixtures, so no subset of the suite runs offline.
- `Dockerfile` pins exact Alpine package versions (`ca-certificates=20260611-r0`, `curl=8.22.0-r0`). Alpine drops old package revisions from its repositories, so an unchanged `Dockerfile` stops building whenever the upstream index moves. Commit `6234c7b` is the most recent repair of this kind.
- The deployable unit is a container. `compose.yaml` uses `build: .`, and the repository ships no binary to anyone.

## Goals / Non-Goals

**Goals:**

- A badge whose red state always means the code is broken.
- A separate, scheduled signal for the parsers, whose red state means a target site changed.
- One tag push produces both a published image and a GitHub Release, with no manual step in between.

**Non-Goals:**

- Running the integration tests on pull requests, or making them hermetic with recorded fixtures. Recording fixtures would freeze the markup the tests exist to watch.
- Publishing binaries or Go module artifacts.
- Changing how the service is deployed. Pointing `compose.yaml` at the published image is a separate change.
- Automating the Alpine package pins in `Dockerfile`.

## Decisions

### Two CI workflows instead of one

`ci.yml` runs `make install`, `gofmt -l .`, and `go vet ./...` on push to master and on pull requests. `integration.yml` runs `go test -count=1 ./site/...` on a daily schedule and on `workflow_dispatch`.

A single workflow running everything was the obvious alternative, and having no other tests invites it. It was rejected because the two signals answer different questions. `go test ./site/...` fails when 批踢踢 is slow, when 巴哈姆特 rate-limits a datacenter address, or when a site is down for maintenance, none of which is a reason to block a merge. Merging the two makes the badge mean "the code compiles and every external site was healthy at that moment", which is a claim nobody can act on.

Rejecting the tests entirely was the other alternative. It discards their real value: they are the only mechanism that detects a parser silently breaking. A daily schedule keeps that value without letting an outage block a merge.

### The gate runs the project's build command

`Makefile` already defines the build as `go build -o bin/webserver web/main.go`, and `CLAUDE.md` names `make install` as the project's build command. The gate invokes that rather than a generic `go build ./...`, so CI compiles what the project ships instead of a parallel formulation that happens to cover the same packages today.

Nothing is lost by the narrower command, because `go vet ./...` type-checks every package in the module, test files included, and reports a package that fails to compile. `go build ./...` would sit between the two, adding only packages that no import path reaches, of which there are none.

Setting `CGO_ENABLED=0` in the gate was considered and rejected. `Dockerfile` already sets it for the artifact that ships, and the release workflow builds that `Dockerfile`, so the static-link path is covered where it matters. A third copy of the build configuration in a workflow file would be one more place to forget.

### The integration run defeats the test result cache

`integration.yml` passes `-count=1`. `actions/setup-go` restores `GOCACHE` between runs by default, and Go's test result cache keys on the test binary and its observed inputs, none of which change when the schedule fires against an unchanged commit. The run would report a cached pass without issuing a single HTTP request, and issuing them is the one thing it exists to do. This was confirmed by running the suite twice locally and watching the second run return `(cached)`.

`CLAUDE.md` documents the same flag on the project's test command, so a local run does not fall into the same trap.

### Every job carries a timeout, and the gate cancels superseded runs

Each of the three workflows sets `timeout-minutes` on its job. `integration.yml`'s `go test -timeout 20m` bounds only the test binary; a job that hangs in checkout, in module download, or on the runner itself would otherwise run to GitHub's 360-minute default.

`ci.yml` additionally declares a `concurrency` group keyed on workflow and ref with `cancel-in-progress`, so pushing twice to a pull request stops the superseded run. `integration.yml` and `release.yml` omit it: neither a daily schedule nor a tag push tends to start while the previous run is still going, and canceling a half-finished release would be worse than letting it finish.

### Tag glob `[0-9]+.[0-9]+.[0-9]+`, no `v` prefix

The five existing tags are unprefixed, and SemVer 2.0.0 states that `v1.2.3` is not a semantic version, only a common tagging convention. The `v` prefix would buy Go module resolution (`go get` requires `vX.Y.Z`), but the packages under this module have no external importers. Consistency with the existing tags wins, and the decision is symmetric: later tags can adopt the prefix without rewriting old ones.

The glob is anchored on all three components, so a prerelease tag such as `1.0.5-rc1` does not trigger a release. The project publishes no prereleases today; if that changes, the glob and the `latest` handling both need revisiting together.

### No GoReleaser

The release job is GHCR login, `docker/metadata-action`, `docker/build-push-action`, then `gh release create "$GITHUB_REF_NAME" --generate-notes`. GoReleaser's value is cross-compiled binary archives and package manifests, and nothing consumes a feedgen binary. Those four steps of plain YAML are less to maintain than a `.goreleaser.yaml` whose binary half would be dead weight.

`docker/metadata-action` maps the git tag to the image tag and applies `latest` through its default `latest=auto` behavior, which tags any non-prerelease semver version without comparing it against earlier tags. Since the glob already excludes prereleases, every release that runs is a candidate for `latest`.

### Image built only at release time

`docker build` does not run on push or on pull requests. Given the Alpine version pins, a pull request touching only Go code would fail on an unrelated upstream package bump, which is exactly the false signal the two-workflow split exists to avoid. At release time the same failure is useful: it says the pins need updating before this version ships, and the fix is to update them and move the tag.

### `linux/amd64` only

`platforms` lists only `linux/amd64`. Adding `linux/arm64` would require QEMU emulation, because `Dockerfile`'s final stage runs `apk add` in the target architecture rather than only cross-compiling Go, and that roughly triples release build time. No arm64 deployment target exists.

`docker/setup-buildx-action` is deliberately absent. Its own documentation describes it as not required, and recommends it for building multi-platform images or exporting cache. This build does neither, so including it would start a BuildKit container on every release to save a single line in a hypothetical future. Adding arm64 later means adding a QEMU step anyway, at which point buildx comes back alongside it.

### Go version read from `go.mod`

`actions/setup-go` uses `go-version-file: go.mod`. The toolchain version is already written twice, in `go.mod` and in `Dockerfile`'s builder stage; a hardcoded version in the workflows would be a third copy to keep in step.

### Action versions pinned to major tags

Resolved at the time of writing: `actions/checkout@v7`, `actions/setup-go@v7`, `docker/login-action@v4`, `docker/metadata-action@v6`, and `docker/build-push-action@v7`. Major tags accept patch and minor updates automatically while holding back breaking changes. Full SHA pinning is the stricter alternative and is warranted for workflows handling more sensitive secrets; the GHCR token here is scoped to this repository's own packages and expires with the job.

### Permissions declared per workflow

`ci.yml` and `integration.yml` declare `contents: read`. `release.yml` declares `contents: write` for creating the Release and `packages: write` for pushing to GHCR. Declaring the narrower set at the top of each file means the write permissions exist only in the workflow that needs them, rather than being granted repository-wide.

### Badge event filters

The primary badge filters `?branch=master&event=push`, matching the reference repository, so pull request runs do not move it. The integration badge filters `?branch=master&event=schedule`, so a manual `workflow_dispatch` run used for debugging does not overwrite the nightly result.

## Risks / Trade-offs

- A newly created GHCR package is private, and the first `docker pull` by an anonymous user fails with a misleading authentication error → the package visibility has to be switched to public once in the GitHub UI after the first release; this is not something the workflow can do for itself.
- GitHub disables scheduled workflows in a repository with 60 days of no activity, and the canary would stop without announcing it → the integration badge goes stale rather than red, so a stale timestamp on the badge is the thing to watch, not just its color.
- A daily integration run covering eight parsers across seven external domains will fail intermittently for reasons outside the repository, producing recurring failure notifications → accepted as the cost of the signal; the alternative is not knowing when a parser breaks.
- Alpine package pins can break the release build at the moment a version is being shipped → accepted deliberately, since the alternative is the same breakage arriving on unrelated pull requests.
- `go vet` and `gofmt` become merge blockers, so any future contribution has to satisfy them → this is the intent, and the pre-existing violations are fixed as part of this change rather than being grandfathered with suppressions.
