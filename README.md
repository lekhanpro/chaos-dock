# Chaos-Dock

Chaos-Dock is a local chaos engineering toolkit for Docker workloads. It is designed as a production-style Go project with clean architecture, explicit safety controls, and reproducible experiment definitions.

Primary goals:

- Validate resilience behavior before deployment.
- Make fault injection explicit, observable, and reversible.
- Keep implementation idiomatic for infrastructure engineering (Go, Docker SDK, Linux namespaces, `tc`).

## Why This Project

Most local testing focuses on happy-path behavior. Chaos-Dock targets the opposite: dependency failure, latency spikes, and process termination.

Instead of modifying application code, Chaos-Dock injects failures at the container and OS namespace level:

- Network faults: `tc netem` inside the target container network namespace.
- Process faults: Unix signals via Docker API (`SIGTERM`, `SIGKILL`, etc.).
- Recovery: panic-button rollback that removes qdisc state and restarts impacted containers.

## Current Features

- Container discovery from Docker daemon (`-list` mode).
- Latency injector (`network-latency`) using `nsenter` + `tc qdisc netem`.
- Kill injector (`kill`) with signal validation.
- `chaos.yaml` experiment definitions.
- One-shot execution (`-run-once`).
- Continuous schedule execution with jitter (`-run-scheduled`).
- Panic rollback (`-panic`) with target registry support.
- Bubble Tea TUI entrypoint.
- GitHub Actions CI (`golangci-lint`, `go test`) + GitHub Pages deployment.

## Repository Layout

```text
chaos-dock/
|-- cmd/
|   `-- chaos-dock/                  # CLI entrypoint
|-- internal/
|   |-- domain/
|   |   |-- config/                  # experiment model
|   |   `-- fault/                   # domain fault contracts + errors
|   |-- application/
|   |   |-- engine/                  # runner + scheduler
|   |   |-- safety/                  # panic button + target registry
|   |   `-- ui/                      # Bubble Tea model
|   `-- infrastructure/
|       |-- config/                  # YAML loader + validation
|       |-- docker/                  # Docker runtime adapter
|       `-- fault/                   # network + kill injectors
|-- pkg/
|   `-- chaosdock/                   # public version package
|-- docs/                            # GitHub Pages website
`-- .github/workflows/               # CI + Pages workflows
```

## Architecture

### Domain Layer

- `internal/domain/fault`: fault interfaces and typed domain errors.
- `internal/domain/config`: experiment schema.

No Docker or CLI dependencies exist here.

### Application Layer

- `engine.Runner`: executes experiment intent (`network-latency`, `kill`).
- `engine.RunScheduled`: recurring execution with schedule + jitter.
- `safety.PanicButton`: rollback and restart orchestration.
- `safety.TargetRegistry`: tracks impacted containers for fast rollback.

This layer coordinates use-cases and policy.

### Infrastructure Layer

- Docker runtime adapter (`ContainerPID`, `ListRunningContainers`, `Kill`, `Restart`).
- YAML config loader + validation.
- Linux latency injector (namespace entry + `tc` execution).
- Kill injector (signal normalization and delivery via Docker API).

This layer talks to the outside world.

## How It Works (OS-Level)

For network latency injection:

1. Inspect container through Docker SDK and read `State.Pid`.
2. Enter namespace with:
   - `nsenter --target <PID> --net --mount -- tc ...`
3. Apply qdisc:
   - `tc qdisc replace dev eth0 root netem delay 500ms`
4. Revert:
   - `tc qdisc del dev eth0 root`

For process kill faults:

1. Resolve target container ID/name.
2. Validate requested signal (`SIGTERM`, `SIGKILL`, etc.).
3. Call Docker API `ContainerKill`.

For panic recovery:

1. Resolve explicit or tracked target set.
2. Best-effort revert network qdisc per target.
3. Restart containers per target.
4. Clear target registry.

## Sidecar Pattern (Fallback)

Some images (distroless/scratch) do not include `iproute2` / `tc`. When that happens, a privileged helper sidecar can join the same network namespace and apply `tc` from that helper image.

Current implementation surfaces this case explicitly (`ErrIPRoute2Missing`) so a sidecar strategy can be added without changing domain contracts.

## Configuration (`chaos.yaml`)

Example:

```yaml
experiments:
  - name: db-latency
    targetContainer: postgres
    enabled: true
    fault:
      type: network-latency
      delay: 500ms
    schedule:
      every: 60s
      jitter: 5s

  - name: kill-db
    targetContainer: postgres
    enabled: true
    fault:
      type: kill
      signal: SIGTERM
    schedule:
      every: 120s
```

Supported fields:

- `experiments[].name`: required
- `experiments[].targetContainer`: required
- `experiments[].enabled`: required
- `experiments[].fault.type`: `network-latency` or `kill`
- `experiments[].fault.delay`: required for `network-latency`
- `experiments[].fault.signal`: optional for `kill`, defaults to `SIGKILL`
- `experiments[].schedule.every`: required duration
- `experiments[].schedule.jitter`: optional duration

## CLI Usage

### TUI Mode

```bash
go run ./cmd/chaos-dock
```

### List Running Containers

```bash
go run ./cmd/chaos-dock -list
```

### Run Enabled Experiments Once

```bash
go run ./cmd/chaos-dock -run-once -config chaos.yaml
```

### Run Scheduled Experiments Continuously

```bash
go run ./cmd/chaos-dock -run-scheduled -config chaos.yaml
```

### Panic Rollback

```bash
go run ./cmd/chaos-dock -panic -targets "postgres,api"
```

If `-targets` is omitted, panic rollback can use tracked targets captured during runtime.

## Safety Model

- Explicit timeout controls for host command execution (`exec.CommandContext`).
- Typed error mapping for:
  - missing namespace tooling,
  - missing `tc` / `iproute2`,
  - permission errors,
  - command timeout,
  - generic command failure.
- Best-effort rollback keeps moving even when one target fails.
- Signal validation prevents arbitrary or malformed kill requests.

## Development

Prerequisites:

- Go 1.22+
- Docker Engine running locally
- Linux host for network namespace + `tc` features
- Privileges to run namespace traffic control commands

Clone and run:

```bash
git clone https://github.com/lekhanpro/chaos-dock.git
cd chaos-dock
go run ./cmd/chaos-dock -list
```

## CI and Quality Gates

GitHub Actions:

- `.github/workflows/ci.yml`
  - `go mod tidy`
  - `golangci-lint`
  - `go test ./...`
- `.github/workflows/pages.yml`
  - deploy static site from `docs/`

## GitHub Pages

Project website:

- https://lekhanpro.github.io/chaos-dock/

Source:

- `docs/index.html`
- `docs/assets/styles.css`
- `docs/assets/app.js`

## Roadmap

Near-term:

- Fault chaining and experiment phases (warmup, blast, cooldown).
- Metrics sink integration (Prometheus/OpenTelemetry).
- TUI container graph and live event stream.
- Sidecar-assisted latency fallback mode.
- Multi-fault campaign runner with report export.

## Contributing

1. Fork the repository.
2. Create a feature branch.
3. Keep architecture boundaries intact (`domain` -> `application` -> `infrastructure`).
4. Add or update tests where behavior changes.
5. Open a pull request with rationale and risk notes.

## Version

Current project version: `0.2.0` (`pkg/chaosdock/version.go`)

