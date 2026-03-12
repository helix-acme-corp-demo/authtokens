# Requirements: Add E2E Tests for authtokens

## Context

The `authtokens` library is a zero-dependency Go package for HS256 JWT token issuance, validation, and HTTP middleware. It already has comprehensive unit tests (`authtokens_test.go`, `middleware_test.go`) and example tests (`example_test.go`), but all tests use `httptest.NewRequest`/`httptest.NewRecorder` — none exercise the full HTTP round-trip through a real server.

## User Stories

1. **As a library maintainer**, I want end-to-end tests that exercise the full HTTP request lifecycle (real TCP server → real HTTP client → middleware → handler → response), so I can be confident the middleware works correctly in production-like conditions.

2. **As a library consumer**, I want to see e2e tests demonstrating realistic usage patterns (issue token, make HTTP request, get authenticated response), so I can trust the library and use the tests as integration examples.

## Acceptance Criteria

### Full HTTP Round-Trip Tests
- [ ] Tests use `httptest.NewServer` (real TCP listener) instead of `httptest.NewRecorder`
- [ ] Tests make actual `http.Client` requests to the running server
- [ ] Tests verify response status codes, headers (`Content-Type`), and body content

### Scenarios Covered
- [ ] Valid token: issue → request with Bearer header → 200 with claims in response body
- [ ] Missing Authorization header → 401 with JSON error body
- [ ] Invalid/malformed token → 401 with JSON error body
- [ ] Expired token → 401 with JSON error body
- [ ] Wrong signing secret → 401 with JSON error body
- [ ] Scope-based authorization: valid scopes pass, missing scopes → 401
- [ ] Revocation checking: revoked token → 401, non-revoked token → 200
- [ ] Token refresh: issue → refresh → use refreshed token successfully
- [ ] Audience mismatch → 401

### Code Quality
- [ ] Tests live in `example_e2e_test.go` using the `authtokens_test` package (external/black-box perspective)
- [ ] No new dependencies — only Go standard library
- [ ] Tests pass with `go test ./...`
