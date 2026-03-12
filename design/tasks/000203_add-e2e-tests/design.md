# Design: Add E2E Tests for authtokens

## Overview

Add end-to-end tests that exercise the `authtokens` library through real HTTP round-trips. The existing unit tests use `httptest.NewRecorder` (in-memory); e2e tests will use `httptest.NewServer` (real TCP listener) with `http.Client` to validate the full request lifecycle.

## Codebase Learnings

- **Package structure:** Single-package Go library at `authtokens/`, module `github.com/helix-acme-corp-demo/authtokens`, Go 1.22, zero external dependencies.
- **Public API surface:** `NewIssuer`, `NewValidator`, `Middleware`, `ClaimsFromContext`, plus options (`WithSecret`, `WithDefaultTTL`, `WithAudience`, `WithRequiredScopes`, `WithRevocationCheck`).
- **Existing test pattern:** Unit tests in `authtokens_test.go` (internal package), middleware tests in `middleware_test.go` (internal), example tests in `example_test.go` (external `authtokens_test` package). All use `httptest.NewRecorder`.
- **Error types:** `ErrInvalidToken`, `ErrExpiredToken`, `ErrInvalidSignature`, `ErrInsufficientScopes`, `ErrRevokedToken`, `ErrMissingToken` — all returned as JSON `{"error": "..."}` by middleware.
- **RevocationChecker:** Interface with `IsRevoked(id string) bool`. Tests will need a simple fake implementation (similar to `fakeRevocationList` in `authtokens_test.go`, but in the external test package).

## Architecture

### Test File

Single file: **`e2e_test.go`** in the `authtokens` directory, using external test package `authtokens_test`.

This follows the existing convention of `example_test.go` which also uses the external package. The external package perspective ensures we test only the public API — the same surface a real consumer would use.

### Test Structure

Each test function:
1. Creates an `Issuer` and `Validator` with appropriate options
2. Sets up a handler behind `Middleware(validator)` that echoes claims as JSON
3. Starts a real server with `httptest.NewServer`
4. Makes HTTP requests with `http.Client`
5. Asserts status code, `Content-Type` header, and JSON response body

### Helper Pattern

A small test helper function creates the standard protected server (middleware + handler that returns claims as JSON). Each test scenario calls this helper, avoiding boilerplate. The helper returns the server URL; `defer server.Close()` handles cleanup.

A minimal `revocationList` struct implements `RevocationChecker` for revocation tests.

### Key Decision: Single File vs. Multiple Files

**Chose single file.** The library is small and all e2e scenarios are variations of "make HTTP request, check response." Splitting across files would add complexity without benefit.

### Key Decision: External Package

**Chose `authtokens_test` (external).** This ensures tests only access the public API, matching how real consumers use the library. Matches the existing `example_test.go` convention.

## Test Scenarios

| Scenario | Token Setup | Expected Status | Expected Body |
|---|---|---|---|
| Valid token | Standard claims, 1h TTL | 200 | `{"subject":"user:123"}` |
| Missing header | No Authorization header | 401 | `{"error":"authorization token not provided"}` |
| Malformed token | `"not-a-jwt"` | 401 | `{"error":"token is malformed"}` |
| Expired token | ExpiresAt in the past | 401 | `{"error":"token has expired"}` |
| Wrong secret | Signed with different key | 401 | `{"error":"signature verification failed"}` |
| Valid scopes | Token has required scopes | 200 | Claims JSON |
| Missing scopes | Token lacks required scope | 401 | `{"error":"token missing required scopes"}` |
| Revoked token | Token ID in revocation list | 401 | `{"error":"token has been revoked"}` |
| Non-revoked token | Token ID not in list | 200 | Claims JSON |
| Audience mismatch | Token aud ≠ validator aud | 401 | Error JSON |
| Token refresh flow | Issue → refresh → use new | 200 | Claims JSON |

## Constraints

- No new dependencies — standard library only (`net/http`, `net/http/httptest`, `encoding/json`, `testing`)
- Must pass with `go test ./...`
- Must not conflict with existing test files