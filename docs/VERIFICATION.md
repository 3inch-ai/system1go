# Verification record — 2026-11-06

Implementation tested on macOS arm64. No inference credentials read; no paid/provider inference, local model operations, Git commits, push or publication. Synthetic HTTP is clearly labeled and distinct from a model. [Live boundary](LIVE-LOCAL.md).

## Executed gates

| Command | Observed result |
|---|---|
| `go test -race ./...` | PASS on Go 1.27.3 and minimum-series Go 1.25.14 |
| `go vet ./...` | PASS on both toolchains, no diagnostics |
| `go test -coverprofile=coverage.out ./...` | PASS; library package 89.7% statement coverage; aggregate including standalone example main packages 86.2% |
| `go run ./examples/quickstart` | `ok: 0.95 (synthetic HTTP, no inference)` |
| `go run ./examples/providers` | PASS; printed separate Jev state/map/noul, OpenAI native input_json/ordered questions, H2O native shapes, no HTTP |
| `gofmt -w .` | Applied to all Go source/tests/examples |
| `go test -json ./...` | 29 passing top-level tests; 44 passing test/subtest events (Go 1.27.3, after score label presence regression) |
| Fresh external consumer with local module replacement, Go 1.25.14 `go run .` | `consumer: ok 0.8, encoded 183 bytes`; no network/inference |
| Relative Markdown link check | PASS, all local link targets exist |
| `git diff --check` | PASS after intent-to-add of new files, no commit |
| Source scan for machine-private absolute paths | PASS, none found |

Standalone example main packages report 0% in `go test` because they are independently executed with `go run`, not unit-tested main functions. The ignored coverage.out is not a release artifact. GitHub multi-OS/version CI is configured but **has not run** before publication. No claim of live provider acceptance, judgment accuracy, calibration, latency/memory superiority or hosted CI success is made.

## Toolchain provenance

Official `https://go.dev/dl/?mode=json` (and include=all) metadata and archives from `https://go.dev/dl/`; SHA256 checked before extraction. Toolchains are session-local scratch installations, not system-wide changes:

* go1.27.3.darwin-arm64.tar.gz: `dd3908cf20959d798c1e06ebe5b72213f0505ce01373b40ce3110975a6cd99c2`.
* go1.25.14.darwin-arm64.tar.gz: `5b26c0b6f308240fca2614fb02f622cfcc8c0cc3b69c78bba4845489a4590259`.

Set PATH to the verified toolchain bin directory for the commands below. No private machine paths or credentials are necessary for reproduction.

## Actual vertical RED → GREEN record

Research was written before production code. The focused feature-slice commands below were actually run, observed failing for the stated missing behavior, then followed by implementation and a passing `go test ./...`. Early undefined-symbol compiler failures represent new public API slices; no fabricated test output is included. Subsequent anomaly tests extended existing exercised slices rather than inventing all contracts up front.

| Focused RED command | Observed failure before change | GREEN verification |
|---|---|---|
| `go test ./...` (initial NativePredicate test) | undefined NewClient/Config/SystemOne/Request/Question | Native HTTP state/noul/model/raw metadata slice, full suite PASS |
| `go test -run TestTypedChoice -v` | OpenAI JSON input incorrectly encoded as native state | Native input_json and typed ordered choice decode, full suite PASS |
| `go test -run TestScore -v` | missing native criteria and OpenAI levels | Both score codecs with legend/original/computed separation, full suite PASS |
| `go test -run 'TestValidation\|TestOptions' -v` | empty/invalid/unknown requests accepted, cache fields missing | Snapshot validation and explicit option forwarding, full suite PASS |
| `go test -run TestInvalidAnswers -v` | explicit refusal classified invalid | Refusal/invalid/positional-name/duplicate handling, full suite PASS |
| `go test -run TestTransportSafety -v` | missing safe HTTP metadata, timeout/cancel codes, invalid key preflight | Bounded response/error bodies, no redirects, cancellation through body, safe headers/errors, full suite PASS |
| `go test -run 'TestSnapshot\|TestMalformedDistributions' -v` | MarshalJSON called twice; nonmaximum choice accepted | Materialize once, validate/snapshot exact evidence, maximum-selection check, full suite PASS |
| `go test -run 'TestPrecisionAndHeaderIsolation\|TestScoreAnomalies\|TestSafeErrorDiagnostics' -v` | evidence integer 9007199254740993 rounded | UseNumber snapshot, case-insensitive auth conflict, full suite PASS (anomaly/diagnostic assertions already passed existing behavior) |
| `go test -run 'TestPresetsAndCapabilities\|TestStructuredLegend' -v` | undefined presets/capabilities; then numeric structured legend mismatch | Defaults/capability copies plus structural JSON comparison, full suite PASS |
| `go test -run 'TestNoHTTPOnUnsupported\|TestConcurrentClientIsolation' -v` | missing error Protocol provenance | Structured unsupported provenance and concurrency isolation, full suite PASS |
| `go test -run 'TestScoreIndexPresence\|TestScopeLimits' -v` | missing/null score index treated as zero | Presence-aware integer indices, full suite PASS (existing scope checks passed) |
| `go test -run 'TestConfigurationRejectsIgnoredControls\|TestMalformedModelMetadata\|TestLargeStructuredLegend' -v` | ignored OpenAI config accepted, malformed model unreported, large structured legend rounded | Explicit rejection, metadata issues and exact numeric legend parsing, full suite PASS |
| `go test -run TestConfidenceProvenance -v` | missing result protocol/provider/confidence source | Explicit source fields, full suite PASS |
| `go test -run 'TestPublicDecodeRejectsInvalidContract\|TestHeaderSyntaxBeforeOverwrite\|TestEncodedBodyBudget' -v` | invalid direct Decode choice contract panicked | Validate public decode contracts before interpretation; header validation before overriding, full suite PASS |
| `go test -run TestEncodedBodyBudget -v` (extended boundary) | final wire model bytes bypassed Encode budget | Budget applies to final wire body as well as snapshot, full suite PASS |
| `go test -run TestInvalidUTF8IsNotReplaced -v` | malformed source strings/keys/names silently replaced | Pre-serialization UTF-8 validation with depth guard, full suite PASS |
| `go test -run TestScoreLabelPresence -v` | Missing and null OpenAI score labels decoded as OK when the requested label was empty; explicit empty label already passed | Presence-aware label pointer and nil check; all three subtests PASS |

Score label presence regression reverified on Go 1.27.3: `go test -race ./...` PASS, `go vet ./...` PASS with no diagnostics, and `go test -coverprofile=coverage.out ./...` PASS. Fresh `go tool cover -func=coverage.out` reports unchanged library coverage of 89.7%, aggregate including examples 86.2%, and decodeScore 89.4%. The parent additionally reran `go test -count=1 -race ./...` and `go vet ./...` on Go 1.25.14 after this fix; both passed. Only score.go and score_test.go were gofmt-formatted for this fix.

Tests and examples use synthetic inline fixtures. Probability/distribution/score checks are SDK semantic policies; tests do not prove that every hosted provider always returns complete distributions. Tests preserve absence rather than fabricating zeros, retain original JSON and provider fields, and never rename boolean choices for native System One.

## Release boundaries and issue

No commits exist in the isolated main repository; local author/committer identity is configured to the authorized GitHub noreply identity. The implementation agent deliberately did not push, create a remote repository, tag or publish. Public release requires the independent review and exact-SHA checks in [RELEASING](RELEASING.md).

Contributor, architecture and release guidance is maintained in CONTRIBUTING.md and the topical docs linked from README.md. No root agent instruction file is included in this release.
