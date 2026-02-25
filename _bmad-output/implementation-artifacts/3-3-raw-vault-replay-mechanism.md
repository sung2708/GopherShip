# Story 3.3: Raw Vault Replay Mechanism

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an SRE,
I want to "replay" the Raw Vault once a traffic spike subsides,
so that I can recover all logs for deferred parsing.

## Acceptance Criteria

1. **[AC1]** The `internal/vault` package MUST provide a `Replayer` or `Scanner` that can iterate through multiple WAL segments in chronological order.
2. **[AC2]** The replayer MUST correctly decompress blocks using the binary framing defined in Story 3.2 (`0x564C5A34`).
3. **[AC3]** The replay rate MUST be throttleable (e.g., in records per second or MB/s) to ensure replay doesn't re-trigger the "Red Zone" somatic pivot.
4. **[AC4]** Replayed data MUST be bit-identical to the original data (NFR.R2).

## Tasks / Subtasks

- [x] Task 1: Implement Segment Iterator (AC: #1)
  - [x] Implement `ListSegmentsOrdered` in `wal.go`.
  - [x] Create `Replayer` struct that tracks current segment and offset.
- [x] Task 2: Implement Block Decoding Logic (AC: #2, #4)
  - [x] Integrate with `vault.DecompressBlock`.
  - [x] Handle EOF and segment transitions gracefully.
- [x] Task 3: Throttling & Ingestion Integration (AC: #3)
  - [x] Add `RateLimiter` interface or simple `t.Sleep` logic to `Replayer`.
  - [x] Implement a `ReplayAll` or `StreamTo(sink func([]byte))` method.
- [x] Task 4: [VERIFICATION] Integrity & Performance
  - [x] Create `replay_test.go` verifying multiple compressed segments.
  - [x] Benchmark replay speed at different throttle levels.

## Dev Notes

- **Zero-Allocation**: Replay iteration MUST use `sync.Pool` for decompression buffers. Avoid re-allocating `[]byte` for every record.
- **Throttling**: Use `time.Ticker` for a steady stream rather than bursts, as it's more "Somatic Friendly".
- **Integrity**: Each block should be verified against its header before being passed to the sink.

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/vault`
- **Interface Interaction**: Replay should feed back into the `Ingester` via its internal channel or a dedicated `ReplayIngest` method.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-3.3) - Replay Requirement.
- [3-2-implement-lz4-fast-compression.md](../../internal/vault/compress.go) - LZ4 Framing implementation.
- [wal.go](../../internal/vault/wal.go) - Target for iteration logic.

## Senior Developer Review (AI)

- **Documentation Sync**: Fixed. Story status set to 'done' and tasks marked [x] to match implementation.
- **Technical Debt**: Refactored `replay.go` to use standard `encoding/binary` and optimized zero-skipping.
- **Security**: Validated decompression framing against AC2.
- **Completeness**: Added `internal/vault/compress.go` to the File List.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

### File List

- [internal/vault/wal.go](../../internal/vault/wal.go)
- [internal/vault/replay.go](../../internal/vault/replay.go)
- [internal/vault/replay_test.go](../../internal/vault/replay_test.go)
- [internal/vault/compress.go](../../internal/vault/compress.go)
- [internal/vault/compress_test.go](../../internal/vault/compress_test.go)

