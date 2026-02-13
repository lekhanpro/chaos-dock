package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	domainconfig "github.com/lekhanpro/chaos-dock/internal/domain/config"
	"gopkg.in/yaml.v3"
)

func LoadChaosConfig(path string) (domainconfig.ChaosConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return domainconfig.ChaosConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg domainconfig.ChaosConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return domainconfig.ChaosConfig{}, fmt.Errorf("parse yaml %q: %w", path, err)
	}

	if err := validate(cfg); err != nil {
		return domainconfig.ChaosConfig{}, err
	}

	return cfg, nil
}

func validate(cfg domainconfig.ChaosConfig) error {
	if len(cfg.Experiments) == 0 {
		return fmt.Errorf("config requires at least one experiment")
	}

	for i, exp := range cfg.Experiments {
		if strings.TrimSpace(exp.Name) == "" {
			return fmt.Errorf("experiments[%d].name is required", i)
		}
		if strings.TrimSpace(exp.TargetContainer) == "" {
			return fmt.Errorf("experiments[%d].targetContainer is required", i)
		}
		if strings.TrimSpace(exp.Fault.Type) == "" {
			return fmt.Errorf("experiments[%d].fault.type is required", i)
		}
		if strings.TrimSpace(exp.Schedule.Every) == "" {
			return fmt.Errorf("experiments[%d].schedule.every is required", i)
		}
		if _, err := time.ParseDuration(exp.Schedule.Every); err != nil {
			return fmt.Errorf("experiments[%d].schedule.every must be a valid duration: %w", i, err)
		}
		if exp.Schedule.Jitter != "" {
			if _, err := time.ParseDuration(exp.Schedule.Jitter); err != nil {
				return fmt.Errorf("experiments[%d].schedule.jitter must be a valid duration: %w", i, err)
			}
		}

		switch exp.Fault.Type {
		case "network-latency":
			if strings.TrimSpace(exp.Fault.Delay) == "" {
				return fmt.Errorf("experiments[%d].fault.delay is required for network-latency", i)
			}
			if _, err := time.ParseDuration(exp.Fault.Delay); err != nil {
				return fmt.Errorf("experiments[%d].fault.delay must be a valid duration: %w", i, err)
			}
		case "kill":
			if exp.Fault.Signal != "" && !isSupportedSignal(exp.Fault.Signal) {
				return fmt.Errorf("experiments[%d].fault.signal %q is not supported", i, exp.Fault.Signal)
			}
		default:
			return fmt.Errorf("experiments[%d].fault.type %q is unsupported", i, exp.Fault.Type)
		}
	}

	return nil
}

func isSupportedSignal(raw string) bool {
	signal := strings.ToUpper(strings.TrimSpace(raw))
	if signal == "" {
		return true
	}
	if !strings.HasPrefix(signal, "SIG") {
		signal = "SIG" + signal
	}

	switch signal {
	case "SIGKILL", "SIGTERM", "SIGINT", "SIGQUIT", "SIGHUP", "SIGUSR1", "SIGUSR2":
		return true
	default:
		return false
	}
}
