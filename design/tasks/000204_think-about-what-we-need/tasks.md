# Implementation Tasks

## Documentation
- [ ] Create `CONTRIBUTING.md` documenting coding conventions: functional options pattern (`WithX` on `config` struct), interface-based extension (`Issuer`/`Validator`/`RevocationChecker`), zero-dependency policy, single-package layout, and test expectations (unit test + `Example*` test for every exported function)
- [ ] Add architecture section to `CONTRIBUTING.md` explaining the JWT flow: `Issue() → sign → base64url encode` and `Validate() → split → verify HMAC → decode → check claims`

## CI Pipeline
- [ ] Add CI workflow file (e.g., `.github/workflows/ci.yml` or equivalent) with steps: `go fmt` check, `go vet ./...`, `go test -race -coverprofile=coverage.out ./...`
- [ ] Add `golangci-lint` step to CI with a `.golangci-lint.yml` config (keep rules minimal: `govet`, `errcheck`, `staticcheck`, `gofmt`)
- [ ] Establish test coverage baseline by running coverage once and recording the threshold in CI config

## Developer Tooling
- [ ] Create a `Makefile` with targets: `test`, `lint`, `fmt`, `check` (runs all three), `cover` (opens HTML coverage report)

## Task Backlog (example tasks for parallel agents)
- [ ] Draft 3-5 well-scoped feature tasks that touch **separate files** to enable parallel agent work (e.g., `cookie.go` for cookie-based token extraction, `bench_test.go` for benchmarks, `requestid.go` for request-ID middleware)
- [ ] For each drafted task, document which files it touches and confirm no overlap with other tasks

## Validation
- [ ] Run `go test ./...` to confirm all existing tests still pass after adding new files
- [ ] Run `go vet ./...` to confirm no issues introduced
- [ ] Verify CI pipeline triggers correctly on a test push