---
stepsCompleted: [1, 2, 3, 4]
inputDocuments: 
  - prd.md
  - architecture.md
project_name: 'GopherShip'
---

# GopherShip - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for GopherShip, decomposing the requirements from the PRD and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: **Ingestion Pipeline** can ingest logs via non-blocking reflexes to prevent upstream backpressure.
FR2: **Somatic Engine** can detect buffer saturation in < 1ms to trigger defensive pivots.
FR3: **Somatic Engine** can switch between full enrichment and Raw Vault capture instantly based on pressure zones.
FR4: **Raw Vault** can flush raw, unparsed bytes to a local WAL when in the "Red Zone."
FR5: **Raw Vault** can replay stored segments for deferred parsing once pressure subsides.
FR6: **Raw Vault** can maintain cryptographic checksums for all raw segments to ensure data integrity.
FR7: **Engine** can monitor global environment health via lazy status updates.
FR8: **Somatic Engine** can throttle internal components based on stochastic awareness to eliminate cache contention.
FR9: **Engine** can manage internal memory budgets to prevent host-level OOM events.

### NonFunctional Requirements

- **NFR.P1 - Zero-Allocation**: Zero heap allocations in the "Ingest Reflex" path.
- **NFR.P2 - Reflex Latency**: < 500μs (P99) from wire to somatic buffer.
- **NFR.S1 - Linear Scaling**: Linear throughput scaling on machines up to 128 cores.
- **NFR.S2 - High-Density**: Support for 1M+ LPS using < 2 vCPU cores.
- **NFR.R1 - Zero-Crash**: Process survival with 100% full buffers without OOM or deadlocks.
- **NFR.R2 - Data Integrity**: 100% bit-identical data preservation upon Raw Vault replay.
- **NFR.Sec1 - Encryption**: TLS 1.3 support for all ingestion endpoints.
- **NFR.Sec2 - Management Access**: Unix socket restricted permissions for the control interface.

### Additional Requirements

- **Starter Layout**: Standard Go High-Performance Layout (`cmd/`, `internal/`, `pkg/`).
- **Memory Hot Path**: All `[]byte` buffers MUST use `sync.Pool` with `Must` prefix for acquisition functions.
- **I/O Strategy**: Custom WAL implementation using `O_DIRECT`/`mmap` to bypass Go heap.
- **Compression**: Hybrid model - LZ4 for Raw Vault (speed), Zstd for background cloud sync.
- **CLI Commands**: CLI (`gs-ctl`) supporting `status`, `replay`, and `drain` with Table/JSON/YAML outputs.
- **Configuration**: YAML-based schema for somatic sensitivity thresholds and vault storage limits.
- **Security**: Kernel-level `SO_PEERCRED` for local CLI authentication.
- **Compliance**: Strict mapping to the OpenTelemetry Log Data Model.

### FR Coverage Map

FR1: Epic 2 - Ingestion Pipeline non-blocking reflexes
FR2: Epic 2 - Somatic Engine <1ms saturation detection
FR3: Epic 2 - Somatic Engine instant pressure-based pivots
FR4: Epic 3 - Raw Vault WAL flushing in "Red Zone"
FR5: Epic 3 - Raw Vault deferred parsing replay
FR6: Epic 3 - Raw Vault bit-identical integrity checksums
FR7: Epic 4 - Engine lazy environment status updates
FR8: Epic 4 - Somatic Engine stochastic awareness throttling
FR9: Epic 4 - Engine internal memory budget management

## Epic List

### Epic 1: Foundation & High-Performance Skeleton
Initialize the "Hardware Honest" Go layout and OTel gRPC foundation. Establish the zero-allocation architectural baseline.
**FRs covered:** Foundation for all.

### Epic 2: The Somatic Ingester (Ingestion Reflex)
Implement the core non-blocking ingestion path and `select-default` reflex mechanism. Allows users to ingest 1M+ LPS without blocking.
**FRs covered:** FR1, FR2, FR3.

### Epic 3: Raw Vault & Durable Preservation
Build the custom WAL with `O_DIRECT/mmap` for high-speed "Shock Absorber" capture. Ensures zero data loss during traffic spikes.
**FRs covered:** FR4, FR5, FR6.

### Epic 4: Stochastic Monitor & Stability Optimizer
Implement "Lazy Atomic" monitoring and memory budgeting to prevent host OOMs. Ensures site reliability via hardware limits.
**FRs covered:** FR7, FR8, FR9.

### Epic 5: Control Plane & gs-ctl Management
Develop the secure Unix socket interface and the management CLI (`status`, `replay`, `drain`).
**FRs covered:** CLI Support.

### Epic 6: OTel Compliance & Production Hardening
Finalize OTel data mapping, TLS 1.3 encryption, and container packaging for K8s deployment.
**FRs covered:** NFR.Sec1, NFR.Sec2.

## Epic 1: Foundation & High-Performance Skeleton

Initialize the "Hardware Honest" Go layout and OTel gRPC foundation.

