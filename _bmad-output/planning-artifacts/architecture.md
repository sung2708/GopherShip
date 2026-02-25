---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments: 
  - prd.md
  - product-brief-GopherShip-2026-02-24.md
  - market-high-performance-log-middleware-2026-02-24.md
  - brainstorming-session-2026-02-24.md
workflowType: 'architecture'
project_name: 'GopherShip'
user_name: 'sungp'
date: '2026-02-24'
lastStep: 8
status: 'complete'
completedAt: '2026-02-24'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**
Architecturally, GopherShip must maintain a non-blocking "Hot Path" for ingestion that can instantly pivot to a "Debt Path" (Raw Vault) without logical queuing delays.

**Non-Functional Requirements:**
The targets of **<500μs reflex latency** and **Zero-Allocation** drive the need for a pooled memory architecture and a lock-free status monitoring system (Stochastic Awareness).

**Scale & Complexity:**
High-complexity observability infrastructure. Requires coordination of 5 core components: Ingester, Somatic Controller, Raw Vault/WAL, Stochastic Monitor, and CLI Control.

### Technical Constraints & Dependencies
- **Go 1.22+**: Utilizing latest concurrency and memory primitives.
- **Linux Kernel 5.8+**: Necessary for high-performance I/O and future eBPF collection.
- **OTel Compliance**: Strict mapping to the OpenTelemetry Log Data Model.

### Cross-Cutting Concerns Identified
- **Memory Budgeting**: Global and per-component limits to prevent host OOM.
- **Cache Local State**: Minimizing global atomic synchronization to maintain core density.

## Starter Template Evaluation

### Primary Technology Domain
**Systems Middleware / CLI Tool** based on requirements for zero-loss ingestion and hardware-honest reliability.

### Starter Options Considered
1. **Standard Go Layout**: Clean separation of `cmd/`, `internal/`, and `pkg/`. Best for performance-critical engines.
2. **Cobra-CLI Starter**: Optimized for building the `gs-ctl` management utility.
3. **Internal Skeleton**: Custom initialization using `sync.Pool` and NUMA-aware worker groups.

### Selected Starter: Standard Go High-Performance Layout
**Rationale for Selection:**
GopherShip's "biological" nature requires absolute control over memory and CPU scheduling. A standard layout allows us to enforce strict `internal/` encapsulation while providing a clear entry point for both the daemon and CLI.

**Initialization Command:**
```bash
# Core manual setup for high-performance encapsulation
mkdir -p cmd/gophership cmd/gs-ctl internal/{ingester,vault,somatic,stochastic} pkg/otel
go mod init github.com/sungp/gophership
```

**Architectural Decisions Provided by Starter:**

- **Language & Runtime**: **Go 1.22+** utilizing `sync.Pool` for zero-allocation hotspots.
- **Code Organization**: 
  - `cmd/`: Thin entry points for executables.
  - `internal/`: Private core logic (Ingester, Raw Vault, Stochastic Monitor).
  - `pkg/`: Public OTel compliance logic and shared primitives.
- **Concurrency**: NUMA-aware worker pools with `select-default` reflex paths.
- **Build Tooling**: Static binary compilation for dependency-free distribution.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- Custom Go-native WAL with `mmap` and `sync.Pool` for zero-allocation durability.
- OTLP gRPC as the uniform protocol for both ingestion and management.

**Important Decisions (Shape Architecture):**
- Hybrid compression: LZ4 for high-speed local storage; Zstd for cloud archival.
- Kernel-level `SO_PEERCRED` authentication for local management CLI.
- TLS 1.3 enforcement for all network-based ingestion.

### Data Architecture
- **WAL Implementation**: Custom implementation using O_DIRECT/mmap to bypass Go's heap during the reflex path.
- **Compression**: `pierrec/lz4` for the Raw Vault (Hardware Speed); `klauspost/compress/zstd` for background sync to deep storage.
- **Integrity**: CRC32 checksums stored alongside WAL entries to ensure replay bit-identity.

### Authentication & Security
- **Ingestion**: TLS 1.3 required. Optional mTLS configuration for secure tunnels.
- **CLI Management**: Unix Domain Socket with filesystem permissions (0600) and `SO_PEERCRED` peer authentication.

### API & Communication Patterns
- **Ingestion**: OTLP over gRPC using standard OTel Go SDK interceptors.
- **Management API**: gRPC over Unix Sockets or TCP Port 9092. Shares the same protobuf definitions as ingestion for consistency.
- **Observability**: Prometheus metrics export on port 9091 (default).

### Infrastructure & Deployment
- **Packaging**: Static Go binary for Linux/ARM64.
- **Scaling**: Horizontal scaling via Sidecar/DaemonSet pattern; internal scaling via cache-local worker groups.
- **Self-Observability**: GopherShip will use its own "Somatic" telemetry to monitor its internal pressure zones.

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:**
4 key areas to ensure hardware-honesty and zero-allocation consistency across AI agents.

### Naming Patterns
- **Go Symbols**: `CamelCase` for exported, `camelCase` for internal.
- **Hot Path Signaling**: Functions in the high-performance ingestion path MUST use the `Must` prefix (e.g., `MustAcquireBuffer`) to indicate `sync.Pool` usage and zero-allocation constraints.
- **Errors**: Static `Err` variables (e.g., `ErrBufferFull`) for hot paths; no dynamic `fmt.Errorf` allowed in reflexes.

