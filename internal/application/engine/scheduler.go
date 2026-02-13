package engine

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	domainconfig "github.com/lekhanpro/chaos-dock/internal/domain/config"
)

type scheduledExperiment struct {
	experiment domainconfig.Experiment
	every      time.Duration
	jitter     time.Duration
}

// RunScheduled continuously executes enabled experiments until ctx is canceled.
func (r *Runner) RunScheduled(ctx context.Context, cfg domainconfig.ChaosConfig, onResult func(ExperimentResult)) error {
	if r == nil {
		return fmt.Errorf("runner is required")
	}
	if len(cfg.Experiments) == 0 {
		return fmt.Errorf("config has no experiments")
	}

	scheduled, err := parseSchedule(cfg.Experiments)
	if err != nil {
		return err
	}
	if len(scheduled) == 0 {
		return fmt.Errorf("config has no enabled experiments")
	}

	var wg sync.WaitGroup
	for _, job := range scheduled {
		job := job
		wg.Add(1)

		go func() {
			defer wg.Done()

			timer := time.NewTimer(nextInterval(job.every, job.jitter))
			defer timer.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					res := r.ExecuteExperiment(ctx, job.experiment)
					if onResult != nil {
						onResult(res)
					}

					timer.Reset(nextInterval(job.every, job.jitter))
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func parseSchedule(experiments []domainconfig.Experiment) ([]scheduledExperiment, error) {
	scheduled := make([]scheduledExperiment, 0, len(experiments))
	var errs []error

	for _, exp := range experiments {
		if !exp.Enabled {
			continue
		}

		every, err := time.ParseDuration(exp.Schedule.Every)
		if err != nil {
			errs = append(errs, fmt.Errorf("experiment %q has invalid schedule.every: %w", exp.Name, err))
			continue
		}
		if every <= 0 {
			errs = append(errs, fmt.Errorf("experiment %q schedule.every must be greater than zero", exp.Name))
			continue
		}

		var jitter time.Duration
		if exp.Schedule.Jitter != "" {
			jitter, err = time.ParseDuration(exp.Schedule.Jitter)
			if err != nil {
				errs = append(errs, fmt.Errorf("experiment %q has invalid schedule.jitter: %w", exp.Name, err))
				continue
			}
			if jitter < 0 {
				errs = append(errs, fmt.Errorf("experiment %q schedule.jitter must be zero or positive", exp.Name))
				continue
			}
		}

		scheduled = append(scheduled, scheduledExperiment{
			experiment: exp,
			every:      every,
			jitter:     jitter,
		})
	}

	return scheduled, errors.Join(errs...)
}

func nextInterval(every time.Duration, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return every
	}

	return every + time.Duration(rand.Int63n(int64(jitter)+1))
}

