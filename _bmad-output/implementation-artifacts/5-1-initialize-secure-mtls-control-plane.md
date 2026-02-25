# Story 5.1: Initialize Secure mTLS Control Plane

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security officer,
I want a gRPC control plane secured by Mutual TLS (mTLS),
so that only authorized administrators can execute sensitive management commands.

## Acceptance Criteria

1. **[AC1]** GopherShip MUST initialize a gRPC server in `internal/control` that listens for management commands.
2. **[AC2]** The control plane MUST support Mutual TLS (mTLS) authentication for all network-based gRPC calls, ensuring both client and server identities are verified.
3. **[AC3]** For local CLI access, the server MUST listen on a Unix Domain Socket (default: `/tmp/gophership.sock`) and implement `SO_PEERCRED` (on Linux) to verify the calling process's UID/GID.
4. **[AC4]** The server MUST implement a base `ControlService` defined in `pkg/protocol/control.proto` with at least a `Heartbeat` and `Status` method.
5. **[AC5]** The `gs-ctl` client MUST be updated to support connecting via either mTLS (remote) or Unix socket (local) with appropriate security credentials.

## Tasks / Subtasks

- [x] Task 1: Define Management Protobufs (`pkg/protocol/control.proto`) (AC: #4)
  - [x] Define `ControlService` with `Ping(Empty) returns (PingResponse)` and `GetSomaticStatus(Empty) returns (StatusResponse)`.
  - [x] Ensure `StatusResponse` includes current Somatic Zone (Green/Yellow/Red) and basic telemetry.
  - [x] Generate Go gRPC bindings using `protoc`. (Note: Manually implemented in pkg/protocol/control.go due to protoc absence)
- [x] Task 2: Implement Secure gRPC Server (`internal/control`) (AC: #1, #2, #3)
  - [x] Implement `ControlServiceServer` interface in `internal/control/server.go`.
  - [x] Implement mTLS listener logic using `crypto/tls` and `credentials.NewTLS`.
  - [x] Implement Unix Domain Socket listener with `SO_PEERCRED` validation (handling Linux vs others).
- [x] Task 3: Engine Integration and Lifecycle (AC: #1)
  - [x] Integrate `control.Server` into `cmd/gophership/main.go` worker group.
  - [x] Ensure the control plane shuts down gracefully upon context cancellation.
- [x] Task 4: CLI Foundation (`cmd/gs-ctl`) (AC: #5)
  - [x] Implement a basic `gs-ctl status` command.
  - [x] Add flags for `--socket` path and `--cert/--key/--ca` for mTLS connection.
- [x] Task 5: [VERIFICATION] Security Audit
  - [x] Verify that a client without a certificate is rejected.
  - [x] Verify that a client with an unauthorized certificate is rejected.
  - [x] Verify that `gs-ctl` works over the Unix socket with correct local permissions.

## Dev Notes

- **Architecture Compliance**: Management API MUST share the same protobuf definitions as ingestion for consistency where possible. [Source: architecture.md#API-Communication-Patterns]
- **Unix Permissions**: `gophership.sock` must be restricted to `0600` or `0660` as per NFR.Sec2.
- **Library Requirements**: Use `google.golang.org/grpc` for all control plane communication.
- **Stochastic Awareness**: The `Status` command should query `stochastic.GetAmbientStatus()` (atomic load) to report zone health.

### Project Structure Notes

- Package: `github.com/sungp/gophership/internal/control`
- Protobuf Path: `pkg/protocol/control.proto`
- Output: `implementation-artifacts/5-1-initialize-secure-mtls-control-plane.md`

### References

- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Authentication-Security) - NFR.Sec2, SO_PEERCRED details.
- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-5.1) - Functional requirement mapping.
- [internal/control/server.go](../../internal/control/server.go) - Target implementation file.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

### File List

- [pkg/protocol/control.proto](../../pkg/protocol/control.proto)
- [pkg/protocol/control.go](../../pkg/protocol/control.go)
- [internal/control/server.go](../../internal/control/server.go)
- [internal/control/security_linux.go](../../internal/control/security_linux.go)
- [internal/control/security_stub.go](../../internal/control/security_stub.go)
- [internal/control/security_test.go](../../internal/control/security_test.go)
- [cmd/gs-ctl/main.go](../../cmd/gs-ctl/main.go)
- [internal/stochastic/monitor.go](../../internal/stochastic/monitor.go) (Modified for telemetry)
