# Story 5.2: gs-ctl Core & Automation Support

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an SRE,
I want `gs-ctl` commands to support multiple output formats and automation-friendly exit codes,
so that I can easily integrate GopherShip into automated shell scripts and recovery workflows.

## Acceptance Criteria

1. **[AC1]** `gs-ctl status` MUST support the `--output` flag with values: `table` (default), `json`, and `yaml`.
2. **[AC2]** The `json` output MUST be valid JSON, suitable for piping into `jq`.
3. **[AC3]** The `yaml` output MUST be valid YAML, following the project's configuration style.
4. **[AC4]** `gs-ctl status` MUST return numeric exit codes corresponding to the engine's Somatic Zone:
   - Exit 0: `ZONE_GREEN`
   - Exit 1: `ZONE_YELLOW` (Warning)
   - Exit 2: `ZONE_RED` (Critical)
5. **[AC5]** Automation support MUST be verified by a shell script that parses JSON output and checks exit codes against simulated zone transitions.

## Tasks / Subtasks

- [x] Task 1: CLI Flag & Argument Refactoring (AC: #1)
  - [x] Implement a robust flag parser for global options like `--output`.
  - [x] Ensure `--socket` and `--addr` flags are still respected.
- [x] Task 2: Multi-Format Formatter Implementation (AC: #1, #2, #3)
  - [x] Implement a `Formatter` interface for `table`, `json`, and `yaml`.
  - [x] Integrate `encoding/json` and `gopkg.in/yaml.v3` for structured output.
  - [x] Maintain the existing `tabwriter` logic for the `table` format.
- [x] Task 3: Automation Exit Code Logic (AC: #4)
  - [x] Map `protocol.SomaticZone` to OS exit codes in `main.go`.
  - [x] Ensure errors (connectivity, auth) still return non-zero (e.g., 128+).
- [x] Task 4: [VERIFICATION] Automation Script (AC: #5)
  - [x] Create `scripts/test-automation.sh` to verify `gs-ctl` output and exit codes.
  - [x] Verify `jq` compatibility for JSON output.

## Dev Notes

- **Architecture Compliance**: CLI management MUST remain secondary to ingestion; the formatter should be lightweight and not block the main engine's telemetry loop if called frequently. [Source: architecture.md#Infrastructure-&amp;-Deployment]
- **Library Requirements**: Use `gopkg.in/yaml.v3` for YAML output to ensure consistency with project config.
- **Exit Code Convention**: Follow standard Unix conventions; specific somatic exit codes help monitoring agents (like Nagios/Zabbix) respond without heavy parsing.

### Project Structure Notes

- Alignment with `cmd/gs-ctl` and `pkg/protocol`.
- The formatter logic should likely reside in `internal/control/formatter.go` if shared, or locally in `cmd/gs-ctl` if CLI-only.

### References

- [architecture.md](../../_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns) - Management API patterns.
- [epics.md](../../_bmad-output/planning-artifacts/epics.md#Story-5.2) - Core automation requirement.
- [cmd/gs-ctl/main.go](../../cmd/gs-ctl/main.go) - Current CLI implementation.

## Dev Agent Record

### Agent Model Used

Antigravity

### Debug Log References

### Completion Notes List

### File List

- [cmd/gs-ctl/main.go](../../cmd/gs-ctl/main.go)
- [cmd/gs-ctl/formatter.go](../../cmd/gs-ctl/formatter.go)
- [cmd/gs-ctl/formatter_test.go](../../cmd/gs-ctl/formatter_test.go)
- [scripts/test-automation.sh](../../scripts/test-automation.sh)
- [go.mod](../../go.mod) (updated for yaml.v3)

## Adversarial Review Summary

- **Status**: ✅ PASSED after follow-up fixes.
- **Key Hardening**:
  - Deterministic table output (sorted keys).
  - Precise exit codes (129: Connection, 130: Auth).
  - Automation script simulates zone transitions (AC5).
  - Zero-dependency verification (removed jq requirement).
