# Release checklist

No release was performed by the implementation agent. Target initial version: **v0.1.0**, module `github.com/3inch-ai/system1go`, MIT. Parent reviewer owns commits, public repository creation, push and tag.

1. Review official contract research, all tracked files, Git status/diff, reachable history, author/committer noreply identity, secrets/private paths and packed module contents. Independently review source/security tests. There should be no private fixtures or toolchains in the module.
2. Run gofmt check, `go test -race ./...`, `go vet ./...`, coverage, both examples, docs link validation and `git diff --check`. Test with minimum Go 1.25 and current supported Go versions in CI. Run no hosted inference without separate authorization.
3. Update CHANGELOG and verification with exact reviewed versions/results. Breaking 0.x changes require minor bump and migration notes; compatible fixes patch. Do not automatically release API/dependency updates.
4. After approval, create public `3inch-ai/system1go`, push reviewed main using authorized noreply identity, and read back exact SHA and visibility. Wait for exact-SHA matrix CI, not stale branch results. Only then tag the same reviewed SHA `v0.1.0` and push the tag; verify remote tag SHA. Draft release notes with explicit no-live-inference limitation.
5. In a fresh directory outside the repo, run `go mod init consumer.example/smoke`, `go get github.com/3inch-ai/system1go@v0.1.0`, then compile/run an offline Encode/Decode consumer with no replace directive. Inspect `go list -m -json github.com/3inch-ai/system1go` and module zip contents. Proxy indexing may lag; distinguish direct VCS fetch from proxy availability. Confirm pkg.go.dev only when actually indexed.
6. Read back public visibility, commit/tag identity, CI outcomes and consumer output. Report publication, tag, CI and registry/doc indexing separately; never infer one from another.

Go modules need no npm publication. No other profile, system toolchain or upstream repository modifications are required.
