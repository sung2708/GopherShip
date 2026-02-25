# Story 4.1: Lazy Atomic Environment Sensing

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an engine,
I want to monitor host memory and CPU pressure only every 1024 operations,
so that I can maintain "Stochastic Awareness" without creating cache contention on high-core systems.

## Acceptance Criteria

1. **[AC1]** The `internal/stochastic` package MUST provide a mechanism to track operations and trigger host sensing every $N$ operations (default 1024).
2. **[AC2]** The operation counter MUST be implemented using `sync/atomic` to avoid lock contention in the hot path.
3. **[AC3]** Environment checks MUST include memory pressure (via `runtime.ReadMemStats`) and fallback to "Red Zone" if limits are approached.
4. **[AC4]** On Windows (USER OS), memory sensing MUST be robust; CPU sensing can use a simplified model or a background ticker if native load average is unavailable.
5. **[AC5]** Sensing thresholds MUST be configurable (e.g., 80% RAM for Yellow, 95% for Red).

## Tasks / Subtasks

- [x] Task 1: Implement Atomic Counter Logic (AC: #1, #2)
  - [x] Define `SensingMonitor` struct in `internal/stochastic/monitor.go`.
  - [x] Implement `ShouldCheck()` using `atomic.AddUint64` and modulo or bitmask logic.
- [x] Task 2: Implement Host Sensing (AC: #3, #4)
  - [x] Implement `checkMemory()` using `runtime.MemStats`.
  - [x] Implement `checkCPU()` (basic model for Windows).
- [x] Task 3: State Integration (AC: #5)
  - [x] Update `MustSetAmbientStatus` based on sensor results.
  - [x] Ensure transitions are logged via `rs/zerolog`.
- [x] Task 4: [VERIFICATION] Stress Testing
  - [x] Create `monitor_test.go` verifying that checks only occur every $N$ calls.
  - [x] Verify zero-allocation of the check-decision path.

## Dev Notes

- **Zero-Allocation**: The check for "should I sense now?" MUST be 0 B/op.
- **Hardware Honest**: Use a bitmask (e.g., `counter & 0x3FF == 0`) instead of modulo for the 1024 check if possible, as it's slightly faster on some architectures.
- **Windows Considerations**: `runtime.ReadMemStats` is cross-platform. For CPU on Windows, consider a background goroutine that updates a local "cheap" atomic float64 to avoid expensive syscalls in the hot path.

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/stochastic`
- **Encapsulation**: The `Ingester` will call `monitor.ShouldCheck()` in its main loop.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-4.1) - Functional Requirement FR7.
- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Concurrency-&-Reflex-Patterns) - Stochastic Awareness decision.
- [internal/stochastic/state.go](../../internal/stochastic/state.go) - Target for state integration.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

### File List

- [internal/stochastic/monitor.go](../../internal/stochastic/monitor.go)
- [internal/stochastic/monitor_test.go](../../internal/stochastic/monitor_test.go)
- [internal/stochastic/state.go](../../internal/stochastic/state.go)

## Senior Developer Review (AI)

- **Outcome**: Approved
- **Action Items**:
    - [x] [HIGH] Threshold percentages (80/95) are currently hardcoded; MUST be configurable via `NewSensingMonitor`.
    - [x] [MED] `SensingMonitor.counter` lacks cache-line padding; will cause contention on high-core systems.
    - [x] [LOW] `sampleCPU` is a static mock; implement a simple goroutine-count or ticker delta for more "Hardware Honest" signal on Windows.
