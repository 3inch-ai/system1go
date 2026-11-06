# system1go

A small, dependency-free Go client for **native System One** and **OpenAI Decisions**, from [3inch.ai Labs](https://3inch.ai). Go port of [system1ts](https://github.com/3inch-ai/system1ts).

同じ `Decide(ctx, request)` で Jev・OpenAI・ローカル H2O を利用できます。変換できない機能は送信前にエラーにし、回答拒否と不正・欠落回答を区別します。

## Install

Go 1.25 or later. The reviewed v0.1.0 release is published:

```sh
go get github.com/3inch-ai/system1go@v0.1.0
```

## Quickstart

```go
client, err := system1go.Jev(system1go.Config{APIKey: os.Getenv("TYPESAFE_API_KEY")})
if err != nil { log.Fatal(err) }
decision, err := client.Decide(ctx, system1go.Request{
    Input: "I was charged twice. Please refund the extra charge.",
    Questions: []system1go.Question{
        {Name: "billing", Type: system1go.Predicate, Instructions: "Is this about billing?"},
        {Name: "route", Type: system1go.Choice, Instructions: "Choose a team",
            Choices: []system1go.Option{{Value: "billing"}, {Value: "shipping"}}},
        {Name: "urgency", Type: system1go.Score, Instructions: "How urgent?",
            Levels: []system1go.Level{{Label: "low"}, {Label: "high", Description: "Needs attention today"}}},
    },
})
if err != nil { log.Fatal(err) }
answer := decision.Answers["billing"]
if answer.Status == system1go.OK { fmt.Println(*answer.Probability) }
```

The snippet belongs inside an application using `context`, `os`, `log`, `fmt`, and `github.com/3inch-ai/system1go`. Complete, runnable **no-key/no-inference** examples:

```sh
go run ./examples/quickstart   # real loopback HTTP, synthetic response
go run ./examples/providers    # inspect Jev/OpenAI/H2O wire shapes, no HTTP
```

For OpenAI use `system1go.OpenAI(Config{APIKey: ...})`; for an **already running native shim** use `system1go.H2O(Config{})`. Defaults: Jev `https://api.typesafe.ai` / `jev-latest`; OpenAI `https://api.openai.com` / `gpt-6-luna`; H2O `http://127.0.0.1:8741` / `h2oai/h2o-lightning-4b`. An ordinary chat-completions server is not a native shim. No server is started, model downloaded, environment variable read, or HTTP request sent at import/construction.

For explicit composition:

```go
client, err := system1go.NewClient(system1go.Config{
    Protocol: system1go.SystemOne, Provider: "my-deployment",
    BaseURL: "http://localhost:8741/prefix/v1", Model: "verified-model",
    HTTPClient: &http.Client{Transport: myTransport},
    Timeout: 20 * time.Second, MaxResponseBytes: 2 << 20,
    ExtraBodyKeys: []string{"attention_mode"}, // verified compatible-server extension
})
```

`Provider` is descriptive, never routing/model rewriting. Root or `/v1` URLs both work and custom prefixes survive. Configured `http.Client` is copied; redirects are disabled even on injected clients. Its Transport and CookieJar remain shared and must be concurrency-safe and honor request contexts. Do not mutate request data during calls.

## Results, errors and semantics

* `Answers[name].Status`: `OK`, `Refused` (only explicit OpenAI refusal), or `Invalid`. Missing/wrong-type/unaligned answers are **not** refusals; valid siblings survive. Check status before dereferencing optional numeric pointers.
* `Probability` is truth probability, never a thresholded boolean. `Choice` and distribution values preserve `true` versus `"true"`; native System One rejects booleans rather than inventing labels.
* `Score` retains the returned score; `ComputedScore` is separate verification. `Levels` retains ordered labels/descriptions. `Confidence` plus `ConfidenceSource` is provider-specific certainty, **not** portable correctness probability.
* `Raw` retains original response bytes. `Metadata` retains extra top-level fields, `Usage` remains nil if absent (explicit JSON null is retained), and requested/actual models remain distinct. Raw evidence and metadata can be sensitive: do not log without review.
* Request/transport/envelope failures return `*system1go.Error`; inspect via `errors.As`. Codes: `invalid_request`, `unsupported`, `authentication`, `rate_limit`, `timeout`, `aborted`, `network`, `http`, `invalid_response`, `request_too_large`, `response_too_large`. No retries, batching, fallback or hidden telemetry. `errors.Is(err, context.Canceled/DeadlineExceeded)` works. Normal error strings never include upstream body, endpoint or credential values. `Diagnostics()` and `Cause()` explicitly expose sensitive debugging material.
* Timeouts default to 30s and cover HTTP headers and body consumption. Request/response budgets default to 32 MiB / 8 MiB (configurable up to 1 GiB). JSON is buffered; budgets do not bound allocations inside caller-provided MarshalJSON or custom transports. Plain HTTP is allowed for local servers; use HTTPS for hosted credentials.

### Explicit provider options

OpenAI `Request.Options` accepts `prompt_cache_key`, `prompt_cache_options` (`mode`: implicit/explicit/null, `ttl`: 30m), `safety_identifier`, and `inputTextCacheBreakpoint` (text-only explicit end marker). Nullable top-level controls preserve explicit null. Unknown options fail before HTTP. Example:

```go
Options: map[string]any{
    "prompt_cache_key": "support-v1",
    "prompt_cache_options": map[string]any{"mode": "explicit", "ttl": "30m"},
    "inputTextCacheBreakpoint": true,
},
```

System One options are direct extra body keys, **not nested `extraBody`**. Only `Config.ExtraBodyKeys` declared keys pass; protected model/state/questions/assets cannot be overridden. Official Jev allows none by default. Long-context opt-in uses `Config.Headers`. Caller-declared extensions/limits are promises about a separately verified deployment, not new server compatibility claims.

## Scope and verification

| Common surface | Native System One | OpenAI Decisions |
|---|---|---|
| Text / JSON input | state string/object/array | text or native input_json object/object-array |
| Predicate | noul, optional structured/null instructions and criteria | text instructions, no native predicate criteria |
| Choice | 1–255 string options, structured descriptions | 2–255 distinct typed string/boolean options, text descriptions |
| Score | 1–10 ordered structured criteria | 2–10 labeled levels, text descriptions |
| Questions | no default fixed count cap | 1–400, ordered with echoed names checked |
| Refusal | no assumed variant | explicit refusal supported |

No rich media, Box, conversations/tool histories, model management, proxy, automatic retries or pricing estimator. Complete probability coverage, sums (absolute tolerance .01), maximum-probability choice, and weighted-score consistency (.02) are **SDK semantic validation policies**; OpenAI's response schema alone does not guarantee all of these. No normalization or missing-value filling occurs. See [API contracts](docs/API.md), [design](docs/DESIGN.md), [competitor evidence](docs/COMPETITOR-RESEARCH.md), [verification](docs/VERIFICATION.md), and [release steps](docs/RELEASING.md).

Synthetic HTTP tests are not provider/model tests. **No live Jev, OpenAI, or H2O inference has been run for this Go port.** Upstream TypeScript live results do not establish Go compatibility. GitHub CI is configured for Linux/macOS/Windows; hosted runs require publication and must be checked separately.

## Development

```sh
go test -race ./...
go vet ./...
go test -coverprofile=coverage.out ./...
go run ./examples/quickstart
go run ./examples/providers
```

See [contributing](CONTRIBUTING.md), [security](SECURITY.md), [roadmap](docs/ROADMAP.md) and [changelog](CHANGELOG.md). MIT license.
