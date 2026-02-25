# Story 6.2: Production TLS 1.3 Ingestion

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security engineer,
I want all network-based ingestion endpoints to enforce TLS 1.3,
so that data in transit is protected by modern cipher suites.

## Acceptance Criteria

1. **[AC1]** The ingestion server MUST reject any protocol version lower than TLS 1.3.
2. **[AC2]** The implementation MUST support modern cipher suites (`TLS_AES_128_GCM_SHA256`, `TLS_AES_256_GCM_SHA384`, `TLS_CHACHA20_POLY1305_SHA256`).
3. **[AC3]** Mutual TLS (mTLS) MUST be supported, allowing the engine to verify client certificates against a provided CA.
4. **[AC4]** TLS configuration integration MUST be implemented for the gRPC/OTLP server in `internal/ingester`.
5. **[AC5]** Performance benchmark MUST verify that TLS termination does not introduce excessive latency beyond the <500μs (P99) reflex limit (NFR.P2).

## Tasks / Subtasks

- [x] TLS Infrastructure Setup (AC: 1, 2)
  - [x] Implement `CreateIngestionTLSConfig` in `pkg/otel/telemetry.go` (or similar helper).
  - [x] Enforce `MinVersion: tls.VersionTLS13`.
  - [x] Configure `ClientAuth` for mTLS support (AC: 3).
- [x] Ingester Integration (AC: 4)
  - [x] Update `internal/ingester/ingester.go` to accept `Credentials` during server initialization.
  - [x] Support loading certificates from paths defined in YAML configuration.
- [x] Verification & Testing (AC: 1, 5)
  - [x] Implement a test in `internal/ingester/ingester_test.go` to verify rejection of TLS 1.2.
  - [x] Run latency benchmarks under TLS load to verify adherence to NFR.P2.

## Dev Notes

- **Architecture Patterns**: Follow the **Authentication & Security** patterns defined in [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Authentication%20&%20Security).
- **Go Best Practices**: For TLS 1.3, Go automatically manages cipher suites. Explicit configuration of `CipherSuites` in `tls.Config` is ignored for version 1.3.
- **Source Tree**:
  - `internal/ingester/ingester.go`: Main server setup.
  - `pkg/otel/telemetry.go`: Shared telemetry/security helpers.

### Project Structure Notes

- GopherShip uses a Standard Go High-Performance Layout. 
- Ensure that certificate loading is robust and fails gracefully if paths are invalid.

### References

- [Architecture: TLS 1.3 Enforcement](../../_bmad-output/planning-artifacts/architecture.md#Decision%20Priority%20Analysis)
- [PRD: NFR.Sec1 - Encryption](../../_bmad-output/planning-artifacts/prd.md#Non-Functional%20Requirements)
- [Otel Best Practices: TLS](https://opentelemetry.io/docs/collector/configuration/#tls)

## Dev Agent Record

### Agent Model Used

Antigravity (BMad-Story-Creator-v1)

### Debug Log References

- Identified `internal/ingester` as the primary TLS termination point (Step 721).
- Confirmed TLS 1.3 mandate from Epic 6.2 (Step 721).

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.

### File List

- internal/ingester/ingester.go
- pkg/otel/telemetry.go
- internal/ingester/ingester_test.go
- internal/ingester/ingester_perf_test.go
- cmd/gophership/main.go
- internal/config/config.go
- go.mod
