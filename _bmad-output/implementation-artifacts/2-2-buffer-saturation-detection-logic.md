# Story 2.2: Buffer Saturation Detection Logic

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an engine,
I want to detect buffer saturation in <1ms,
so that I can trigger the somatic pivot before upstream backpressure occurs.

## Acceptance Criteria

1. **[AC1]** The `somatic.Controller` MUST be able to evaluate the current occupancy of the `ingester.Ingester` buffer channel.
2. **[AC2]** The pivot decision (Reassess) MUST be calculated in <1ms to prevent blocking the ingestion pipeline.
3. **[AC3]** The detection logic MUST satisfy the "Stochastic Awareness" mandate: No global locks may be held during the saturation check.
4. **[AC4]** The `stochastic.AmbientStatus` MUST transition to `StatusRed` when the buffer exceeds a specific "High Watermark" (default 85%).
5. **[AC5]** Verify that the detection loop does not cause measurable overhead in the `BenchmarkPivotLatency` test suite.

## Tasks / Subtasks

- [x] Task 1: Implement Ingester Depth Exposure (AC: #1, #3)
  - [x] Add `BufferDepth()` and `BufferCap()` methods to `internal/ingester/ingester.go` returning current channel stats.
- [x] Task 2: Implement Somatic Saturation Logic (AC: #2, #4)
  - [x] Refactor `internal/somatic/controller.go` to accept an `Ingester` reference (or a Provider interface).
  - [x] Implement `Reassess()` logic: If `depth / cap > 0.85` -> `StatusRed`, else `StatusGreen`.
- [x] Task 3: Integrate Controller into Ingester Hot Path (AC: #3)
  - [x] Update `IngestData` to trigger `controller.Reassess()` every 1024 operations (reuse `processedCount`).
- [x] Task 4: [SAFETY] Implement hysteresis (cooldown) to prevent Rapid Status Oscillations.

## Dev Notes

- **Stochastic Balance**: Buffer sensing is expensive if done on every packet. Reuse the 1024-op stochastic trigger established in Story 2.1.
- **Hysteresis**: Once in `StatusRed`, don't flip back to `StatusGreen` until the buffer drops below 20% (the "Low Watermark").
- **No Defer/No Locks**: The assessment path must be extremely lean. Use simple arithmetic on raw integers.

### Project Structure Notes

- Package: `github.com/sungp/gophership/internal/somatic`
- Package: `github.com/sungp/gophership/internal/stochastic`
- Dependency: `somatic` depends on `ingester` for depth sensing.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-2.2) - Functional Requirement FR2
- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Concurrecy-&-Reflex-Patterns) - Stochastic Awareness Invariants
- [ingester.go](../../internal/ingester/ingester.go) - Current 1024-op trigger implementation.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

- Verified 0 allocations in `TestIngester_ZeroAllocationPivot`.
- Verified StatusRed trigger at 85% occupancy and recovery at 20% in `internal/somatic/controller_test.go`.
- **Code Review Fixes**: Implemented atomic state, lazy status updates, and replaced float math with integer thresholds.
- **Safety Tradeoff**: Nulling the slice pointer in `IngestData` (`*data = nil`) was identified as causing heap escapes (2 allocations) on the Windows toolchain. It has been omitted in favor of strict `NFR.P1` compliance; ownership handover is now a documented invariant.

### Completion Notes List

- [x] Implemented `BufferDepth()` and `BufferCap()` in `Ingester`.
- [x] Implemented `somatic.Controller` with hysteresis (85% high / 20% low).
- [x] Integrated `Controller` into `Ingester` hot path with 1024-op stochastic trigger.
- [x] Verified sub-microsecond latency and zero allocations.

### File List

- [MODIFY] [internal/ingester/ingester.go](../../internal/ingester/ingester.go)
- [MODIFY] [internal/somatic/controller.go](../../internal/somatic/controller.go)
- [NEW] [internal/somatic/controller_test.go](../../internal/somatic/controller_test.go)
### Senior Developer Review (AI)

- **Outcome**: APPROVED with fixes applied.
- **Performance**: Confirmed sub-microsecond latency and zero-allocation hot path.
- **Safety**: Atomic state transitions and lazy updates implemented to eliminate cache-line bouncing.
