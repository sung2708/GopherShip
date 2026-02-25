# Story 6.3: Optimized Container Packaging (K8s Sidecar)

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a DevOps engineer,
I want a multi-stage Docker build for a minimal static binary,
so that I can deploy GopherShip as a lightweight sidecar or DaemonSet in Kubernetes.

## Acceptance Criteria

1. **Minimal Distroless Image**: The final image must contain ONLY the static GopherShip binary. No shell or package managers.
2. **Size Constraint**: The final image size must be less than 20MB (NFR.S2 optimized).
3. **Somatic Health Probes**: The image must include/support health check probes that query the internal somatic state (Green/Yellow/Red).
4. **Hardware Honest Builds**: Supporting Linux/AMD64 and ARM64 targets.
5. **Static Linking**: Binary must be fully statically linked to avoid `glibc` issues in minimal containers.

## Tasks / Subtasks

- [x] Implement Standard gRPC Health Check (AC: #3)
  - [x] Use `google.golang.org/grpc/health/grpc_health_v1`
  - [x] Map internal `stochastic.Status` to gRPC `ServingStatus`
  - [x] Register health server in `internal/control` or `main.go`
- [x] Create Production Dockerfile (AC: #1, #2, #4, #5)
  - [x] Use multi-stage build (`golang:alpine` for build, `gcr.io/distroless/static` for final)
  - [x] Set `CGO_ENABLED=0` and use `-ldflags="-s -w"` for size optimization
  - [x] Include `.dockerignore` to keep context small
- [x] Verification & Benchmarking
  - [x] Build image and verify size (`docker images | grep gophership`)
  - [x] Run container and verify health probe using `grpc-health-probe` or equivalent
  - [x] Verify static linking via `ldd` or `file` command

## Dev Notes

### Architecture Compliance
- **NFR.P1 (Zero-Allocation)**: Health checks should use cached somatic status to prevent allocation spikes during probing.
- **NFR.Sec2 (Management Access)**: Health probes should ideally use the same secure control plane or a dedicated public port if required by K8s.

### Source Tree Components
- `cmd/gophership/main.go`: Register health service.
- `internal/control/server.go`: Integrate health check logic.
- `Dockerfile`: (New) Multi-stage recipe.
- `.dockerignore`: (New) Build context optimization.

### References
- [Architecture Decision Document](../../_bmad-output/planning-artifacts/architecture.md#L101-L102)
- [PRD - Success Criteria](../../_bmad-output/planning-artifacts/prd.md#L119)
- [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)

## Dev Agent Record

### Agent Model Used

Antigravity (Gemini 2.0 Pro)

### Completion Notes List

- Implemented gRPC Health Checking Protocol (v1) mapping stochastic status to Serving/NotServing.
- Created multi-stage Dockerfile using Distroless static base.
- Verified static binary size: 14MB (Requirement < 20MB).
- Verified full static linking (CGO_ENABLED=0).
- Fixed test pollution by adding status cleanup in `health_test.go`.

### File List

- `internal/control/server.go`: Added health service.
- `internal/control/health_test.go`: (New) Tests for health mapping.
- `Dockerfile`: (New) Production image definition.
- `.dockerignore`: (New) Build context optimization.
