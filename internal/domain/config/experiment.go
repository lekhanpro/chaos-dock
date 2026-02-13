package config

type ChaosConfig struct {
	Experiments []Experiment `yaml:"experiments"`
}

type Experiment struct {
	Name            string   `yaml:"name"`
	TargetContainer string   `yaml:"targetContainer"`
	Enabled         bool     `yaml:"enabled"`
	Fault           Fault    `yaml:"fault"`
	Schedule        Schedule `yaml:"schedule"`
}

type Fault struct {
	Type  string `yaml:"type"`            // network-latency | kill
	Delay string `yaml:"delay,omitempty"` // e.g. 500ms
}

type Schedule struct {
	Every  string `yaml:"every"`            // e.g. 60s
	Jitter string `yaml:"jitter,omitempty"` // e.g. 5s
}

