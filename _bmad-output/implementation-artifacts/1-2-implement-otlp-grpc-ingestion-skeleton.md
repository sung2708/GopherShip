# Story 1.2: Implement OTLP gRPC Ingestion Skeleton

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an SRE,
I want GopherShip to listen for OTLP/gRPC signals,
so that I can begin sending telemetry to the engine.

## Acceptance Criteria

1. **[AC1]** GopherShip must initialize a gRPC server listening on port 4317 (default OTLP port).
2. **[AC2]** The server must implement the `ExportLogsServiceServer` interface from the OpenTelemetry proto definitions.
3. **[AC3]** Upon receiving a valid `ExportLogsRequest`, the server must acknowledge success (status OK) after handing data to the internal `IngestData` reflex.
4. **[AC4]** Ingestion events must be logged using `rs/zerolog` with structured metadata (trace ID if available).

## Tasks / Subtasks

- [x] Task 1: Initialize OTLP gRPC Service (`internal/ingester`)
  - [x] Implement `ExportLogs` method on `Ingester` struct.
  - [x] Integrate with existing `IngestData` (context-aware reflex).
- [x] Task 2: Setup gRPC Server Lifecycle
  - [x] Create `StartGRPCServer(ctx context.Context, addr string)` in `internal/ingester`.
  - [x] Ensure graceful shutdown integration with the engine's root context.
- [x] Task 3: Engine Integration (`cmd/gophership/main.go`)
  - [x] Initialize `Ingester` and start the gRPC worker loop.
  - [x] Start the gRPC server on port 4317.
- [x] Task 4: Local Verification
  - [x] Run a test using `grpcurl` or a mock OTel producer to verify connectivity.

## Dev Notes

- **Architecture Pattern**: Hybrid Somatic Model (Ingestion Reflex). [Source: architecture.md#Decision-Priority-Analysis]
- **Protocol**: OTLP/gRPC (Port 4317).
- **Graceful Shutdown**: Use `signal.NotifyContext` from `main.go`.
- **Logging**: Use `rs/zerolog` exclusively.
- **Reference Implementation**: Use `go.opentelemetry.io/proto/otlp/collector/logs/v1` for service definitions.

### Project Structure Notes

- Logic remains encapsulated in `internal/ingester`.
- Public gRPC port 4317 is an architectural mandate for OTel interoperability.

### References

- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md) - Section: "Architecture Decisions - API & Communication"
- [epics.md](../../_bmad-output/planning-artifacts/epics.md) - Story 1.2 Details
- [internal/ingester/ingester.go](../../internal/ingester/ingester.go) - Current Skeleton

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

- [grpc_test.go](../../internal/ingester/grpc_test.go)

### Completion Notes List

- Implemented `ExportLogsServiceServer` in `Ingester`.
- Added `StartGRPCServer` with graceful shutdown support.
- Fully integrated OTLP ingestion into `cmd/gophership/main.go`.
- Verified via `TestIngester_OTLPgRPCIngestion`.

### File List

- [internal/ingester/ingester.go](../../internal/ingester/ingester.go)
- [internal/ingester/grpc_test.go](../../internal/ingester/grpc_test.go)
- [cmd/gophership/main.go](../../cmd/gophership/main.go)
- [go.mod](../../go.mod)
- [go.sum](../../go.sum)
