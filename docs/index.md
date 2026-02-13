---
layout: default
title: Chaos-Dock
---

# Chaos-Dock

Local chaos engineering for Docker workloads.

- Enumerates running containers via Docker API
- Injects network latency using Linux namespace + `tc netem`
- Supports experiment definitions via `chaos.yaml`
- Includes panic-button rollback flow

See the repository README for architecture and implementation details.

