package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
)

type Runtime struct {
	client *client.Client
}

func NewRuntimeFromEnv() (*Runtime, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Runtime{client: c}, nil
}

func (r *Runtime) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Close()
}

func (r *Runtime) ContainerPID(ctx context.Context, containerID string) (int, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return 0, fmt.Errorf("container id is required")
	}

	inspect, err := r.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, fmt.Errorf("inspect container %q: %w", containerID, err)
	}
	if inspect.State == nil || !inspect.State.Running {
		return 0, fmt.Errorf("container %q is not running", containerID)
	}
	if inspect.State.Pid <= 0 {
		return 0, fmt.Errorf("container %q has invalid pid %d", containerID, inspect.State.Pid)
	}

	return inspect.State.Pid, nil
}

