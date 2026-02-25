# Story 1.3: Establish Zero-Allocation Buffer Pool (`internal/buffer`)

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a platform engineer,
I want to implement a global `sync.Pool` for byte buffers,
so that the hot path can operate with zero heap allocations.

## Acceptance Criteria

1. **[AC1]** Implement `internal/buffer.MustAcquire(size int) *[]byte` and `internal/buffer.MustRelease(buf *[]byte)`.
2. **[AC2]** The pool MUST store pointers to slices (`*[]byte`) to avoid interface-conversion allocations (Architecture Mandate NFR.P1).
3. **[AC3]** `MustAcquire` must return a buffer of at least the requested size, or create a new one if the pool is empty.
4. **[AC4]** `MustRelease` must reset the slice length to 0 before returning it to the pool to prevent data poisoning.
5. **[AC5]** `go test -benchmem` must show **0 B/op** and **0 allocs/op** for the acquisition/release lifecycle.

## Tasks / Subtasks

- [x] Task 1: Initialize Buffer Package (`internal/buffer`) (AC: #1, #2, #3, #4)
  - [x] Create `pool.go` with global `sync.Pool`.
  - [x] Implement `MustAcquire` and `MustRelease` logic.
- [x] Task 2: Implement Slice Pointer Pattern (AC: #2)
  - [x] Ensure `sync.Pool` handles `*[]byte` to prevent slice header escape.
  - [x] Implement protective checks for buffer capacity.
- [x] Task 3: Performance Validation (AC: #5)
  - [x] Create `pool_test.go` with `BenchmarkPoolAcquisition`.
  - [x] Verify zero-allocation status via `go test -bench`.
- [x] Task 4: [AI-Review][CRITICAL] Hardened `MustRelease` with `magic` number validation to prevent `unsafe` corruption.
- [x] Task 5: [AI-Review][HIGH] Improved `MustAcquire` recycling to prevent small-buffer thrashing.
- [x] Task 6: [AI-Review][MEDIUM] Added safety tests for random pointer and double-release scenarios.

## Dev Notes

- **NFR.P1 Compliance**: Storing `[]byte` in `interface{}` triggers an allocation for the header. Always store `*[]byte`. [Source: Research/Technical-Best-Practices]
- **Architectural Mandate**: Use the `Must` prefix for acquisition functions as defined in `architecture.md#Additional-Requirements`.
- **Buffer Safety**: Slices must be resliced to zero length before `Put`, but their capacity must be preserved.
- **Hardening**: Magic number validation in `MustRelease` ensures pointer arithmetic only executes on verified `pooledBuffer` structures.

### Project Structure Notes

- Package: `github.com/sungp/gophership/internal/buffer`
- Location: `internal/buffer/pool.go`

### References

- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md) - Section: "Architecture Decisions - Memory & Performance"
- [epics.md](../../_bmad-output/planning-artifacts/epics.md) - Story 1.3 Details

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

- Successfully implemented pointers-to-slice strategy for `sync.Pool` to achieve zero allocations.
- Verified 0 B/op and 0 allocs/op via Go benchmarks (~14ns/op acquisition).
- Integrated `MustAcquire` and `MustRelease` with protective capacity checks.
- ✅ Resolved review finding [CRITICAL]: Hardened unsafe arithmetic with magic number validation.
- ✅ Resolved review finding [HIGH]: Implemented thrash prevention in `MustAcquire` via multi-stage Get.
- ✅ Resolved review finding [MEDIUM]: Added comprehensive safety and stress tests.

### File List

- [internal/buffer/pool.go](../../internal/buffer/pool.go)
- [internal/buffer/pool_test.go](../../internal/buffer/pool_test.go)
