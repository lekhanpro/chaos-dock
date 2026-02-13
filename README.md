# Chaos-Dock

Chaos-Dock is a local chaos engineering CLI/TUI for Docker applications.
It lets you inspect running containers and inject controlled faults (latency, kill) to validate resilience.

## Phase 1 Scope

- Clean architecture skeleton (`domain`, `application`, `infrastructure`)
- YAML-based experiment definitions (`chaos.yaml`)
- Core latency injector (`tc netem`) using container network namespaces
- Panic button service for best-effort rollback + restart

## Repository Layout

```text
chaos-dock/
|-- cmd/
|   `-- chaos-dock/
|-- internal/
|   |-- application/
|   |-- domain/
|   `-- infrastructure/
|-- pkg/
|   `-- chaosdock/
`-- .github/workflows/
```

## How It Works

Chaos-Dock applies network faults at the Linux namespace level instead of modifying application code.

1. Discover containers through the Docker API.
2. Inspect a target container and read `State.Pid` (host-side process ID of the container init process).
3. Enter the target process namespace with `nsenter`:
   - `--net` enters the container network namespace.
   - `--mount` aligns with container mount namespace so binary lookup behaves like inside the container rootfs.
4. Execute `tc qdisc ... netem delay ...` through `exec.CommandContext`.
   - Example: add `500ms` latency on `eth0`.
   - Revert by removing the qdisc.
5. Panic button performs best-effort cleanup:
   - remove applied latency controls,
   - restart impacted containers.

This model gives deterministic OS-level fault injection while still orchestrating through Docker metadata.

### Sidecar Pattern (Fallback Concept)

Some minimal images (scratch/distroless) may not have `iproute2/tc`. In that case, a privileged helper container (sidecar) can be launched to join the target network namespace and run `tc` there. The injector currently surfaces this condition as a dedicated error so the application layer can trigger a sidecar fallback strategy.

## Configuration (`chaos.yaml`)

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
```

## Local Development

```bash
go run ./cmd/chaos-dock
```

## CI

GitHub Actions runs `golangci-lint` on every push and pull request.

## GitHub Pages

A Pages workflow is included and deploys `docs/` from `main`.
