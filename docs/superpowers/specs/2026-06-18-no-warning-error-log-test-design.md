# No Warning/Error Log Test for nrdot-collector

**Issue:** [kbauer/personal-work#42](https://source.datanerd.us/kbauer/personal-work/issues/42) — "no warn/error-log test nrdot-collector"
**Date:** 2026-06-18
**Branch:** main worktree at `.claude/worktrees/no-warnings-test-enforcement`
**Scope:** `distributions/nrdot-collector` only (host scenario + internal/receiver scenario in `spec-local.yaml`)

## Problem

`spec-local.yaml` for `nrdot-collector` validates that telemetry data lands in New Relic, but does not assert that the collector itself is running cleanly. A deprecation warning, a misconfigured component, or a regression that surfaces only as a `WARN`/`ERROR` log line will pass tests today.

We want a test that fails when the collector emits unexpected `WARN` or `ERROR` log lines. Some warnings are expected:

- The **internal** scenario (`spec-local.yaml` second scenario) deliberately misconfigures exporters with unreachable endpoints to drive `enqueue_failure` / `send_failure` metrics. These produce error logs by design.
- The **host** scenario may produce warnings from the `kind` test environment (e.g., `file_log` paths that don't exist in containers, `resource_detection/cloud` failing without cloud metadata).

Both expected sources need an allow-list. New deprecation warnings or other unexpected lines must fail the test.

## Non-Goals

- Equivalent assertions for `nrdot-collector-experimental` (handled separately if needed).
- Changes to nightly E2E tests (`spec-nightly-*.yaml`) — out of scope; align later if useful.
- A general-purpose log-tail framework. The assertion is a single NRQL query per scenario; the spec format already supports it.
- Changes to `examples/internal-telemetry-config.yaml`. That file is a customer-facing template with a specific documented purpose. We do not modify it.

## Design

### Architecture

One new test-only collector config file plus two NRQL assertions (one per scenario) in the existing `spec-local.yaml`.

**For the host scenario:** the default config does not export collector-internal logs to NR, so we add a small dedicated config snippet that does. Sampling is disabled so low-frequency warnings (e.g., once-at-startup deprecation messages) are never dropped.

**For the internal scenario:** internal telemetry is already exported via `examples/internal-telemetry-config.yaml`, but with sampling enabled (`initial: 10, thereafter: 100`). We override sampling off via an inline `--config=yaml:` flag. No file change.

Both scenarios get a single new NRQL query that counts collector log records with severity `WARN` or `ERROR`, scoped by a static `service.name` and the existing `testKey` resource attribute, excluding allow-listed message prefixes. Expected count is `0`.

### Files Touched

| Path | Change |
|------|--------|
| `distributions/nrdot-collector/test/host-internal-logs-config.yaml` | **New.** Test-only OTel config snippet that exports `service.telemetry.logs` to NR via OTLP and tags resource with `service.name=nrdot-collector-e2e-host`, `testKey=${env:SCENARIO_TAG}`. Sampling off. |
| `distributions/nrdot-collector/test/host-collector-values.yaml` | Add `extraVolumes` + `extraVolumeMounts` for a configmap mounted at `/etc/nrdot-collector/extra/`. Append `--config=/etc/nrdot-collector/extra/host-internal-logs-config.yaml` to `command.extraArgs`. |
| `distributions/nrdot-collector/test/internal-collector-values.yaml` | Append one `--config=yaml:` flag to `command.extraArgs`: `service::telemetry::logs::sampling::enabled: false`. |
| `distributions/nrdot-collector/test/spec-local.yaml` | Host `before`: create configmap from the new file. Host `tests.nrqls`: append `*host_log_assertions` anchor. Internal `tests.nrqls`: append `*internal_log_assertions` anchor. Both anchors defined at top of file alongside existing `&host_metrics_nrqls` / `&internal_telemetry_nrqls`. |

No other files change.

### New Config File Contents

`distributions/nrdot-collector/test/host-internal-logs-config.yaml`:

```yaml
# Test-only: exports the host scenario's internal collector logs to NR so the
# no-warn/error-log assertion in spec-local.yaml can query them. Sampling is
# OFF so low-frequency warnings (e.g., deprecations at startup) are not dropped.
# Not for production use; reuses the same OTLP endpoint and license-key env
# vars that the data-plane otlphttp exporter already consumes.
service:
  telemetry:
    logs:
      level: INFO
      sampling:
        enabled: false
      processors:
        - batch:
            exporter:
              otlp:
                protocol: http/protobuf
                endpoint: "${env:OTEL_EXPORTER_OTLP_ENDPOINT:-https://otlp.nr-data.net}"
                headers:
                  - name: api-key
                    value: "${env:NEW_RELIC_LICENSE_KEY}"
    resource:
      service.name: nrdot-collector-e2e-host
      testKey: "${env:SCENARIO_TAG}"
```

The `service.name` value matches the existing convention in `spec-local.yaml` (host scenario's secret already passes `serviceName=nrdot-collector-e2e-host`); we hardcode it here so the file is self-contained.

### NRQL Assertion Form

Both scenarios use the same shape, differing only in `service.name` and allow-list. Initial form (allow-list empty):

```sql
FROM Log SELECT count(*) AS unexpected_warnings
WHERE service.name = '<scenario service name>'
  AND testKey = '${SCENARIO_TAG}'
  AND severity_text IN ('WARN','ERROR')
```

with `upperBoundedValue: 0`.

The actual attribute name (`severity_text` vs `level` vs `severity.text`) and value casing (`WARN` vs `Warn` vs `warn`) **must be confirmed by a discovery NRQL on the first run** — see iteration plan. The form above is the expected one but is provisional until verified.

As warnings are triaged, the allow-list grows as additional clauses inside the same query:

```sql
... AND NOT (
  message LIKE 'expected pattern 1%'
  OR message LIKE 'expected pattern 2%'
)
```

Each `LIKE` pattern is anchored at the start (suffix `%`) and gets a one-line YAML comment naming the source (component, version, why expected).

### Spec Wiring

`spec-local.yaml` host scenario `before` adds before the helm install:

```yaml
- kubectl create configmap host-extra-config --namespace=nr-${SCENARIO_TAG} \
    --from-file=host-internal-logs-config.yaml=./host-internal-logs-config.yaml
```

`host-collector-values.yaml` adds:

```yaml
extraVolumes:
  - name: extra-config
    configMap:
      name: host-extra-config
extraVolumeMounts:
  - name: extra-config
    mountPath: /etc/nrdot-collector/extra
```

and an additional entry in `command.extraArgs`:

```yaml
- --config=/etc/nrdot-collector/extra/host-internal-logs-config.yaml
```

`internal-collector-values.yaml` adds one entry to `command.extraArgs`:

```yaml
- '"--config=yaml:service::telemetry::logs::sampling::enabled: false"'
```

`spec-local.yaml` adds two new anchor blocks at the top alongside the existing ones:

```yaml
.host_log_assertions: &host_log_assertions
  - query: |
      FROM Log SELECT count(*) AS unexpected_warnings
      WHERE service.name = 'nrdot-collector-e2e-host'
        AND testKey = '${SCENARIO_TAG}'
        AND severity_text IN ('WARN','ERROR')
    expected_results:
      - key: unexpected_warnings
        upperBoundedValue: 0

.internal_log_assertions: &internal_log_assertions
  - query: |
      FROM Log SELECT count(*) AS unexpected_warnings
      WHERE service.name = 'nrdot-collector-e2e-internal'
        AND testKey = '${SCENARIO_TAG}'
        AND severity_text IN ('WARN','ERROR')
    expected_results:
      - key: unexpected_warnings
        upperBoundedValue: 0
```

and references each anchor from its scenario's `tests.nrqls` list:

```yaml
nrqls:
  - *host_metrics_nrqls
  - *host_log_assertions
```

(YAML list-of-lists is supported by the action's spec; if not, we'll inline the entries instead.)

## Iteration Plan

The user will iterate locally; final allow-list contents are discovered, not designed.

1. Implement the file/spec changes with empty allow-lists.
2. Create a throwaway workflow `.github/workflows/ci-warnlogs-iterate.yaml` (untracked while iterating; deleted before PR): free-disk → setup-go → cached source-gen → cached goreleaser → kind → load-image → e2e action. Skips FIPS, GPG signing, Trivy, artifact upload to keep iteration fast on `act`.
3. `act push -W .github/workflows/ci-warnlogs-iterate.yaml` with the staging ingest key supplied via a gitignored `.tmp/act-secrets` file. First run cold; subsequent runs hit caches.
4. **Validate the assertion form** with a discovery NRQL via the `newrelic` CLI (koala/staging profile) before treating any failure as real:
   ```sql
   FROM Log SELECT uniques(severity_text), uniques(severity_number)
   WHERE service.name = 'nrdot-collector-e2e-host'
     AND testKey = '<scenario tag>'
   SINCE 30 minutes ago
   ```
   If the attribute name or value casing differs, correct the assertion form before continuing.
5. Inspect the actual log lines via NR; for each genuine expected warning, add a `LIKE` clause to the allow-list with a comment naming the source. Re-run until green.
6. Repeat 4–5 for the internal scenario.
7. Delete the throwaway workflow. Commit the final state.

## Failure Modes

- **Allow-list too narrow** — a genuine expected warning fails the assertion. Fix: add a pattern with a comment.
- **Allow-list too broad** — a real new warning is silently swallowed. Mitigation: keep patterns anchored prefixes (no broad `%word%` middles); reviewer checks new entries.
- **NR ingestion lag** — assertion runs before the log line lands. Existing action retry config (`retry_seconds: 15`, `retry_attempts: 20`, ~5 min window) covers this.
- **Sampling-off floods NR with INFO logs** — accepted cost for the test scenario; documented in the new file's header comment. Production users continue to use `examples/internal-telemetry-config.yaml` with sampling enabled.
- **Severity attribute name differs from `severity_text`** — caught by the discovery NRQL in step 4 of the iteration plan, before the assertion is finalised.

## Out of Scope / Follow-ups

- Same assertion in the nightly E2E specs (`spec-nightly-kind.yaml`, `spec-nightly-ubuntu.yaml`, `spec-nightly-windows.yaml`). Likely valuable; not in this PR.
- Same assertion for `nrdot-collector-experimental`.
- Extracting the NRQL/allow-list into a shared YAML loaded from multiple specs. Premature until we have a second consumer.
