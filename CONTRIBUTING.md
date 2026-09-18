# Contributing

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/), for example
`fix: handle missing skill metadata` or `feat: add skill search`. Release-
please uses commit types to prepare the release PR and changelog.

Breaking changes before version 1.0 increase the minor version: use a
`BREAKING CHANGE:` footer (or `!` in the commit subject) and release-please
will produce `0.x.0`. After 1.0, breaking changes use the usual major bump.

## Releases

Release-please runs on pushes to `master`. The empty manifest means there is
no published release yet. With no existing release tag, `initial-version`
sets the first release PR to `0.0.1`. The `simple` strategy writes the plain
`VERSION` file and `CHANGELOG.md` in that PR. The workflow automatically
merges the release PR into `master`, then creates the `v0.0.1` tag and a draft
GitHub release. GoReleaser validates the tagged `VERSION`, builds six CGO-free
binaries, uploads their archives and `SHA256SUMS`, and publishes the draft only
after successful uploads. A build or validation failure leaves the release as
a draft for investigation.

After `v0.0.1` is published, a `fix:` commit produces `0.0.2`; a `feat:`
commit produces `0.1.0`. Before 1.0, a breaking commit increases the minor
version (for example, `0.1.0` to `0.2.0`). Do not manually increment
`VERSION` or use `scripts/release.sh` for published releases. That script is
deprecated; `scripts/build-release.sh` remains available for local artifact
and smoke-test builds.
