# 🎓 GopherShip Tutorial: Zero-Allocation Ingestion in Action

This tutorial walks you through setting up GopherShip locally, sending your first logs, and observing its "Somatic Reflexes" as it handles backpressure.

---

## 🛠️ Step 1: Build the Binaries

First, ensure you have Go 1.22+ installed.

```bash
# Clone the repository
git clone https://github.com/sungp/gophership.git
cd gophership

# Build the main engine
go build -o bin/gophership ./cmd/gophership

# Build the control CLI
go build -o bin/gs-ctl ./cmd/gs-ctl
```

---

## 🔑 Step 2: Generate Test Certificates (Experimental/Local)

GopherShip enforces **TLS 1.3** for ingestion. For local testing, you can use the provided script or generate self-signed certs:

```bash
# Using the provided script (if available)
./scripts/generate-certs.sh
```

*If the script isn't available, ensure you have `ca.crt`, `server.crt`, and `server.key` in your current directory.*

---

## 🚀 Step 3: Launch the Engine

Start GopherShip with the default configuration. We'll run it in a way that allows us to see the structured logs.

```bash
# Run the engine
./bin/gophership
```

You should see:
`{"level":"info","msg":"Starting GopherShip Engine (Tier 1 Foundation)"...}`

---

## 📡 Step 4: Send Your First Logs

GopherShip implements the **OpenTelemetry (OTLP) gRPC** protocol. You can use any OTel-compatible producer or a tool like `grpcurl` to send logs.

### Using `grpcurl` (Simulated Log)
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "scopeLogs": [{
      "logRecords": [{
        "body": { "stringValue": "Biological reflex test successful" },
        "severityText": "INFO"
      }]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```

---

## 🔍 Step 5: Monitor Engine Health

Open a new terminal and use `gs-ctl` to check the engine's internal state.

### Real-time Dashboard
```bash
./bin/gs-ctl top
```

### Static Status Check
```bash
./bin/gs-ctl status
```

---

## 🔴 Step 6: Trigger a Somatic Reflex (Simulation)

To see GopherShip pivot to the **Raw Vault (Red Zone)**, we can simulate memory pressure or use a manual override.

### Manual Pivot to Red Zone
```bash
./bin/gs-ctl override --zone red
```

Now, check the engine logs. You'll notice it stops parsing and starts flushing raw bytes directly to disk at wire speed.

---

## 🤝 Next Steps
- Explore `deployments/` for Kubernetes Sidecar patterns.
- Read the [Architecture Guide](architecture.md) to understand the "Hardware-Honest" design principles.
