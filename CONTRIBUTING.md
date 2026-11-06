# Contributing

Read AGENTS.md and docs/API.md before changing conversions. Retrieve current official contract and pin reviewed competitor source before public comparisons. Use a failing public-contract test for one vertical behavior, run and observe RED, implement minimally, then run GREEN and the full suite. Keep transport and pure Encode/Decode separate. Record commands/results in docs/VERIFICATION.md, update CHANGELOG and related documentation together.

Use Go 1.25+, gofmt, `go test -race ./...`, `go vet ./...`, `go test -coverprofile=coverage.out ./...`, and both runnable examples. Tests use synthetic, non-personal evidence and local httptest servers; no environment key discovery or live inference. Fixtures are synthetic unless individually sourced. External model accuracy is not an SDK test.

Repository-local Git identity for authorized commits must use GitHub noreply, for both author and committer. Never commit keys, private machine paths, coverage profiles, downloaded toolchains or unsanitized provider data. Independent review precedes publication. Breaking 0.x changes require a minor bump and migration notes; compatible fixes are patches.
