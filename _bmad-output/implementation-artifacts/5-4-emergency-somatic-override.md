# Story 5.4: Emergency Somatic Override

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator,
I want to manually force the system into a specific somatic state,
so that I can proactively protect the host before a known traffic spike hits the sensors.

## Acceptance Criteria

1. **[AC1]** The `gs-ctl` CLI MUST support an `override` command to force the somatic zone (e.g., `gs-ctl override --zone red`).
2. **[AC2]** When an override is active, the engine MUST stay in that zone regardless of sensor data from the `StochasticMonitor`.
3. **[AC3]** The override MUST be clearable (e.g., `gs-ctl override --zone none` or similar) to return control to the sensors.
4. **[AC4]** Every override action MUST be logged at `INFO` level in the GopherShip daemon for auditability.
5. **[AC5]** The control plane MUST be secured via the established mTLS or UDS authentication patterns.

## Tasks / Subtasks

- [x] Protocol Update (AC: 1, 5)
  - [x] Add `OverrideSomaticZone` RPC to `ControlService` in `pkg/protocol/control.proto`.
  - [x] Define `OverrideSomaticZoneRequest` message with `SomaticZone` field.
  - [x] Regenerate gRPC Go code using `protoc` or `buf`.
- [x] Core Engine Integration (AC: 2, 4)
  - [x] Update `internal/somatic/controller.go` to store and respect a manual override state.
  - [x] Add `Override(zone)` and `ClearOverride()` methods to `somatic.Controller`.
  - [x] Update `Reassess()` logic to bypass sensor checks if an override is active.
- [x] Control Server Implementation (AC: 4, 5)
  - [x] Implement `OverrideSomaticZone` in `internal/control/server.go`.
  - [x] Ensure the server logs the override action with caller details (if available via UDS/mTLS).
- [x] CLI Command Implementation (AC: 1, 3)
  - [x] Add `override` command case to `cmd/gs-ctl/main.go`.
  - [x] Add `--zone` flag to specify desired state (green, yellow, red, none).
- [x] Verification
  - [x] Verify `gs-ctl override --zone red` instantly changes the zone reported by `gs-ctl status`.
  - [x] Check daemon logs for the audit entry.
- [x] Review Follow-ups (AI)
  - [x] [AI-Review][High] Fix sync lag in `ClearOverride` immediately refreshing sensor state.
  - [x] [AI-Review][Medium] Harden security tests by passing mock controller instead of `nil`.
  - [x] [AI-Review][Low] Add internal documentation for `overrideZone` magic offsets.

## Dev Notes

- **Somatic Controller**: The `somatic.Controller` is the central brain for zone transitions. It currently monitors buffer depth. The override should take precedence over `Reassess` logic.
- **Auditability**: Use `rs/zerolog` for daemon logging. Include the targeted zone and acknowledgment of the override.
- **Zero-Allocation**: Ensure the new RPC implementation follows the zero-allocation patterns (NFR.P1), although management commands are not on the "hot path," consistency is preferred.

### Project Structure Notes

- Adheres to the unified project structure: `pkg/protocol/` for contracts, `internal/somatic/` for logic, `internal/control/` for transport, and `cmd/gs-ctl/` for the user interface.

### References

- [Architecture Decision: Management API](../../_bmad-output/planning-artifacts/architecture.md#API%20&%20Communication%20Patterns)
- [Epic 5.4 Definition](../../_bmad-output/planning-artifacts/epics.md#Story%205.4:%20Emergency%20Somatic%20Override)

## Dev Agent Record

### Agent Model Used

Antigravity (BMad-Story-Creator-v1)

### Debug Log References

- Reviewed `internal/somatic/controller.go` (Step 183) - identified atomic `curr` state.
- Reviewed `pkg/protocol/control.proto` (Step 177) - confirmed missing RPC.
- Reviewed `cmd/gs-ctl/main.go` (Step 184) - identified command switch.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created for manual somatic overrides.

### File List

- pkg/protocol/control.proto
- pkg/protocol/control.go
- internal/somatic/controller.go
- internal/control/server.go
- internal/ingester/ingester.go
- cmd/gs-ctl/main.go
- cmd/gophership/main.go

## Senior Developer Review (AI)

**Outcome**: 🔴 Changes Requested
**Date**: 2026-02-25

### Action Items
- [x] [High] Fix sync lag in `ClearOverride` [controller.go]
- [x] [Medium] Harden security tests: pass mock instead of `nil` [security_test.go]
- [x] [Low] Document `overrideZone` offsets [controller.go]
- [x] [Medium] Document missing files in Story File List