### Zero-Allocation Patterns (NFR.P1)
- **Object Reuse**: All `[]byte` buffers MUST be acquired from a global `sync.Pool` and returned immediately after use.
- **No String Churn**: Use `[]byte` for all metadata and header lookups until the deferred parsing stage.
- **Escape Analysis**: Avoid interfaces and closure captures in hot path loops to prevent heap escapes.

### Concurrency & Reflex Patterns
- **Local Reflex**: Every ingestion worker MUST implement a `select-default` block. The `default` branch MUST trigger the `Raw Vault` fallback immediately (non-blocking).
- **Stochastic Awareness**: Environment sensing (memory pressure, buffer depth) MUST be "lazy"—triggered only every `N` operations (default 1024) to eliminate atomic contention.

### Observability Patterns
- **Logging**: Use `rs/zerolog` exclusively for structured, zero-allocation logging.
- **Metrics**: Prometheus `Counter` and `Gauge` naming strategy: `gophership_{component}_{metric}_{unit}`.
- **Tracing**: OTel span decoration only on the deferred parsing path; minimal instrumentation on the ingestion reflex.

### Enforcement Guidelines
**All AI Agents MUST:**
- Run `go test -bench . -benchmem` on ingestion modules to verify zero-allocation.
- Adhere to the `internal/` package encapsulation to protect the "biological" core.
- Implement the "Linear Scaling" benchmark for all new concurrency workers.

## Project Structure & Boundaries

### Complete Project Directory Structure

```
gophership/
├── cmd/
│   ├── gophership/          # The Resilient Daemon (Ingester)
│   │   └── main.go
│   └── gs-ctl/              # Management CLI
│       └── main.go
├── internal/
│   ├── ingester/            # Reflex logic & sync.Pool
│   ├── vault/               # WAL, mmap, LZ4 vault core
│   ├── somatic/             # The Pivot Controller (Reflex triggers)
│   ├── stochastic/          # Lazy Atomic state monitoring
│   ├── control/             # gRPC-over-Unix management service
│   └── buffer/              # Shared high-performance primitives
├── pkg/
│   ├── protocol/            # Protobuf definitions (Ingestion + Management)
│   └── otel/                # OTel data model mapping
├── scripts/                 # Performance benchmarking & stress tests
├── go.mod
└── README.md
```

### Architectural Boundaries

**API Boundaries:**
- **External**: OTLP/gRPC on Port 4317; NDJSON/Syslog on dynamic ports.
- **Internal**: `gophership.sock` (Unix) for `gs-ctl` communication, secured by `SO_PEERCRED`.

**Component Boundaries:**
- **Hot-Path**: Zero-allocation channels between `ingester` and `somatic`.
- **Debt-Path**: Asynchronous bit-stream from `somatic` to `vault` (WAL).

### Requirements to Structure Mapping

**Feature/Epic Mapping:**
- **Somatic Ingestion** (`FR1-FR3`): `internal/ingester/`
- **Raw Data Preservation** (`FR4-FR6`): `internal/vault/`
- **Hardware Optimizer** (`FR7-FR9`): `internal/stochastic/`
- **OTel & Interop** (`FR10-FR12`): `pkg/otel/` and `pkg/protocol/`
- **Management & CLI** (`FR13-FR15`): `cmd/gs-ctl/` and `internal/control/`

**Cross-Cutting Concerns:**
- **Environment Sensing**: Shared logic in `internal/stochastic/`.

## Architecture Validation Results

### Coherence Validation ✅
- **Decision Compatibility**: All technology choices (Go 1.22+, LZ4, gRPC) are internally consistent and optimized for systems middleware.
- **Pattern Consistency**: Zero-allocation patterns are physically supported by the modular `internal/` package structure.
- **Structure Alignment**: The separation of `ingester`, `vault`, and `stochastic` components prevents logic leaks and enables granular performance tuning.

### Requirements Coverage Validation ✅
- **Functional Requirements**: 100% coverage mapped to specific entry points and internal packages.
- **Non-Functional Requirements**: The design explicitly addresses the **<500μs reflex latency** via pooled buffers and **security** via kernel-level authentication.

### Implementation Readiness Validation ✅
- **Decision Completeness**: Critical durability and integration decisions are finalized.
- **Structure Completeness**: A full Go-standard directory tree is defined.
- **Pattern Completeness**: Clear rules for naming, concurrency, and error handling are established.

### Architecture Completeness Checklist
- [x] Project context thoroughly analyzed
- [x] Technical constraints (Go 1.22+, Linux 5.8+) identified
- [x] Custom WAL and Hybrid Compression decisions recorded
- [x] Zero-allocation and non-blocking reflex patterns defined
- [x] Complete directory structure and boundaries mapped
- [x] Validation of coherence and requirements coverage complete

### Architecture Readiness Assessment
**Overall Status: READY FOR IMPLEMENTATION**
**Confidence Level: HIGH**

**Key Strengths:**
- **Hardware-Honest Design**: Minimal abstraction on the hot path ensures predictable performance.
- **Biological Resilience**: Built-in fallback to Raw Vault prevents system failure under load.

### Implementation Handoff
**AI Agent Guidelines:**
- Follow zero-allocation patterns strictly using the global `sync.Pool`.
- Adhere to `internal/` encapsulation; never expose "biological" primitives to external packages.
- Implement benchmarking alongside every core module to verify hardware-honest scaling.

**First Implementation Priority:**
Initialize the project structure and establish the primary OTel gRPC listener in `internal/ingester/`.
