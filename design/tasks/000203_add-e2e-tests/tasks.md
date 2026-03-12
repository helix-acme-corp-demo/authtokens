# Implementation Tasks

- [ ] Create `e2e_test.go` in `authtokens/` with package `authtokens_test`
- [ ] Implement test helper: `revocationList` struct implementing `authtokens.RevocationChecker`
- [ ] Implement test helper: `startServer(validator)` that wraps a claims-echoing handler with `authtokens.Middleware`, starts `httptest.NewServer`, and returns the server
- [ ] Test: valid token → 200 with claims JSON in response body
- [ ] Test: missing Authorization header → 401 with `{"error":"authorization token not provided"}`
- [ ] Test: malformed token string → 401 with `{"error":"token is malformed"}`
- [ ] Test: expired token → 401 with `{"error":"token has expired"}`
- [ ] Test: wrong signing secret → 401 with `{"error":"signature verification failed"}`
- [ ] Test: valid scopes present → 200
- [ ] Test: missing required scopes → 401 with `{"error":"token missing required scopes"}`
- [ ] Test: revoked token → 401 with `{"error":"token has been revoked"}`
- [ ] Test: non-revoked token with revocation checker enabled → 200
- [ ] Test: audience mismatch → 401
- [ ] Test: token refresh flow — issue, refresh, use refreshed token → 200
- [ ] Verify all tests pass with `go test ./...`
