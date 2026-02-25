# Story 4.3: Stochastic Component Throttling

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an optimizer,
I want to slow down background sync and parsing tasks when the Stochastic Monitor detects high host load,
so that the Ingestion reflex always has priority access to CPU cycles.

## Acceptance Criteria

1. **[AC1]** Background workers (e.g., `Replayer`, any background sync) MUST query the global `stochastic.AmbientStatus` before processing a new batch.
2. **[AC2]** When `AmbientStatus` is `StatusYellow`, background tasks MUST double their default sleep interval or reduce their batch size by 50%.
3. **[AC3]** When `AmbientStatus` is `StatusRed`, background tasks MUST enter a "Deep Sleep" state (at least 5s sleep per batch) or Suspend entirely until pressure returns to `StatusGreen` or `StatusYellow`.
4. **[AC4]** Throttling MUST be implemented using "Lazy Sensing" (atomic status load) to avoid adding overhead to the hot path.
5. **[AC5]** Throttling events MUST be logged with a "Starvation Score" (time spent in throttle vs processing) via `rs/zerolog` to enable auditing of system stress.

## Tasks / Subtasks

- [x] Task 1: Throttling Interface in `internal/stochastic` (AC: #1, #4)
  - [x] Implement a `Throttleable` interface or helper function that returns a dynamic sleep multiplier based on current status.
- [x] Task 2: Dynamic Replayer Throttling (AC: #2, #3, #5)
  - [x] Modify `internal/vault/Replayer` to use a `stochastic.ThrottleMultiplier()` for dynamic sleep adjustments.
  - [x] Implement the "Deep Sleep" loop (5s+ wait) for `StatusRed`.
- [x] Task 3: Ingestion Priority Weighting (AC: #2, #5)
  - [x] Implement "Starvation Score" metrics that track the cumulative delay introduced by throttling.
  - [x] Ensure logging highlights when background tasks are actively yielding for the Ingester.
- [x] Task 4: [VERIFICATION] Throttling Stability Test
  - [x] Create a test simulating a "Yellow Zone" and verifying `Replayer` throughput reduction.
  - [x] Verify background tasks resume normal operation when zone returns to "Green".
- [x] Task 5: Review Follow-ups (AI)
  - [x] Use `atomic.Int64` for `starvationTime` and `processingTime` to prevent metrics race conditions.
  - [x] Expose `minDeepSleep` in `Replayer` to allow stable, fast verification tests.
  - [x] Add `cumulative_starvation` to throttle event logs for better audit observability.
  - [x] Fix naming inconsistency (`StatusMultiplier` -> `ThrottleMultiplier`).

## Dev Notes

- **Priority Weighting**: The goal is to ensure that even if CPU usage is high, the ingestion reflex (select-default) never hits scheduler delay because background workers are yielding.
- **Lazy Sensing**: Use `stochastic.GetAmbientStatus()` which is an atomic load; keep it efficient.
- **Starvation Risk**: We accept background task starvation (e.g. replayer lag) during Red Zones to save the host from crash.

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/vault`, `github.com/sungp/gophership/internal/stochastic`
- **Dependency**: Background loops in `internal/vault/replay.go`.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-4.3) - Functional Requirement FR8.
- [internal/stochastic/monitor.go](../../internal/stochastic/monitor.go) - Source of truth for status.
- [internal/vault/replay.go](../../internal/vault/replay.go) - Targeted background component.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

- Implemented `ThrottleMultiplier` in `internal/stochastic/state.go` with lock-free atomic lookups.
- Integrated dynamic duty cycling into `internal/vault/Replayer`.
- Implemented "Deep Sleep" loop (5s wait) for `StatusRed` events.
- Added Starvation Score metrics and logging for background task yielding.
- Verified behavior via `internal/vault/throttle_replayer_test.go`.

### File List

- `internal/stochastic/state.go`
- `internal/stochastic/throttle_test.go`
- `internal/vault/replay.go`
- `internal/vault/throttle_replayer_test.go`
