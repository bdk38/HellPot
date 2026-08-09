# Releasing HellPot Community

This repository uses **one** release path to avoid the branch/tag confusion
that previously came from overlapping workflows and slash-commands.

## Canonical process

1. Land all intended changes on `main` via PR.
2. Update version metadata on `main` before tagging:
   - `internal/config/globals.go` → `Version`
   - `CHANGELOG.md` → new section
   - `README.md` latest-release links (if needed)
   - `SECURITY.md` supported versions (if needed)
3. Ensure CI (`Vibe Check`) is green on `main`.
4. Tag **from `main` only**:

   ```bash
   git checkout main
   git pull origin main
   git tag -a vX.Y.Z -m "HellPot Community vX.Y.Z"
   git push origin main --tags
   ```

5. GitHub Actions workflow **Release** (`.github/workflows/release.yml`) will:
   - build Linux/macOS binaries
   - attach checksums
   - create the GitHub Release
   - optionally push `bdk38/hellpot-community` to Docker Hub when
     `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` secrets exist

## Rules of the road

- **Do not** create releases from feature branches.
- **Do not** use issue comment slash-commands for releases (removed).
- **Do not** manually draft a GitHub Release and expect a second workflow to
  rebuild everything — tagging `main` is enough.
- Prefer annotated tags: `git tag -a`.
- Tag names must match `v*` (example: `v0.7.1`).

## After a release

1. Confirm the Actions run succeeded.
2. Spot-check one Linux and one macOS asset download + checksum.
3. If Docker secrets are configured, confirm the image tag on Docker Hub.
4. Close any milestone / note unreleased items still on `main`.

## Why this is strict

Earlier maintenance used multiple release workflows (`release.yml`,
`release-command.yml`) plus a `/release` slash-command dispatcher. That made
it easy for Copilot/automation to cut tags from the wrong branch or publish
duplicate/conflicting assets. The extra paths were removed on purpose.
