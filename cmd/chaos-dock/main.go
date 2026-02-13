package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lekhanpro/chaos-dock/internal/application/engine"
	"github.com/lekhanpro/chaos-dock/internal/application/safety"
	"github.com/lekhanpro/chaos-dock/internal/application/ui"
	domainconfig "github.com/lekhanpro/chaos-dock/internal/domain/config"
	configinfra "github.com/lekhanpro/chaos-dock/internal/infrastructure/config"
	dockerinfra "github.com/lekhanpro/chaos-dock/internal/infrastructure/docker"
	faultinfra "github.com/lekhanpro/chaos-dock/internal/infrastructure/fault"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	opts := parseFlags()

	if !opts.runOnce && !opts.runScheduled && !opts.panic && !opts.list && !opts.initConfig && !opts.validateConfig {
		runTUI(ctx)
		return
	}

	if opts.initConfig {
		if err := writeDefaultConfig(opts.configPath, opts.force); err != nil {
			log.Fatalf("initialize config: %v", err)
		}
		log.Printf("created %s", opts.configPath)
		return
	}

	if opts.validateConfig {
		cfg, err := configinfra.LoadChaosConfig(opts.configPath)
		if err != nil {
			log.Fatalf("config validation failed: %v", err)
		}
		log.Printf("config validation successful: %d experiments", len(cfg.Experiments))
		for _, exp := range cfg.Experiments {
			status := "disabled"
			if exp.Enabled {
				status = "enabled"
			}
			log.Printf("- %s [%s] target=%s fault=%s every=%s", exp.Name, status, exp.TargetContainer, exp.Fault.Type, exp.Schedule.Every)
		}
		return
	}

	runtime, err := dockerinfra.NewRuntimeFromEnv()
	if err != nil {
		log.Fatalf("initialize docker runtime: %v", err)
	}
	defer runtime.Close()

	latencyInjector := faultinfra.NewNetworkLatencyInjector(runtime)
	killInjector := faultinfra.NewContainerKillInjector(runtime)
	registry := safety.NewTargetRegistry()

	runner := &engine.Runner{
		Injector: latencyInjector,
		Killer:   killInjector,
		Tracker:  registry,
	}

	panicButton := &safety.PanicButton{
		Injector:  latencyInjector,
		Restarter: runtime,
		Registry:  registry,
	}

	if opts.list {
		listContainers(ctx, runtime)
		return
	}

	if opts.panic {
		if err := panicButton.Trigger(ctx, splitCSV(opts.targets)); err != nil {
			log.Fatalf("panic rollback failed: %v", err)
		}
		log.Println("panic rollback completed")
		return
	}

	cfg, err := configinfra.LoadChaosConfig(opts.configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	switch {
	case opts.runOnce:
		runOnce(ctx, runner, cfg)
	case opts.runScheduled:
		runScheduled(ctx, runner, cfg)
	}
}

type runOptions struct {
	configPath     string
	runOnce        bool
	runScheduled   bool
	panic          bool
	targets        string
	list           bool
	initConfig     bool
	validateConfig bool
	force          bool
}

func parseFlags() runOptions {
	var opts runOptions

	flag.StringVar(&opts.configPath, "config", "chaos.yaml", "path to chaos experiment config")
	flag.BoolVar(&opts.runOnce, "run-once", false, "execute enabled experiments exactly once")
	flag.BoolVar(&opts.runScheduled, "run-scheduled", false, "execute enabled experiments continuously by schedule")
	flag.BoolVar(&opts.panic, "panic", false, "revert network faults and restart containers")
	flag.StringVar(&opts.targets, "targets", "", "comma-separated container IDs/names used by -panic")
	flag.BoolVar(&opts.list, "list", false, "list running containers from the Docker daemon")
	flag.BoolVar(&opts.initConfig, "init-config", false, "create a starter chaos config at -config path")
	flag.BoolVar(&opts.validateConfig, "validate-config", false, "validate chaos config and print experiment summary")
	flag.BoolVar(&opts.force, "force", false, "allow overwrite when used with -init-config")
	flag.Parse()

	if opts.runOnce && opts.runScheduled {
		log.Fatalf("choose exactly one of -run-once or -run-scheduled")
	}
	if opts.initConfig && opts.validateConfig {
		log.Fatalf("choose exactly one of -init-config or -validate-config")
	}
	if opts.initConfig && (opts.runOnce || opts.runScheduled || opts.panic || opts.list) {
		log.Fatalf("-init-config cannot be combined with runtime fault commands")
	}
	if opts.validateConfig && (opts.runOnce || opts.runScheduled || opts.panic || opts.list) {
		log.Fatalf("-validate-config cannot be combined with runtime fault commands")
	}

	return opts
}

