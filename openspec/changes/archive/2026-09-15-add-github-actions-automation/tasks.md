## 1. Clear the Pre-Existing Lint Failures

- [x] 1.1 Convert the unkeyed error struct literal in `site/chrb.go` to keyed fields, checking each value against the field order declared in `error.go`
- [x] 1.2 Convert the four unkeyed error struct literals in `site/gamer_forum.go` to keyed fields
- [x] 1.3 Convert the three unkeyed error struct literals in `site/hackernews.go` to keyed fields
- [x] 1.4 Convert the unkeyed error struct literal in `site/hackmd.go` to keyed fields
- [x] 1.5 Convert the two unkeyed error struct literals in `site/ptt.go` to keyed fields
- [x] 1.6 Run `gofmt -w web/main.go`
- [x] 1.7 Confirm `go vet ./...` and `gofmt -l .` both report nothing
- [x] 1.8 Run `go test ./site/...` to confirm the keyed-field conversion preserved every error's field values, rerunning any case that fails on a network error before treating it as a regression

## 2. CI Workflow

- [x] 2.1 Create `.github/workflows/ci.yml` triggered on push to master and on pull requests, with `permissions: contents: read`
- [x] 2.2 Add a single job using `actions/checkout@v7` and `actions/setup-go@v7` with `go-version-file: go.mod`
- [x] 2.3 Add the three gate steps: `make install`, `gofmt -l .` failing when its output is non-empty, and `go vet ./...`
- [x] 2.4 Set `timeout-minutes` on the job and declare a `concurrency` group keyed on workflow and ref with `cancel-in-progress`

## 3. Integration Canary Workflow

- [x] 3.1 Create `.github/workflows/integration.yml` triggered on `schedule` with a daily cron at `0 20 * * *` (04:00 UTC+8) and on `workflow_dispatch`, with `permissions: contents: read`
- [x] 3.2 Add a job that checks out, sets up Go from `go.mod`, and runs `go test -count=1 -timeout 20m ./site/...`; `-count=1` bypasses the test result cache that `setup-go` restores, which would otherwise report a pass without reaching any site
- [x] 3.3 Set `timeout-minutes` on the job, since `go test -timeout` bounds only the test binary

## 4. Release Workflow

- [x] 4.1 Create `.github/workflows/release.yml` triggered on pushing tags matching `[0-9]+.[0-9]+.[0-9]+`, with `permissions: contents: write` and `packages: write`
- [x] 4.2 Add the GHCR steps: `docker/login-action@v4` against `ghcr.io` using `github.actor` and `secrets.GITHUB_TOKEN`, then `docker/metadata-action@v6` for `ghcr.io/${{ github.repository }}`, then `docker/build-push-action@v7` with `platforms: linux/amd64` and `push: true`, omitting `docker/setup-buildx-action` because a single-platform build that exports no cache does not need it
- [x] 4.3 Add a final step running `gh release create "$GITHUB_REF_NAME" --generate-notes` with `GH_TOKEN` from `secrets.GITHUB_TOKEN`
- [x] 4.4 Set `timeout-minutes` on the job
- [x] 4.5 Confirm the `metadata-action` tag configuration yields both the version tag and `latest` for a tag such as `1.0.5`, by reading the action's resolved `tags` output format rather than by pushing a tag

## 5. README Badges

- [x] 5.1 Add the CI badge below the `# feedgen` heading, linking to the `ci.yml` workflow page with `?branch=master&event=push` on the badge image URL
- [x] 5.2 Add the integration badge next to it, linking to the `integration.yml` workflow page with `?branch=master&event=schedule` on the badge image URL

## 6. Documentation

- [x] 6.1 Update the Test command in `CLAUDE.md` to `go test -count=1 ./site/...` and say why, so a local run does not hit the cache the CI run now bypasses

## 7. Local Verification

- [x] 7.1 Run `yamllint .github/workflows/` and resolve every finding, adding a project `.yamllint` config if the default `truthy` or `line-length` rules flag valid workflow syntax such as the `on:` key
- [x] 7.2 Run `docker compose build` to confirm that the `Dockerfile` the release workflow builds still succeeds with its current Alpine package pins
- [x] 7.3 Re-run `make install`, `go vet ./...`, and `gofmt -l .` as a final check that this change did not break the gate it introduces
