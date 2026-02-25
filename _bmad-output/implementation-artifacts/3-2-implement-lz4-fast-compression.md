# Story 3.2: Implement LZ4 Fast Compression

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a platform engineer,
I want the Raw Vault to use LZ4 compression,
so that I can maximize disk throughput without saturating the CPU.

## Acceptance Criteria

1. **[AC1]** Segments MUST be compressed using `github.com/pierrec/lz4/v4`.
2. **[AC2]** The implementation MUST use a block-based compression strategy (e.g., 64KB blocks) to allow for efficient partial replay.
3. **[AC3]** Write throughput MUST remain within 10% of raw disk I/O speed (validated via benchmarks).
4. **[AC4]** Decompression logic MUST be implemented to verify data integrity in tests.

## Tasks / Subtasks

- [x] Task 1: Integrate LZ4 Dependency (AC: #1)
  - [x] Add `github.com/pierrec/lz4/v4` to `go.mod`.
  - [x] Research and select optimal block size (default 64KB).
- [x] Task 2: Implement Block-Based Framing (AC: #1, #2)
  - [x] Define a binary block header (e.g., `[Magic:4][UncompressedSize:4][CompressedSize:4]`).
  - [x] Modify `MustWrite` to accumulate data into a block buffer before compressing.
- [x] Task 3: Update WAL to handle Compressed Streams (AC: #2, #3)
  - [x] Implement compressed block flushing to `mmap` segments.
  - [x] Ensure `MustWrite` still follows zero-allocation patterns (pooling compression buffers).
- [x] Task 4: [VERIFICATION] Benchmarking & Integrity (AC: #3, #4)
  - [x] Implement a basic `Reader` for verification tests.
  - [x] Compare `BenchmarkWAL_MustWrite` results with and without compression.

## Dev Notes

- **Zero-Allocation**: Use `sync.Pool` for both the "Accumulation Block" (uncompressed) and the "Output Block" (compressed).
- **Latency**: Compression should happen within the `rotateLocked` or a dedicated block-flush call to minimize hot-path lock time.
- **Framing**: Each segment should start with a 4-byte Magic to identify LZ4-compressed files vs. legacy raw files if needed (though Story 3.1 didn't have a header, we should establish one now).

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/vault`
- **Dependencies**: `pierrec/lz4/v4`, `internal/buffer`.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-3.2) - LZ4 Requirement.
- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Data-Architecture) - Hybrid Compression Decision.
- [wal.go](../../internal/vault/wal.go) - Current WAL implementation.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

- Fixed segment size panic: NewWAL now validates `segmentSize >= DefaultBlockSize + HeaderSize`.
- Verified compression integrity with `TestWAL_LargeWrite`.

### Completion Notes List

- Implemented binary framing with Magic `0x564C5A34`.
- Integrated `pierrec/lz4/v4` with zero-allocation pooling.
- Verified WAL rotation and large write spanning across segments.

### File List

- [internal/vault/wal.go](../../internal/vault/wal.go)
- [internal/vault/compress.go](../../internal/vault/compress.go)
- [internal/vault/compress_test.go](../../internal/vault/compress_test.go)
- [internal/vault/wal_test.go](../../internal/vault/wal_test.go)