### Story 1.1: Initialize Standard Go Infrastructure Layout
As a platform engineer, I want the standardized GopherShip repository structure, so that I have a consistent environment for performance-critical engine development.

**Acceptance Criteria:**
- **Given** the project has been authorized for initialization
- **When** the project layout is created
- **Then** root directories `cmd/`, `internal/`, and `pkg/` must exist
- **And** `go.mod` must be initialized with the module name `github.com/sungp/gophership`

### Story 1.2: Implement OTLP gRPC Ingestion Skeleton
As an SRE, I want GopherShip to listen for OTLP/gRPC signals, so that I can begin sending telemetry to the engine.

**Acceptance Criteria:**
- **Given** the OTel gRPC server is configured
- **When** GopherShip is started
- **Then** it must listen on port 4317 (default)
- **And** it must successfully receive and acknowledge a valid OTel log signal without process failure

### Story 1.3: Establish Zero-Allocation Buffer Pool (`internal/buffer`)
As a system developer, I want a global `sync.Pool` for byte buffers, so that the hot path can operate with zero heap allocations.

**Acceptance Criteria:**
- **Given** the high-performance ingestion path requires memory buffers
- **When** a buffer is requested via `internal/buffer.MustAcquire`
- **Then** the buffer must be returned from the pool
- **And** `go test -benchmem` must show 0 B/op and 0 allocs/op for the acquisition and release cycle

## Epic 2: The Somatic Ingester (Ingestion Reflex)

Implement the core non-blocking ingestion path and `select-default` reflex mechanism.

### Story 2.1: Implement `select-default` Somatic Pivot
As a system developer, I want the ingestion worker to use a `select-default` pattern, so that it can pivot to the Raw Vault in microseconds if the logical buffer is full.

**Acceptance Criteria:**
- **Given** the `internal/ingester` is processing logs
- **When** the internal buffer channel is full
- **Then** the `default` branch of the `select` block must trigger the fallback path immediately
- **And** no blocking of the ingestion worker must occur

### Story 2.2: Buffer Saturation Detection Logic
As an engine, I want to detect buffer saturation in <1ms, so that I can trigger the somatic pivot before upstream backpressure occurs.

**Acceptance Criteria:**
- **Given** high-traffic ingestion is occurring
- **When** the somatic controller evaluates the buffer state
- **Then** the decision to pivot must be calculated in <1ms
- **And** no global locks may be held during this check (Stochastic Awareness)

### Story 2.3: Somatic Ingestion Metric Reporting
As an SRE, I want to see real-time metrics on "Pressure Zones", so that I can visualize when the engine is falling back to the Raw Vault.

**Acceptance Criteria:**
- **Given** the somatic engine is changing zones (Green/Yellow/Red)
- **When** a zone transition occurs
- **Then** the Prometheus gauge `gophership_ingester_zone` must be updated
- **And** a counter `gophership_somatic_pivots_total` must increment on every fallback trigger

## Epic 3: Raw Vault & Durable Preservation

Build the custom WAL with `O_DIRECT/mmap` for high-speed "Shock Absorber" capture.

### Story 3.1: Initialize Custom WAL (`internal/vault`)
As a system developer, I want a custom Write-Ahead Log implementation, so that I can flush raw bytes to disk while bypassing the Go heap and OS page cache.

**Acceptance Criteria:**
- **Given** the raw ingestion path requires persistent storage
- **When** the WAL is initialized
- **Then** it must use `O_DIRECT` (on Linux) or `mmap` for high-performance I/O
- **And** it must support sequential segment rotation based on size or time

### Story 3.2: Implement LZ4 Fast Compression
As a platform engineer, I want the Raw Vault to use LZ4 compression, so that I can maximize disk throughput without saturating the CPU.

**Acceptance Criteria:**
- **Given** raw bytes are being written to the vault
- **When** compression is enabled
- **Then** segments must be compressed using `pierrec/lz4`
- **And** write throughput must remain within 10% of raw disk I/O speed

### Story 3.3: Raw Vault Replay Mechanism
As an SRE, I want to "replay" the Raw Vault once a traffic spike subsides, so that I can recover all logs for deferred parsing.

**Acceptance Criteria:**
- **Given** logs are stored in the Raw Vault
- **When** a replay command is issued
- **Then** the `internal/vault` iterator must stream bytes back to the Somali Controller
- **And** the replay rate must be throttleable to prevent new OOM events

### Story 3.4: Data Integrity & Checksumming
As a compliance officer, I want cryptographic checksums for every WAL segment, so that I can guarantee the bit-identical integrity of replayed logs.

**Acceptance Criteria:**
- **Given** data is being written to a WAL segment
- **When** the segment is finalized
- **Then** a CRC32 checksum must be computed and stored
- **And** any checksum mismatch during replay must trigger an alert and quarantine the segment

## Epic 4: Stochastic Monitor & Stability Optimizer

Implement "Lazy Atomic" monitoring and memory budgeting to prevent host OOMs.

