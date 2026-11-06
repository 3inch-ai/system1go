package system1go

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Is supports context cancellation checks without exposing unsafe transport error strings.
func (e *Error) Is(target error) bool {
	return (e.Code == "timeout" && target == context.DeadlineExceeded) || (e.Code == "aborted" && target == context.Canceled)
}

// Cause explicitly exposes the original transport error; it may contain sensitive URLs.
func (e *Error) Cause() error { return e.cause }
func transportError(ctx context.Context, cause error) *Error {
	code := "network"
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		code = "timeout"
	} else if ctx.Err() != nil {
		code = "aborted"
	}
	return &Error{Code: code, cause: cause}
}
func statusCode(s int) string {
	switch s {
	case 401, 403:
		return "authentication"
	case 429:
		return "rate_limit"
	case 400, 413, 415, 422:
		return "invalid_request"
	}
	return "http"
}
func safeHTTP(r *http.Response) HTTPMetadata {
	h := make(http.Header)
	for _, k := range []string{"Content-Type", "Retry-After", "X-Request-Id", "Request-Id"} {
		if v := r.Header.Get(k); v != "" {
			h.Set(k, v)
		}
	}
	id := h.Get("X-Request-Id")
	if id == "" {
		id = h.Get("Request-Id")
	}
	return HTTPMetadata{Status: r.StatusCode, RequestID: id, Headers: h}
}
func validHeaderName(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c > 127 || !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
			return false
		}
	}
	return true
}
func validHeaderValue(s string) bool {
	for _, c := range []byte(s) {
		if c == 127 || (c < 32 && c != '\t') {
			return false
		}
	}
	return true
}
