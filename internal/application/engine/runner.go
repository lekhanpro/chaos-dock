package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainconfig "github.com/lekhanpro/chaos-dock/internal/domain/config"
	"github.com/lekhanpro/chaos-dock/internal/domain/fault"
)

type TargetTracker interface {
	Mark(containerID string)
}

type Runner struct {
	Injector fault.FaultInjector
	Killer   fault.ContainerKiller
	Tracker  TargetTracker
}

type ExperimentResult struct {
	Name            string
	TargetContainer string
	FaultType       string
	StartedAt       time.Time
	FinishedAt      time.Time
	Skipped         bool
	Message         string
	Err             error
}

func (r ExperimentResult) Duration() time.Duration {
	if r.FinishedAt.Before(r.StartedAt) {
		return 0
	}
	return r.FinishedAt.Sub(r.StartedAt)
}

func (r *Runner) ApplyNetworkLatency(ctx context.Context, containerID string, delay time.Duration) error {
	if r.Injector == nil {
		return fmt.Errorf("fault injector is not configured")
	}

	return r.Injector.InjectNetworkLatency(ctx, containerID, delay)
}

func (r *Runner) ExecuteExperiment(ctx context.Context, exp domainconfig.Experiment) ExperimentResult {
	res := ExperimentResult{
		Name:            exp.Name,
		TargetContainer: exp.TargetContainer,
		FaultType:       exp.Fault.Type,
		StartedAt:       time.Now().UTC(),
	}
	defer func() {
		res.FinishedAt = time.Now().UTC()
	}()

	if !exp.Enabled {
		res.Skipped = true
		res.Message = "experiment is disabled"
		return res
	}

	target := strings.TrimSpace(exp.TargetContainer)
	if target == "" {
		res.Err = fmt.Errorf("target container is required")
		return res
	}

	switch exp.Fault.Type {
	case "network-latency":
		if r.Injector == nil {
			res.Err = fmt.Errorf("fault injector is not configured")
			return res
		}

		delay, err := time.ParseDuration(exp.Fault.Delay)
		if err != nil {
			res.Err = fmt.Errorf("parse network-latency delay %q: %w", exp.Fault.Delay, err)
			return res
		}

		if err := r.Injector.InjectNetworkLatency(ctx, target, delay); err != nil {
			res.Err = fmt.Errorf("inject network latency: %w", err)
			return res
		}

		res.Message = fmt.Sprintf("applied %s network delay to %s", delay, target)
	case "kill":
		if r.Killer == nil {
			res.Err = fmt.Errorf("container killer is not configured")
			return res
		}

		signal := strings.TrimSpace(exp.Fault.Signal)
		if err := r.Killer.KillContainer(ctx, target, signal); err != nil {
			res.Err = fmt.Errorf("kill container: %w", err)
			return res
		}

		if signal == "" {
			signal = "SIGKILL"
		}
		res.Message = fmt.Sprintf("sent %s to %s", signal, target)
	default:
		res.Err = fmt.Errorf("unsupported fault type %q", exp.Fault.Type)
		return res
	}

	if r.Tracker != nil {
		r.Tracker.Mark(target)
	}

	return res
}

func (r *Runner) RunOnce(ctx context.Context, cfg domainconfig.ChaosConfig) []ExperimentResult {
	results := make([]ExperimentResult, 0, len(cfg.Experiments))
	for _, exp := range cfg.Experiments {
		results = append(results, r.ExecuteExperiment(ctx, exp))
	}
	return results
}

