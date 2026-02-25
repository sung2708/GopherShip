# Story 2.3: Somatic Ingestion Metric Reporting

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an SRE,
I want to see real-time metrics on "Pressure Zones",
so that I can visualize when the engine is falling back to the Raw Vault.

## Acceptance Criteria

1. **[AC1]** The Prometheus gauge `gophership_ingester_zone` MUST be updated whenever a somatic zone transition occurs (Green/Yellow/Red).
2. **[AC2]** The Prometheus counter `gophership_somatic_pivots_total` MUST increment on every somatic fallback trigger (reflex path).
3. **[AC3]** Metrics MUST follow the naming pattern: `gophership_{component}_{metric}_{unit}`.
4. **[AC4]** The metrics collection MUST be zero-allocation in the hot path.
5. **[AC5]** A Prometheus metrics endpoint MUST be exposed on port `9091` (as specified in architecture).

## Tasks / Subtasks

- [x] Task 1: Initialize Global Metrics Registry (AC: #3, #5)
  - [x] Create `internal/stochastic/metrics.go` to hold the Prometheus registry and metric definitions.
  - [x] Implement `StartMetricsServer(addr string)` to expose the `/metrics` endpoint.
- [x] Task 2: Implement Somatic Zone Gauge (AC: #1, #4)
  - [x] Integrate `gophership_ingester_zone` gauge into `stochastic.MustSetAmbientStatus`.
  - [x] Ensure gauge updates only happen on actual state changes (lazy update).
- [x] Task 3: Implement Pivot Counter (AC: #2, #4)
  - [x] Integrate `gophership_somatic_pivots_total` into `Ingester.somaticFallback`.
  - [x] Verify that counter increments are performed efficiently using the Prometheus Go client.
- [x] Task 4: [VERIFICATION] Stress test metrics performance.
  - [x] Run `BenchmarkPivotLatency` to ensure metrics don't introduce regressions.

## Dev Notes

- **Metric Naming**:
  - `gophership_ingester_zone`: Gauge, tracking 0 (Green), 1 (Yellow), 2 (Red).
  - `gophership_somatic_pivots_total`: Counter, tracking total reflex triggers.
- **Port Assignment**: Port `9091` for Prometheus (Architecture Decision).
- **Zero Allocation**: Avoid string formatting in label values; use static labels where possible.
- **Lazy Update**: `MustSetAmbientStatus` already performs a delta-check; the gauge update should be inside that conditional block.

### Project Structure Notes

- **Package**: `github.com/sungp/gophership/internal/stochastic` for Registry/Definitions.
- **Integration**: `internal/ingester` for counter increments; `internal/somatic` for status triggers.

### References

- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-2.3) - Functional Requirement FR2/FR3
- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#Observability-Patterns) - Prometheus Metric Naming & Port 9091
- [stochastic/state.go](../../internal/stochastic/state.go) - Current ambient status implementation.
- [ingester.go](../../internal/ingester/ingester.go) - Fallback path for counter trigger.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

- `TestMetricsServer`: PASS (0.10s)
- `TestIngesterZoneGauge`: PASS (0.10s)
- `TestIngester_SomaticPivotsCounter`: PASS (0.00s)
- `TestIngester_P99PivotLatency`: PASS (0.01s, P99=0s)
- `TestIngester_ZeroAllocationPivot`: PASS (0.00s, 0 allocs)

### Completion Notes List

- Initialized Prometheus custom registry in `internal/stochastic/metrics.go`.
- Exposed `/metrics` endpoint on port `9091`.
- Integrated `gophership_ingester_zone` gauge into `MustSetAmbientStatus` and pressure adjustment functions.
- Integrated `gophership_somatic_pivots_total` counter into `Ingester.somaticFallback`.
- Verified zero-allocation requirement for the hot path.

### File List

- [NEW] [internal/stochastic/metrics.go](../../internal/stochastic/metrics.go)
- [NEW] [internal/stochastic/metrics_test.go](../../internal/stochastic/metrics_test.go)
- [MODIFY] [internal/stochastic/state.go](../../internal/stochastic/state.go)
- [MODIFY] [internal/ingester/ingester.go](../../internal/ingester/ingester.go)
- [MODIFY] [internal/ingester/ingester_test.go](../../internal/ingester/ingester_test.go)
- [MODIFY] [cmd/gophership/main.go](../../cmd/gophership/main.go)

## Senior Developer Review (AI)

**Review Date**: 2026-02-25
**Reviewer**: Antigravity (Adversarial)
**Outcome**: Approved (After hardening fixes)

### Action Items
- [x] [AI-Review][High] Start metrics server in `main.go` (AC5 compliance)
- [x] [AI-Review][High] Fix race condition between atomic state and Prometheus gauge
- [x] [AI-Review][Med] Implement graceful shutdown for Prometheus server
- [x] [AI-Review][Med] Rename metric to `gophership_ingester_zone_index` for AC3 pattern
- [x] [AI-Review][Low] Fix test port leakage

### Review Notes
The initial implementation provided the core logic but failed to wire the server into the production entry point and had a critical race condition that could lead to dashboard/engine desync. Hardening fixes have ensured atomicity and graceful lifecycle management.

## Change Log

- 2026-02-25: Initial implementation of somatic metrics.
- 2026-02-25: Addressed code review findings - 5 items resolved.
