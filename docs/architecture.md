# 🏗️ GopherShip Architecture Guide

GopherShip is designed as a **Hardware-Honest** log ingestion engine. Unlike traditional "Log Shippers" that prioritize logical delivery at all costs, GopherShip treats host stability as the primary mandate.

---

## 🏎️ Core Philosophy: The Biological Reflex

GopherShip operates like a biological nervous system. In normal conditions, it performs complex cognitive tasks (parsing, enrichment). When faced with a "Black Swan" event (20x traffic spikes), it reverts to a mandatory **Somatic Reflex**, bypassing logic to protect the host from OOM (Out Of Memory) conditions.

### The Two Mandates
1.  **Mandatory Ingestion**: Never refuse a connection from a local producer.
2.  **Hardware Integrity**: Never allow heap growth to trigger the OS OOM-Killer.

---

## 🌈 The Somatic Zone System

The engine pivots between three zones based on real-time telemetry from the `Stochastic Monitor`.

| Zone | Status | Behavior | Path |
| :--- | :--- | :--- | :--- |
| 🟢 **Green** | Healthy | Full enrichment, real-time OTLP export. | **Hot Path** |
| 🟡 **Yellow** | Pressure | Throttle background tasks, prioritize ingestion. | **Throttled Path** |
| 🔴 **Red** | **Reflex** | Logic bypassed. Raw bytes flushed to **Raw Vault** at wire speed. | **Emergency Path** |

---

## 🧩 Component Breakdown

### 1. Ingester (`internal/ingester`)
The "Mouth" of the system. It handles OTLP gRPC ingestion.
- **Zero-Allocation**: Uses a global `sync.Pool` for internal buffers.
- **Backpressure Aware**: Communicates with the Somatic Pivot to decide whether to enrich or vault.

### 2. Stochastic Monitor (`internal/stochastic`)
The "Sensory Cortex". It samples host telemetry (RAM, CPU, Pressure Stall Information).
- **Sampling**: Checks host state every $N$ operations to avoid overhead.
- **Hysteresis**: Ensures smooth transitions between zones to prevent "flapping".

### 3. Raw Vault (`internal/vault`)
The "Short-term Memory". A high-speed persistence layer used during Red Zone events.
- **O_DIRECT**: Bypasses the OS page cache for deterministic disk I/O.
- **Fast Compression**: Uses LZ4 for high-throughput, low-CPU compression.
- **WAL (Write Ahead Log)**: Ensures data integrity during reflex events.

### 4. Control Plane (`internal/control`)
The "Autonomic Nervous System". Provides a secure portal for management.
- **mTLS Enforced**: Secure communication for CLI (`gs-ctl`) and remote dashboards.
- **Real-time Monitoring**: Streams somatic status via gRPC.

---

## 🌊 Data Flow

### Normal Flow (Green)
`App` ➔ `OTLP gRPC` ➔ `Enrichment` ➔ `Exporter` ➔ `Collector (e.g. Honeycomb/Datadog)`

### Reflex Flow (Red)
`App` ➔ `OTLP gRPC` ➔ `Zero-Alloc Buffer` ➔ `Raw Vault (WAL)` ➔ `Disk`

*When the pressure subsides (Back to Green), the `Replayer` reads from the Vault and feeds the data back into the enrichment pipeline.*

---

## 🛡️ Security Posture
- **TLS 1.3**: Mandatory for all ingestion and control traffic.
- **mTLS**: Client identity is verified for every connection.
- **Distroless**: Production images contain only the binary and CA certs, minimizing attack surface.
