# Internal Core

This directory contains the "biological" core logic of GopherShip.

## Encapsulation Policy

Logic within `internal/` is strictly private to this repository and should NOT be exported or used by external packages. This ensures the integrity of the Hardware Honest reflexes and zero-allocation constraints.

## Design Mandates

- **Hardware Honest**: Code must respect CPU/Cache/Disk physical realities.
- **Biological Resilience**: Treat ingestion as a mandatory reflex, processing as opportunistic debt.
- **Zero Allocation**: No heap allocations in the ingestion hot-path (`ingester` -> `buffer` -> `vault`).

## Components

- **ingester**: OTLP/gRPC ingestion skeleton (Status: Conceptual Skeleton).
- **vault**: Custom WAL implementation with mmap/O_DIRECT.
- **somatic**: The pivot controller for pressure-based transitions.
- **stochastic**: Stochastic Awareness pattern (Lazy atomic state monitoring).
- **control**: Secure mTLS management plane.
- **buffer**: Zero-allocation binary buffer pools (`sync.Pool`).
