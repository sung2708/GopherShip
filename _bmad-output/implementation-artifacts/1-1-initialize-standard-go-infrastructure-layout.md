# Story 1.1: Initialize Standard Go Infrastructure Layout

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a platform engineer,
I want the standardized GopherShip repository structure,
so that I have a consistent environment for performance-critical engine development.

## Acceptance Criteria

1. **[AC1]** Root directories `cmd/`, `internal/`, and `pkg/` must exist.
2. **[AC2]** `go.mod` must be initialized with the module name `github.com/sungp/gophership`.
3. **[AC3]** All core internal modules (`ingester`, `vault`, `somatic`, `stochastic`, `control`, `buffer`) must have their directory skeletons initialized.
4. **[AC4]** Entry point binaries must be identifiable in `cmd/gophership` and `cmd/gs-ctl`.

## Tasks / Subtasks

- [x] Task 1: Initialize Go Module (AC: AC2)
  - [x] Run `go mod init github.com/sungp/gophership` in project root.
- [x] Task 2: Create Hardware-Honest Skeleton Layout (AC: AC1, AC3)
  - [x] Create `cmd/gophership` and `cmd/gs-ctl` directories.
  - [x] Create `internal/ingester`, `internal/vault`, `internal/somatic`, `internal/stochastic`, `internal/control`, `internal/buffer`.
  - [x] Create `pkg/protocol` and `pkg/otel`.
- [x] Task 3: Initialize Thin Entry Points (AC: AC4)
  - [x] Create `cmd/gophership/main.go` with basic "Starting GopherShip Engine" log.
  - [x] Create `cmd/gs-ctl/main.go` with basic "GopherShip Control Utility" log.
- [x] Task 4: Internal Encapsulation Documentation
  - [x] Add a `README.md` to the `internal/` directory explaining that this contains the "biological" core logic and should not be exported to external packages.

## Dev Notes

- **Architecture Pattern**: Standard Go High-Performance Layout. [Source: architecture.md#Selected-Starter]
- **Go Version Target**: 1.22+.
- **Encapsulation Rule**: Strict `internal/` directory usage for core logic. [Source: architecture.md#Code-Organization]
- **Naming Convention**: `CamelCase` for exported, `camelCase` for internal. [Source: architecture.md#Naming-Patterns]

### Project Structure Notes

- Alignment with `architecture.md` directory map.
- Uses `cmd/` for executables as per Go best practices.

### References

- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md) - Section: "Standard Go High-Performance Layout"
- [epics.md](../../_bmad-output/planning-artifacts/epics.md) - Story 1.1 Details

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- Root directories `cmd/`, `internal/`, and `pkg/` established.
- `go.mod` module path corrected to `github.com/sungp/gophership`.
- Core internal modules (`ingester`, `vault`, `somatic`, `stochastic`, `control`, `buffer`) initialized.
- Entry points `cmd/gophership/main.go` and `cmd/gs-ctl/main.go` created.
- `internal/README.md` added.
- Existing `stability` module refactored into `stochastic` to align with architecture.

### File List

- [go.mod](../../go.mod)
- [cmd/gophership/main.go](../../cmd/gophership/main.go)
- [cmd/gs-ctl/main.go](../../cmd/gs-ctl/main.go)
- [internal/README.md](../../internal/README.md)
- [internal/stochastic/state.go](../../internal/stochastic/state.go)
- [internal/ingester/ingester.go](../../internal/ingester/ingester.go)
