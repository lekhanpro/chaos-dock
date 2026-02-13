#!/usr/bin/env bash
set -euo pipefail

action="${1:-tui}"
config="${2:-chaos.yaml}"
targets="${3:-}"

cmd=(go run ./cmd/chaos-dock)

case "$action" in
  init)
    cmd+=(-init-config -config "$config")
    ;;
  validate)
    cmd+=(-validate-config -config "$config")
    ;;
  list)
    cmd+=(-list)
    ;;
  run-once)
    cmd+=(-run-once -config "$config")
    ;;
  run-scheduled)
    cmd+=(-run-scheduled -config "$config")
    ;;
  panic)
    cmd+=(-panic)
    if [[ -n "$targets" ]]; then
      cmd+=(-targets "$targets")
    fi
    ;;
  tui)
    ;;
  *)
    echo "Unknown action: $action"
    echo "Valid: init | validate | list | run-once | run-scheduled | panic | tui"
    exit 1
    ;;
esac

echo "Executing: ${cmd[*]}"
"${cmd[@]}"

