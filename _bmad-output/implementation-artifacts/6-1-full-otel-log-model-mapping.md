# Story 6.1: Full OTel Log Model Mapping

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a data engineer,
I want our ingested logs to strictly follow the OpenTelemetry Log Data Model,
so that GopherShip is compatible with the wider observability ecosystem.

## Acceptance Criteria

1. **[AC1]** Ingested logs MUST strictly follow the `ResourceLogs` and `ScopeLogs` structures of the OTel Log Model.
2. **[AC2]** Standard attributes MUST be correctly mapped:
   - `TimeUnixNano` (Timestamp)
   - `SeverityNumber` and `SeverityText`
   - `Body` (AnyValue)
   - `Attributes` (Key/Value pairs)
3. **[AC3]** Mapping logic MUST reside in `pkg/otel/` as a reusable library.
4. **[AC4]** Implementation MUST adhere to **NFR.P1 (Zero-Allocation)**, ensuring field mapping does not trigger heap churn.

## Tasks / Subtasks

- [x] Initialize OTel Model Types (AC: 1, 3)
  - [x] Define internal `LogRecord` mapping structs if necessary, or leverage `go.opentelemetry.io/proto/otlp/logs/v1`.
  - [x] Implement `ResourceLogs` and `ScopeLogs` encapsulation logic.
- [x] Implement Zero-Allocation Field Mapping (AC: 2, 4)
  - [x] Create mapping functions for `Severity`, `Timestamp`, and `Attributes`.
  - [x] Use `sync.Pool` or pooled buffers for attribute processing to satisfy NFR.P1.
- [x] Verification & Testing (AC: 1, 2)
  - [x] Implement unit tests in `pkg/otel/` to verify bit-perfect mapping against OTel JSON examples.
  - [x] Run `go test -benchmem` to verify zero-allocation compliance in the mapping path.

## Dev Notes

- **Architecture Patterns**: Follow the **Zero-Allocation Pattern** (NFR.P1) defined in [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Zero-Allocation%20Patterns%20(NFR.P1)).
- **Encapsulation**: Public logic belongs in `pkg/otel/`. Avoid dependency on `internal/` packages from `pkg/`.
- **Source Tree**:
  - `pkg/otel/telemetry.go`: Update or add new mapping files here.
  - `pkg/protocol/`: Reference OTLP proto definitions.

### Project Structure Notes

- GopherShip uses a Standard Go High-Performance Layout. 
- Ensure `pkg/otel` remains a clean, public-facing interface for OTel interop.

### References

- [Architecture: OTel Compliance](../../_bmad-output/planning-artifacts/architecture.md#OTel%20Compliance)
- [PRD: OTel Compliance](../../_bmad-output/planning-artifacts/prd.md#OTel%20Compliance)
- [Epic 6.1 Definition](../../_bmad-output/planning-artifacts/epics.md#Story%206.1:%20Full%20OTel%20Log%20Model%20Mapping%20(`pkg/otel`))

## Dev Agent Record

### Agent Model Used

Antigravity (BMad-Story-Creator-v1)

### Debug Log References

- Verified `internal/ingester` currently only marshals raw data (Step 148).
- Identified `pkg/otel/telemetry.go` as a stub (Step 134).

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.

### File List

- pkg/otel/telemetry.go
- pkg/otel/mapping.go [NEW]
- pkg/otel/mapping_test.go [NEW]
