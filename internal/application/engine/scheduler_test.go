package engine

import (
	"testing"
	"time"

	domainconfig "github.com/lekhanpro/chaos-dock/internal/domain/config"
)

func TestParseSchedule_SkipsDisabled(t *testing.T) {
	experiments := []domainconfig.Experiment{
		{
			Name:            "enabled",
			TargetContainer: "api",
			Enabled:         true,
			Schedule: domainconfig.Schedule{
				Every: "30s",
			},
		},
		{
			Name:            "disabled",
			TargetContainer: "db",
			Enabled:         false,
			Schedule: domainconfig.Schedule{
				Every: "10s",
			},
		},
	}

	scheduled, err := parseSchedule(experiments)
	if err != nil {
		t.Fatalf("parseSchedule returned error: %v", err)
	}
	if len(scheduled) != 1 {
		t.Fatalf("expected 1 scheduled experiment, got %d", len(scheduled))
	}
	if scheduled[0].experiment.Name != "enabled" {
		t.Fatalf("unexpected experiment %q", scheduled[0].experiment.Name)
	}
}

func TestParseSchedule_InvalidEvery(t *testing.T) {
	experiments := []domainconfig.Experiment{
		{
			Name:            "bad",
			TargetContainer: "api",
			Enabled:         true,
			Schedule: domainconfig.Schedule{
				Every: "not-a-duration",
			},
		},
	}

	_, err := parseSchedule(experiments)
	if err == nil {
		t.Fatalf("expected parseSchedule error")
	}
}

func TestNextInterval_Range(t *testing.T) {
	every := 10 * time.Second
	jitter := 4 * time.Second

	for i := 0; i < 25; i++ {
		next := nextInterval(every, jitter)
		if next < every || next > every+jitter {
			t.Fatalf("next interval out of range: %s", next)
		}
	}
}