func runTUI(ctx context.Context) {
	model := ui.NewModel(ctx)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatalf("run tui: %v", err)
	}
}

func runOnce(ctx context.Context, runner *engine.Runner, cfg domainconfig.ChaosConfig) {
	results := runner.RunOnce(ctx, cfg)

	hadError := false
	for _, res := range results {
		if res.Skipped {
			log.Printf("[SKIP] %s (%s): %s", res.Name, res.FaultType, res.Message)
			continue
		}
		if res.Err != nil {
			hadError = true
			log.Printf("[FAIL] %s (%s): %v", res.Name, res.FaultType, res.Err)
			continue
		}

		log.Printf("[OK] %s (%s): %s (duration=%s)", res.Name, res.FaultType, res.Message, res.Duration().Round(10*time.Millisecond))
	}

	if hadError {
		os.Exit(1)
	}
}

func runScheduled(ctx context.Context, runner *engine.Runner, cfg domainconfig.ChaosConfig) {
	log.Println("starting scheduled chaos experiments; press Ctrl+C to stop")

	err := runner.RunScheduled(ctx, cfg, func(res engine.ExperimentResult) {
		if res.Skipped {
			log.Printf("[SKIP] %s (%s): %s", res.Name, res.FaultType, res.Message)
			return
		}
		if res.Err != nil {
			log.Printf("[FAIL] %s (%s): %v", res.Name, res.FaultType, res.Err)
			return
		}

		log.Printf("[OK] %s (%s): %s (duration=%s)", res.Name, res.FaultType, res.Message, res.Duration().Round(10*time.Millisecond))
	})
	if err != nil {
		log.Fatalf("run scheduled experiments: %v", err)
	}
}

func listContainers(ctx context.Context, runtime *dockerinfra.Runtime) {
	containers, err := runtime.ListRunningContainers(ctx)
	if err != nil {
		log.Fatalf("list containers: %v", err)
	}
	if len(containers) == 0 {
		log.Println("no running containers found")
		return
	}

	log.Println("running containers:")
	for _, c := range containers {
		shortID := c.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		log.Printf("- %s (%s) image=%s status=%s", c.Name, shortID, c.Image, c.Status)
	}
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}

	return out
}

func writeDefaultConfig(path string, force bool) error {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return fmt.Errorf("config path is required")
	}

	if !force {
		if _, err := os.Stat(cleanPath); err == nil {
			return fmt.Errorf("file %q already exists (use -force to overwrite)", cleanPath)
		}
	}

	return os.WriteFile(cleanPath, []byte(defaultChaosYAML), 0o644)
}

const defaultChaosYAML = `experiments:
  - name: db-latency
    targetContainer: postgres
    enabled: true
    fault:
      type: network-latency
      delay: 500ms
    schedule:
      every: 60s
      jitter: 5s

  - name: api-latency
    targetContainer: api
    enabled: true
    fault:
      type: network-latency
      delay: 120ms
    schedule:
      every: 30s
      jitter: 3s

  - name: kill-db
    targetContainer: postgres
    enabled: true
    fault:
      type: kill
      signal: SIGTERM
    schedule:
      every: 120s
`