### Story 4.1: Lazy Atomic Environment Sensing
As an engine, I want to monitor host memory and CPU pressure only every 1024 operations, so that I can maintain "Stochastic Awareness" without creating cache contention on high-core systems.

**Acceptance Criteria:**
- **Given** the stochastic monitor is active
- **When** the ingestion hot path executes
- **Then** a host pressure check (RAM/CPU) must only occur every N operations (default 1024)
- **And** an atomic counter must be used to track operation count without locking

### Story 4.2: Component-Level Memory Budgeting
As a system administrator, I want to define hard memory limits for the Ingester and Raw Vault, so that GopherShip never causes a host-level OOM crash.

**Acceptance Criteria:**
- **Given** a global memory budget of X MB
- **When** individual component usage approaches its allocated share
- **Then** the somatic zone must transition to "Red" 
- **And** the engine must force all new logs to the Raw Vault WAL to preserve RAM

### Story 4.3: Stochastic Component Throttling
As an optimizer, I want to slow down background sync and parsing tasks when the Stochastic Monitor detects high host load, so that the Ingestion reflex always has priority access to CPU cycles.

**Acceptance Criteria:**
- **Given** the Stochastic Monitor detects high CPU pressure or a "Yellow/Red" zone
- **When** background workers (Vault Sync/Deferred Parsing) check for work
- **Then** they must increase their sleep interval or limit their batch size
- **And** priority must be mathematically weighted toward the ingestion reflex

## Epic 5: Control Plane & gs-ctl Management

Develop the secure management interface and the `gs-ctl` console with real-time observability and emergency overrides.

### Story 5.1: Initialize Secure mTLS Control Plane (`internal/control`)
As a security officer, I want a gRPC control plane secured by Mutual TLS (mTLS), so that only authorized administrators can execute sensitive management commands.

**Acceptance Criteria:**
- **Given** the GopherShip engine is starting
- **When** the control server is initialized
- **Then** it must require valid client certificates for all gRPC methods
- **And** the `gs-ctl` client must provide a valid certificate/key pair to connect

### Story 5.2: `gs-ctl` Core & Automation Support
As an SRE, I want `gs-ctl` commands to support multiple output formats, so that I can easily integrate GopherShip into automated shell scripts and recovery workflows.

**Acceptance Criteria:**
- **Given** I am running `gs-ctl status`, `replay`, or `drain`
- **When** I provide the `--output` flag (json or yaml)
- **Then** the CLI must return the data in the requested structured format
- **And** default output must remain a human-readable table

### Story 5.3: Real-time Somatic Dashboard (`gs-ctl top`)
As a lead SRE, I want a live dashboard of engine health directly in my terminal, so that I can observe "Hardware Honest" metrics during a traffic surge.

**Acceptance Criteria:**
- **Given** the GopherShip engine is active
- **When** I run `gs-ctl top`
- **Then** the terminal must display a live-updating view of Goroutine counts, Memory Pressure, and the current 'Somatic Zone' (Green/Yellow/Red)
- **And** the refresh rate must be configurable (default 1s)

### Story 5.4: Emergency Somatic Override
As an operator, I want to manually force the system into a specific somatic state, so that I can proactively protect the host before a known traffic spike hits the sensors.

**Acceptance Criteria:**
- **Given** an impending traffic surge
- **When** I execute `gs-ctl override --zone red`
- **Then** the engine must instantly pivot to the Raw Vault WAL regardless of current sensor data
- **And** the override must be auditable in the GopherShip logs

## Epic 6: OTel Compliance & Production Hardening

Finalize OTel data mapping, TLS 1.3 encryption for ingestion, and container packaging.

### Story 6.1: Full OTel Log Model Mapping (`pkg/otel`)
As a data engineer, I want our ingested logs to strictly follow the OpenTelemetry Log Data Model, so that GopherShip is compatible with the wider observability ecosystem.

**Acceptance Criteria:**
- **Given** the engine is processing logs
- **When** the internal data is converted for export
- **Then** it must strictly follow the `ResourceLogs` and `ScopeLogs` structures
- **And** all standard attributes (timestamp, severity, body) must be correctly mapped

### Story 6.2: Production TLS 1.3 Ingestion
As a security engineer, I want all network-based ingestion endpoints to enforce TLS 1.3, so that data in transit is protected by modern cipher suites.

**Acceptance Criteria:**
- **Given** a log source is connecting over TLS
- **When** the handshake occurs
- **Then** GopherShip must reject any protocol version lower than TLS 1.3
- **And** it must support modern cipher suites as recommended by OTel best practices

### Story 6.3: Optimized Container Packaging (K8s Sidecar)
As a DevOps engineer, I want a multi-stage Docker build for a minimal static binary, so that I can deploy GopherShip as a lightweight sidecar or DaemonSet in Kubernetes.

**Acceptance Criteria:**
- **Given** the project is ready for deployment
- **When** the Docker image is built
- **Then** the final image must contain only the static binary
- **And** the image size must be less than 20MB
- **And** it must include health check probes that query the internal somatic state
