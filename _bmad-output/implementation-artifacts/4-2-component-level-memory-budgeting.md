# Story 4.2: Component-Level Memory Budgeting

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a system administrator,
I want to define hard memory limits for the Ingester and Raw Vault,
so that GopherShip never causes a host-level OOM crash.

## Acceptance Criteria

1. **[AC1]** The system MUST support defining hard memory limits (budgets) for the Ingester and Raw Vault components (configured in bytes).
2. **[AC2]** When any component's tracked memory usage reaches its "Red" threshold (default 95% of budget) or "Yellow" threshold (default 80%), the global `AmbientStatus` MUST transition accordingly.
3. **[AC3]** If a component-level memory breach occurs, `MustSetAmbientStatus(StatusRed)` MUST be called, forcing the engine onto the "Debt Path" (Raw Vault WAL).
4. **[AC4]** Memory tracking MUST be high-fidelity but "Hardware Honest" (lazy/atomic) to avoid contention in the ingestion reflex.
5. **[AC5]** Component pressure transitions MUST be logged via `rs/zerolog` and Prometheus metrics MUST reflect the resulting somatic zone.

## Tasks / Subtasks

- [x] Task 1: Component Memory Tracking (AC: #1, #4)
  - [x] Implement budget definitions (Ingester vs Vault) in `internal/stochastic`.
  - [x] Extend `SensingMonitor` to handle multiple component budgets.
- [x] Task 2: Budget Enforcement Logic (AC: #2, #3)
  - [x] Update `monitor.Sense()` to evaluate component-specific usage against budgets.
  - [x] Implement transition logic to `StatusRed` upon any component budget breach.
- [x] Task 3: Integration & Logging (AC: #5)
  - [x] Ensure `MustSetAmbientStatus` is triggered with appropriate reason logging.
  - [x] Verify Prometheus gauge `gophership_ingester_zone` updates correctly.
- [x] Task 4: [VERIFICATION] Budget Stress Test
  - [x] Create tests simulating component memory pressure.
  - [x] Verify that breaching an Ingester budget forces the system into the Red Zone.
- [x] Task 5: Review Follow-ups (AI)
  - [x] Pre-calculate budget thresholds for zero-cycle performance.
  - [x] Add `gophership_ingester_usage_bytes` and `gophership_vault_usage_bytes` gauges.
  - [x] Rename `Sense()` to `MustSense()` for standard compliance.

## Dev Notes

- **Stochastic Awareness**: Do not perform expensive memory checks on every packet. Use the `ShouldCheck()` logic established in Story 4.1.
- **Buffer Pool Integration**: Since GopherShip uses `sync.Pool` for buffers, "Ingester usage" can be tracked by the number of active buffers multiplied by buffer size, rather than relying solely on `runtime.MemStats`.
- **Global vs Local**: While Story 4.1 monitors total host pressure, Story 4.2 focuses on *internal component self-regulation*. A component breach is as critical as a host breach.

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/stochastic`
- **Dependency**: Reacts with `internal/ingester` and `internal/vault` to sense their "fullness".

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-4.2) - Functional Requirement FR9.
- [4-1-lazy-atomic-environment-sensing.md](../../_bmad-output/implementation-artifacts/4-1-lazy-atomic-environment-sensing.md) - Patterns established for sensing and atomics.
- [internal/stochastic/monitor.go](../../internal/stochastic/monitor.go) - Core implementation target.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

- Implemented atomic usage reporting in `SensingMonitor`.
- Integrated `Monitor` into `Ingester` and `WAL` for high-fidelity component tracking.
- Verified budget enforcement via `TestSensingMonitor_ComponentBudgets`.
- Confirmed zero-allocation (~4.8ns/op) for the monitoring path.
- Optimized thresholds via pre-calculation in the constructor.
- Added usage metrics for Ingester and Vault.

### File List

- `internal/stochastic/monitor.go`
- `internal/stochastic/state.go`
- `internal/stochastic/metrics.go`
- `internal/stochastic/monitor_test.go`
- `internal/ingester/ingester.go`
- `internal/vault/wal.go`
- `internal/vault/wal_test.go`
- `cmd/gophership/main.go`
