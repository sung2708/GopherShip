---
stepsCompleted: [step-01-init, step-02-discovery, step-02b-vision, step-02c-executive-summary, step-03-success, step-04-journeys, step-05-domain, step-06-innovation, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish]
inputDocuments: 
  - product-brief-GopherShip-2026-02-24.md
  - market-high-performance-log-middleware-2026-02-24.md
  - brainstorming-session-2026-02-24.md
documentCounts:
  briefCount: 1
  researchCount: 1
  brainstormingCount: 1
  projectDocsCount: 1
classification:
  projectType: developer_tool
  domain: general
  complexity: high
  projectContext: brownfield
workflowType: 'prd'
---

# Product Requirements Document - GopherShip

**Author:** sungp
**Date:** 2026-02-24

## Executive Summary

GopherShip is a high-performance log ingestion engine designed as a **"Biological Resilient Engine"** for environments where standard shippers (Vector, Fluent Bit) fail. It addresses the flaw of **"Hardware Blindness"**—where rigid backpressure leads to host-level OOM crashes—by treating ingestion as a mandatory reflex and processing as opportunistic debt. Through its **Hybrid Somatic Model**, GopherShip ensures zero data loss and host stability during extreme traffic bursts.

## Project Classification

- **Project Type**: Developer Tool / Infrastructure
- **Domain**: Systems Middleware & Observability
- **Complexity**: High (Zero-loss, 1M+ LPS targets)
- **Project Context**: Brownfield (Core foundation and Ingester skeleton initialized)

## Success Criteria

### User Success
- **Zero-Loss Relief**: 100% data integrity during "Red Zone" events via Raw Vault fallback.
- **Stability Guarantee**: Zero OOM crashes of the GopherShip process under full buffer saturation.
- **Micro-Visibility**: Real-time telemetry on somatic fallback triggers and buffer state.

### Business Success
- **Efficiency**: < 5% host CPU/RAM overhead for ingestion, minimizing the "Log Tax."
- **Market Fit**: Establishment as the premier Tier-0 "Shock Absorber" for high-traffic SaaS.
- **Predictable Costs**: Lowered observability TCO through deferred/selective parsing.

### Technical Success
- **Zero-Allocation Hot Path**: Network-to-Buffer ingestion with zero heap allocations.
- **Linear Scaling**: Performance scales linearly up to 128+ cores via cache-local state.
- **Physical Sensing**: Non-blocking backpressure reaction time < 1ms.

## Product Scope & Phased Development

### MVP Strategy (Phase 1: The "Shock Absorber")
**Approach:** Problem-Solving MVP focusing on preventing host death during 10x traffic bursts.
**Must-Have Capabilities:**
- **Somatic Core**: Ingester with "Select-Default" reflex and non-blocking paths.
- **Stochastic Awareness**: Lazy global state monitoring to eliminate cache contention.
- **Raw Vault**: Local Write-Ahead Log (WAL) flushing for raw data preservation.
- **Stability Core**: Integrated circuit breakers and hardware-honest pressure sensing.

### Phase 2: Growth (The "Observability Hub")
- **Deep Sync**: Automated background streaming from Raw Vault to S3/Cloud storage.
- **OTel Native**: First-class ingestion and export for OpenTelemetry signals.
- **Visual Dashboard**: Lightweight UI for observing internal somatic health (Green/Yellow/Red).

### Phase 3: Vision (The "Kernel Native")
- **Zero-Copy eBPF**: Kernel-level ingestion for maximum possible throughput.
- **Predictive Reflex**: ML-assisted anticipation of traffic bursts based on historical signals.

## Innovation & Novel Patterns

### The Somatic Reflex
Traditional shippers use logical queues that block; GopherShip uses "physical" reflexes (Go `select-default` primitives) to pivot in microseconds when logical buffers hit physical limits.

### Stochastic Awareness
Instead of global locks (the "Atomic Wall"), GopherShip uses "good enough" periodic checks (e.g., every 1024 ops) to keep memory channels hot and eliminate cache-line bouncing on high-core machines.

### Temporal Decoupling
The **Raw Vault** decouples ingestion speed from parsing weight. During a crisis, GopherShip bypasses CPU-heavy logic to flush raw bytes directly to storage, paying the "Parsing Debt" only when slack capacity returns.

## User Journeys

### 1. The "Black Swan" Crisis (Alex, Lead SRE)
- **Scenario**: 2 AM traffic spike (20x). pod-level shippers are hitting OOM limits and crashing pods.
- **Action**: GopherShip senses buffer pressure and pivots instantly to the **Raw Vault**, flushing unparsed bytes to the local WAL at hardware speed.
- **Outcome**: The primary application remains stable. All logs are preserved for post-mortem analysis.

