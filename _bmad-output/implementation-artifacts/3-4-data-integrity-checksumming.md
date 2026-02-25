# Story 3.4: Data Integrity & Checksumming

Status: done



<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a compliance officer,
I want cryptographic checksums for every WAL segment,
so that I can guarantee the bit-identical integrity of replayed logs.

## Acceptance Criteria

1. **[AC1]** The `internal/vault` package MUST compute a CRC32 checksum for every data block before it is written to the WAL.
2. **[AC2]** The checksum MUST be stored within the binary framing (extending the header to 16 bytes).
3. **[AC3]** The `Replayer` MUST verify the checksum of every block before passing it to the ingestion sink.
4. **[AC4]** Any checksum mismatch MUST trigger a critical alert (log level ERROR or above) and quarantine the affected segment (stop replay for that segment).
5. **[AC5]** Verification MUST be near zero-allocation by reusing a `hash.Hash32` pool.

## Tasks / Subtasks

- [x] Task 1: Update Binary Framing (AC: #1, #2)
  - [x] Update `HeaderSize` to 16 in `compress.go`.
  - [x] Implement `crcPool` in `compress.go` using `sync.Pool` for `crc32.NewIEEE()`.
  - [x] Update `CompressBlock` to compute and store CRC32 of the compressed payload.
- [x] Task 2: Implement Verification Logic (AC: #3)
  - [x] Update `DecompressBlock` to verify the CRC32 before decompression.
  - [x] Update `Replayer` to handle decompression errors as integrity failures.
- [x] Task 3: Error Handling & Quarantining (AC: #4)
  - [x] Implement logic in `Replayer` to stop segment processing on first checksum failure.
  - [x] Ensure the error is bubbled up to the `Ingester` and logged as a critical failure.
- [x] Task 4: [VERIFICATION] Corruption Resistance
  - [x] Create `integrity_test.go` that manually corrupts a segment and verifies it's detected.
  - [x] Verify zero-allocation using `-benchmem`.

## Dev Notes

- **Header Update**: The new header structure is `[Magic:4][UncompLen:4][CompLen:4][CRC32:4]`.
- **Zero-Allocation**: Reuse the `crc32.Hash32` object. Call `Reset()` before each use.
- **Quarantine**: For this story, "quarantine" means logging the absolute path of the corrupted segment and skipping the remainder of that segment to prevent corrupting the somatic state.

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/vault`
- **Integrity Baseline**: This is a direct extension of the `vault` package established in Stories 3.1 and 3.2.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-3.4) - Checksum Requirement.
- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Integrity) - Architecture decision on CRC32.
- [internal/vault/compress.go](../../internal/vault/compress.go) - Target for framing update.
- [internal/vault/replay.go](../../internal/vault/replay.go) - Target for verification integration.

## Developer Context

### Story 3.3 Learnings
- **Synchronous Flushing**: Ensure that the `WAL` always flushes the current block before rotation, or the checksum won't cover trailing data.
- **Replayer Logic**: The `Replayer` uses `mmap` peeking. Ensure the header peek logic is updated for the new `HeaderSize`.

### Architecture Compliance
- Use `hash/crc32` with `crc32.IEEE` polynomial.
- Adhere to the `Must` prefix for all buffer-pooling operations if new ones are added.

## Dev Agent Record

### Agent Model Used

Antigravity

### File List

- [internal/vault/compress.go](../../internal/vault/compress.go)
- [internal/vault/replay.go](../../internal/vault/replay.go)
- [internal/vault/wal.go](../../internal/vault/wal.go)
- [internal/vault/integrity_test.go](../../internal/vault/integrity_test.go)

### Code Review (AI-Review)
- **Status**: Findings Addressed
- **Findings**:
    - [x] Fixed redundant binary peeking in `replay.go` by refactoring `DecompressBlock` to return frame metadata.
    - [x] Updated story documentation to reflect actual implementation status (Tasks/Subtasks).
    - [x] Included `wal.go` in the File List.
