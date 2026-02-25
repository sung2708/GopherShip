# Story 5.3: Real-time Somatic Dashboard (gs-ctl top)

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a lead SRE,
I want a live dashboard of engine health directly in my terminal,
so that I can observe "Hardware Honest" metrics during a traffic surge.

## Acceptance Criteria

1. **[AC1]** `gs-ctl top` MUST start a live-updating terminal dashboard.
2. **[AC2]** The dashboard MUST display:
   - System/Engine Goroutine count.
   - Real-time Memory Pressure Score (0-100%).
   - Current Somatic Zone (GREEN/YELLOW/RED).
   - Heap Object count.
3. **[AC3]** The dashboard MUST refresh automatically. The refresh rate MUST be configurable via a `--refresh` flag (default: 1s).
4. **[AC4]** The implementation SHOULD use a gRPC server-side streaming RPC for efficient telemetry delivery, minimizing polling overhead.
5. **[AC5]** The dashboard MUST handle terminal resizing gracefully without crashing or corrupting display output.

## Tasks / Subtasks

- [x] Task 1: Protobuf & Service Update (AC: #4)
  - [x] Add `WatchSomaticStatus` streaming RPC to `control.proto`.
  - [x] Update `internal/control` to implement the streamer, publishing updates at the requested interval.
- [x] Task 2: Live Dashboard Engine (AC: #1, #5)
  - [x] Implement terminal UI logic in `cmd/gs-ctl/dashboard.go`.
  - [x] Use a lightweight approach (e.g., ANSI cursor positioning or a minimal TUI library) aligned with zero-allocation goals.
- [x] Task 3: Dashboard Integration (AC: #2, #3)
  - [x] Connect the `gs-ctl top` command to the streaming gRPC client.
  - [x] Implement the `--refresh` flag parsing and pass it to the stream request.
- [x] Task 4: [VERIFICATION] Live Stress Observation
  - [x] Verify that the dashboard accurately reflects zone transitions during a simulated traffic surge.
  - [x] Verify that Goroutine counts and memory pressure correlate with engine activity.

## Dev Notes

- **Hardware Honesty**: Use server-side streaming to avoid "Atomic Wall" polling contention.
- **Zero-Allocation**: The dashboard rendering loop should be careful not to create significant heap churn.
- **Library Recommendation**: Consider `github.com/rivo/tview` for structured boxes or raw ANSI if the UI is minimal.
- **Concurrency**: Ensure the TUI goroutine and the gRPC receiver goroutine communicate safely without deadlocks.

### Project Structure Notes

- New file: `cmd/gs-ctl/dashboard.go` for UI components.
- Main entry point: `cmd/gs-ctl/main.go` will invoke the dashboard engine.

### References

- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Management-API-Patterns) - Streaming patterns.
- [pkg/protocol/control.proto](../../pkg/protocol/control.proto) - Base status messages.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

- Implemented gRPC server-side streaming for telemetry.
- Developed the `gs-ctl top` TUI using `tview`.
- Robust mTLS and UDS connection handling refactored into `dialControlPlane`.
- Completed build verification for all components.

### File List

- pkg/protocol/control.proto
- pkg/protocol/control.go
- internal/control/server.go
- internal/control/server_test.go
- internal/control/security_linux.go
- internal/control/security_stub.go
- cmd/gs-ctl/main.go
- cmd/gs-ctl/dashboard.go
