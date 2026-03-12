# Design: Enabling More Agents on the authtokens Project

## Overview

This document outlines what infrastructure, documentation, and task structure we need so that multiple AI agents can work productively on the `authtokens` Go library without stepping on each other or breaking things.

## Current State

| Aspect | Status |
|--------|--------|
| Language | Go 1.22, single package `authtokens` |
| Dependencies | Zero (stdlib only) |
| Files | 8 files: core lib, middleware, tests, examples, go.mod, README, LICENSE |
| Tests | Comprehensive: unit, example, and middleware tests |
| CI | **None** — no automated checks |
| Contributing guide | **None** |
| Lint config | **None** |

## Architecture Decisions

### AD-1: Add a CONTRIBUTING.md with agent-oriented guidance

**Decision:** Create a `CONTRIBUTING.md` that documents the patterns an agent must follow.

**Rationale:** The biggest time-sink for a new agent is rediscovering patterns. Documenting them once saves every future agent 5-10 minutes of code exploration.

**Key content to document:**
- **Functional options pattern** — all config uses `Option` funcs on a `config` struct. Example: `WithSecret()`, `WithDefaultTTL()`. New features MUST add options this way.
- **Interface-based extension** — `Issuer`, `Validator`, `RevocationChecker` are interfaces. New capabilities should be new interfaces or extensions of existing ones.
- **Zero-dependency policy** — no external deps unless absolutely necessary.
- **Test expectations** — every new exported function needs a unit test AND an `Example*` test.
- **Single package** — no sub-packages. All `.go` files go in the repo root.

### AD-2: Add CI pipeline (GitHub Actions or equivalent)

**Decision:** Add a CI config that runs on every push/PR.

**Rationale:** Without CI, an agent can merge broken code. CI is the only reliable safety net when multiple agents work independently.

**Pipeline steps:**
1. `go fmt ./... | diff` — fail if code isn't formatted
2. `go vet ./...` — catch common mistakes
3. `go test -race -coverprofile=coverage.out ./...` — run tests with race detector
4. `golangci-lint run` — extended linting (optional, requires config)

**Kept simple:** One workflow file, one job, ~20 lines of YAML.

### AD-3: Create a task backlog with file-level isolation

**Decision:** Maintain a backlog of well-scoped tasks where each task touches different files.

**Rationale:** Git merge conflicts are the #1 source of agent failures. If two agents edit the same file, one will likely fail on push. Tasks scoped to separate files eliminate this.

**Example task decomposition for this project:**

| Task | Files touched | Independent? |
|------|--------------|-------------|
| Add `WithIssuer` option for custom `iss` claim | `authtokens.go`, `authtokens_test.go` | Yes |
| Add `WithNotBefore` option for `nbf` claim | `authtokens.go`, `authtokens_test.go` | ⚠️ Conflicts with above |
| Add request-ID middleware helper | new `requestid.go`, `requestid_test.go` | Yes |
| Add token-from-cookie extraction | new `cookie.go`, `cookie_test.go` | Yes |
| Add benchmarks | new `bench_test.go` | Yes |

**Key insight:** Tasks that add **new files** are inherently parallelizable. Tasks that modify `authtokens.go` should be serialized or carefully scoped to non-overlapping sections.

### AD-4: Establish a test coverage baseline

**Decision:** Record current test coverage and set a minimum threshold.

**Rationale:** Gives agents a clear quality bar. "Don't decrease coverage" is an easy-to-follow rule.

**Implementation:** Run `go test -coverprofile=coverage.out ./...` in CI and fail if coverage drops below the current baseline (likely ~85-90% given the existing test suite).

### AD-5: Add a Makefile for common commands

**Decision:** A simple `Makefile` with `test`, `lint`, `fmt` targets.

**Rationale:** Agents need a single command to validate their work locally before pushing. `make check` is easier to remember than three separate commands.

## What NOT to do

- **Don't add sub-packages** — the library is small, a flat structure is fine
- **Don't add external dependencies for tooling** — keep `go.mod` clean; CI tools install separately
- **Don't create complex branching strategies** — agents work on feature branches, merge to main
- **Don't over-document** — the code is readable Go; a short CONTRIBUTING.md is enough

## Constraints & Gotchas

- **go.mod has no `require` blocks** — this is intentional (zero deps). Agents must not add deps without explicit approval.
- **All tests use `testing` stdlib** — no testify, no gomock. Keep it that way.
- **The `contextKey` struct in `middleware.go` is unexported** — this is correct for context key isolation. Don't export it.
- **base64 uses `RawURLEncoding`** (no padding) — this is JWT-standard. Don't switch to `StdEncoding`.