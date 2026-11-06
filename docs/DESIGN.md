# Design

One `Decide(context.Context, Request)` operation and explicit per-client protocol/model/URL composition. `Encode` and `Decode` expose pure conversion without network calls; `Decide` handles HTTP and uses the same snapshotted question contract. There is no global registry or plugin framework. Provider labels do not route or rewrite models.

Request JSON is materialized once (custom MarshalJSON executes once), decoded with UseNumber to avoid float64 evidence rounding, then validated and encoded from the snapshot. Callers must not concurrently mutate input while snapshotting. Any fields accepting dynamic JSON are checked at runtime, so nil or illegal dynamic choice values return typed errors rather than panic. This Go API uses validated any for string/bool union values and structured JSON, not TypeScript's compile-time literal inference.

Protocol distinction is semantic, not just endpoint selection: native state/map/noul and structured criterion content stay separate from OpenAI input_json/ordered arrays/predicate/typed values. Requiring one common representation does not justify inventing labels or silently discarding options. Kind-specific input fields and unknown options fail preflight.

Per-question status preserves valid siblings. Malformed top-level JSON/container, HTTP and transport failures abort the whole call. Score stays original and ComputedScore is independent verification. Confidence is provider-specific; absent usage/numbers are not zero. Raw bytes and extension metadata remain available without reconstruction. Diagnostic bodies and original transport causes require explicit access and may contain sensitive data; normal errors render fixed codes only.

Transport uses stdlib net/http and context deadlines through body consumption. A copied client disables redirects even if caller configured follow behavior. Header maps and extension allowlists are copied. Response and error bodies share a bounded reader. Request JSON is buffered; body budgets do not impose a hard process heap bound, especially for caller-provided MarshalJSON/Transport implementations. No retries, fallback or split execution that could surprise spending.

0.x intentional scope is text/JSON and three question kinds on two protocols. Native limit overrides/extension allowlists describe a verified compatible server; they cannot change decoder semantics. Defaults and validation tolerances are API behavior, not implementation details: semantic changes need changelog/migration review.
