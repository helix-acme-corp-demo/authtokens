# Requirements: Enabling More Agents on the authtokens Project

## Context

The `authtokens` project is a small Go library (zero external dependencies) providing JWT token creation, validation, and HTTP middleware. The codebase is clean and well-tested (~8 files). To scale development with multiple AI agents working in parallel, we need supporting infrastructure: documentation, CI guardrails, contribution guidelines, and task decomposition patterns.

## User Stories

### US-1: Agent can understand the codebase quickly
**As** a new agent picking up a task,
**I want** clear project documentation and architecture notes,
**So that** I don't waste time rediscovering how things work.

**Acceptance Criteria:**
- A `CONTRIBUTING.md` exists with coding conventions, naming patterns, and project structure
- Architecture notes describe the options pattern (`WithSecret`, `WithDefaultTTL`, etc.) used throughout
- The `Issuer`/`Validator` interface pattern is documented as the primary extension point

### US-2: Agent work doesn't break existing functionality
**As** a project maintainer,
**I want** CI checks that run automatically on every change,
**So that** agents can't merge code that breaks tests or lowers quality.

**Acceptance Criteria:**
- A CI pipeline runs `go test ./...` and `go vet ./...` on every push
- Test coverage is reported (current baseline established)
- Linting with `golangci-lint` catches common issues

### US-3: Agents can work on independent tasks in parallel
**As** a project coordinator,
**I want** a backlog of well-scoped, independent tasks,
**So that** multiple agents can work simultaneously without conflicts.

**Acceptance Criteria:**
- Tasks are scoped to separate files or clearly bounded areas
- Each task has clear inputs, outputs, and acceptance criteria
- Tasks avoid overlapping the same source files where possible

### US-4: Agent output is consistent in style
**As** a project maintainer,
**I want** enforced code formatting and style rules,
**So that** code from different agents looks like it was written by one developer.

**Acceptance Criteria:**
- `gofmt` / `goimports` is enforced (CI fails on unformatted code)
- A `.golangci-lint.yml` config defines the project's lint rules
- Example code in tests demonstrates the expected patterns

## What Already Exists (Learnings)

- **Pattern: Functional options** — All configuration uses `Option` funcs applied to a `config` struct (`WithSecret()`, `WithDefaultTTL()`, etc.). New features must follow this pattern.
- **Pattern: Interface-based design** — `Issuer`, `Validator`, and `RevocationChecker` are interfaces. This makes testing easy and is the primary extension mechanism.
- **Pattern: Zero dependencies** — `go.mod` has no external deps (only stdlib). This is intentional and should be preserved unless there's a strong reason.
- **Test quality is high** — Unit tests, table-driven tests, and `Example*` tests all exist. New code must include tests at the same level.
- **Single package** — Everything is in package `authtokens`. No sub-packages. New files go in the root.