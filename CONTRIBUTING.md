# Contributing

This is a solo-maintained project run for a specific business (see
[shop_docs](https://github.com/hegarty/shop_docs) for the full picture). It's public for
transparency and portfolio purposes, not actively seeking outside contributors — but
issues and PRs are still welcome, especially bug reports.

## Workflow

Every change goes through a pull request into `main`, even for the maintainer — there is
no direct-push path once the repository's branch protection is active. Required before
merge:

- `build` status check passes: `go build ./...`, `go vet ./...`, `go test ./...`
- `trufflehog` status check passes: no secrets detected in the diff
- Commits are GitHub-verified (signed)

## Local development

```bash
make test   # go test ./...
make vet    # go vet ./...
make lint   # golangci-lint run (install: https://golangci-lint.run/usage/install/)
make fmt    # gofmt -l . — must produce no output
```

Run all of these before opening a PR; CI runs the same checks.

## Commit signing

This repo requires GitHub-verified signed commits. If you haven't set this up:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_signing -N ""
gh api -X POST user/ssh_signing_keys -f title="my signing key" -f key="$(cat ~/.ssh/id_ed25519_signing.pub)"
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519_signing.pub
git config --global commit.gpgsign true
```

Use a dedicated signing key, separate from any key you use for SSH authentication.

## Code style

- Standard `gofmt` formatting, enforced by CI.
- Prefer the standard library and small, well-established dependencies over frameworks.
- Comments explain *why*, not *what* — well-named identifiers should make the "what"
  obvious.
- Table-driven tests where the logic branches on input shape (see `period/period_test.go`
  for the pattern this repo follows).
