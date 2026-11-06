# Security

Do not include credentials, personal input or raw responses in public issues. Report a suspected vulnerability privately to the repository maintainers using GitHub private vulnerability reporting if enabled; otherwise request a private reporting channel without disclosing the issue publicly.

The client never logs, reads environment variables or retries automatically. Redirects are rejected, normal errors are sanitized, headers are validated and response/error bodies share one bounded reader. Debugging methods and raw result fields are deliberately explicit and sensitive, not a secret vault. Custom HTTP transports must honor contexts, TLS policy and concurrency. The caller is responsible for endpoint trust, safe input handling, authorization, spending and downstream threshold calibration. HTTP is supported for local services, not recommended for remote secrets.

Only the latest reviewed 0.x release receives fixes. CI and synthetic contracts cannot prove live provider availability or model safety. Dependency-free means Go standard-library security updates still matter.
