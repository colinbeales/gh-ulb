# Releasing gh-ulb

Releases are automated via [`.github/workflows/release.yml`](../.github/workflows/release.yml): pushing a Git tag that matches `v*` triggers GitHub Actions to build and publish precompiled binaries.

## Cutting a Release

Use this workflow to cut a new release:

1. Ensure `main` is up to date and tests pass.

```bash
git checkout main
git pull --ff-only
go test ./...
```

2. Create and push a new semver tag (example: `v0.1.0`).

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

3. Verify the `Release` workflow succeeds in GitHub Actions.

4. Confirm a new GitHub Release exists with attached binaries.

5. Smoke test the published version:

```bash
gh extension upgrade ulb
gh ulb --help
```

## Local Development

When running from a local checkout, `gh extension install .` requires an executable named `gh-ulb` in the repository root.

Use this workflow for local development and testing:

```bash
cd /path/to/gh-ulb
go test ./...
go build -o gh-ulb .
gh extension remove colinbeales/gh-ulb
gh extension install . --force
gh ulb --help
```

Notes:

- `go build ./...` compiles packages but does not create the root `gh-ulb` executable required by `gh extension install .`.
- If install fails with "there is already an installed extension that provides the \"ulb\" command", remove the existing extension first, then reinstall.
- Re-run `go build -o gh-ulb .` and `gh extension install . --force` after local code changes to ensure `gh ulb` uses your latest build.

## Local Checks 

Recommended local preflight before opening a PR or tagging a release:

```bash
go vet ./... && golangci-lint run && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

If `golangci-lint` is not installed locally:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Note: Actions currently uses Go `1.25.x`, so using the same local Go major/minor reduces CI-only surprises.

## Notes

- The release workflow runs only for tag pushes matching `v*`.
- If a tag is pushed by mistake, delete it locally and remotely before creating the corrected tag.

```bash
git tag -d v0.1.0
git push origin :refs/tags/v0.1.0
```
