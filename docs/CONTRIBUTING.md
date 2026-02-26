# 🛠️ GopherShip Hacking Guide (CONTRIBUTING)

Welcome to the GopherShip engineering community! GopherShip is built to be a high-density, **Zero-Allocation** log engine. To maintain this performance profile, we follow strict coding conventions.

---

## 🏎️ Zero-Allocation Mandate

Every contribution to the GopherShip hot path (ingestion, buffering, vaulting) **must** aim for zero heap allocations.

### 1. Unified Buffering
Never allocate raw slices or strings in the ingestion path. Use the `internal/buffer` package:

```go
// BAD
data := make([]byte, 1024)

// GOOD
buf := buffer.MustAcquire(1024)
defer buffer.Release(buf)
```

### 2. Avoid `fmt.Sprintf` and Reflection
`fmt.Sprintf` almost always triggers allocations. Use `strconv.AppendInt` or direct byte slice manipulation.

### 3. Prefer `[]byte` Over `string`
Strings are immutable and often cause copies. Keep data in `[]byte` as long as possible.

---

## 🧪 Performance Verification

All hot-path changes **must** include allocation tests.

```go
func TestYourFeature_ZeroAllocation(t *testing.T) {
    allocs := testing.AllocsPerRun(100, func() {
        // Execute your performance-critical logic here
    })
    if allocs > 0 {
        t.Errorf("Expected 0 allocations, got %v", allocs)
    }
}
```

---

## 📝 Pull Request Guidelines

1.  **Architecture First**: If changing component boundaries, update the [Architecture Guide](architecture.md).
2.  **Benchmarked**: All performance-critical PRs must include benchmark comparisons.
3.  **No Leaks**: Ensure all `buffer.MustAcquire` calls have a corresponding `buffer.Release`.
4.  **Hardware-Honest**: Test your changes on resource-constrained environments (e.g., restricted RAM).

---

## 🏗️ Local Development

Follow the [Step-by-Step Tutorial](tutorial.md) to get your environment ready.

*GopherShip depends on you to keep the hot path "Biological"—fast, efficient, and resilient.*