### 2. The "Next-Day" Recovery (Jordan, Platform Engineer)
- **Scenario**: 9 AM normalization. Jordan needs to reconcile crisis-period data.
- **Action**: Jordan uses `gs-ctl` to "replay" the Raw Vault.
- **Outcome**: GopherShip parses and enriches the raw bytes using now-available CPU cycles, streaming them to the final sink.

### 3. The "Hardware Optimizer" (Sam, FinOps)
- **Scenario**: Sam investigates high cloud bills from log ingestion CPU usage.
- **Action**: Deployment of GopherShip with **Stochastic Awareness**.
- **Outcome**: CPU efficiency improves by ~70%, cutting the "Log Tax" to single digits.

## Domain-Specific Requirements

### Compliance & Regulatory
- **OTel Compliance**: Must adhere to the OpenTelemetry Log Data Model for ecosystem interoperability.
- **Auditability**: Sequential indexing and cryptographic checksums for all Raw Vault segments.

### Technical Constraints
- **NUMA Awareness**: Priority on cache-local state over global synchronization for 64+ core machines.
- **Zero-Allocation**: Zero heap allocations for the ingestion-to-buffer reflex path.
- **S3 Consistency**: Error handling for out-of-order delivery during Raw Vault cloud replay.

### Risk Mitigations
- **Resource Starvation**: Hard IO limits on Raw Vault to protect host system availability.
- **Stochastic Drift**: Regular (but infrequent) global sync to prevent workers from stalling on stale state.

## Developer Tool Specific Requirements

### Distribution & Integration
- **Static Binary**: Single dependency-free Go binary (Linux/Windows).
- **Containerization**: Optimized Docker images for K8s Sidecar/DaemonSet deployment.
- **Language Native**: Direct Go 1.22+ package support for embedded use cases.

### API & Interface
- **Ingestion Support**: Syslog (UDP/TCP), OTel Log Protocol, and Newline-delimited JSON.
- **Control Plane**: `gs-ctl` CLI for real-time monitoring and manual Vault replay.
- **Observability**: Prometheus-compatible metrics reporting internal pressure zones.

### CLI Command Structure & Output
- **Primary Commands**: `status` (internal health), `replay` (vault processing), `drain` (graceful shutdown).
- **Execution**: `gs-ctl [command] [options]`
- **Output Formats**: Human-readable tables (default), JSON, and YAML for automation.

### Configuration Schema
- **Structure**: YAML-based configuration for engine parameters and vault paths.
- **Somatic Sensitivity**: Tuning parameters for the non-blocking reflex thresholds.
- **Vault Limits**: Configuration for maximum local storage and rotation policies.

## Functional Requirements

### Somatic Ingestion
- FR1: **Ingestion Pipeline** can ingest logs via non-blocking reflexes to prevent upstream backpressure.
- FR2: **Somatic Engine** can detect buffer saturation in < 1ms to trigger defensive pivots.
- FR3: **Somatic Engine** can switch between full enrichment and Raw Vault capture instantly based on pressure zones.

### Raw Data Preservation
- FR4: **Raw Vault** can flush raw, unparsed bytes to a local WAL when in the "Red Zone."
- FR5: **Raw Vault** can replay stored segments for deferred parsing once pressure subsides.
- FR6: **Raw Vault** can maintain cryptographic checksums for all raw segments to ensure data integrity.

### Hardware Optimizer
- FR7: **Engine** can monitor global environment health via lazy status updates.
- FR8: **Somatic Engine** can throttle internal components based on stochastic awareness to eliminate cache contention.
- FR9: **Engine** can manage internal memory budgets to prevent host-level OOM events.

## Non-Functional Requirements

### Performance & Scalability
- **NFR.P1 - Zero-Allocation**: Zero heap allocations in the "Ingest Reflex" path.
- **NFR.P2 - Reflex Latency**: < 500μs (P99) from wire to somatic buffer.
- **NFR.S1 - Linear Scaling**: Linear throughput scaling on machines up to 128 cores.
- **NFR.S2 - High-Density**: Support for 1M+ LPS using < 2 vCPU cores.

### Reliability & Security
- **NFR.R1 - Zero-Crash**: Process survival with 100% full buffers without OOM or deadlocks.
- **NFR.R2 - Data Integrity**: 100% bit-identical data preservation upon Raw Vault replay.
- **NFR.Sec1 - Encryption**: TLS 1.3 support for all ingestion endpoints.
- **NFR.Sec2 - Management Access**: Unix socket restricted permissions for the control interface.
