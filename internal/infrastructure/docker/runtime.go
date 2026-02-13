package docker

import (
	"context"
	"fmt"
	"strings"

	apicontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Runtime struct {
	client *client.Client
}

type ContainerSummary struct {
	ID     string
	Name   string
	Image  string
	Status string
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

func (r *Runtime) ListRunningContainers(ctx context.Context) ([]ContainerSummary, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("docker runtime client is not initialized")
	}

	fl := filters.NewArgs()
	fl.Add("status", "running")

	containers, err := r.client.ContainerList(ctx, apicontainer.ListOptions{
		All:     false,
		Filters: fl,
	})
	if err != nil {
		return nil, fmt.Errorf("list running containers: %w", err)
	}

	out := make([]ContainerSummary, 0, len(containers))
	for _, c := range containers {
		name := ""
		for _, n := range c.Names {
			name = strings.TrimPrefix(strings.TrimSpace(n), "/")
			if name != "" {
				break
			}
		}
		if name == "" {
			name = c.ID
		}

		out = append(out, ContainerSummary{
			ID:     c.ID,
			Name:   name,
			Image:  c.Image,
			Status: c.Status,
		})
	}

	return out, nil
}

func (r *Runtime) Restart(ctx context.Context, containerID string) error {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("container id is required")
	}
	if r == nil || r.client == nil {
		return fmt.Errorf("docker runtime client is not initialized")
	}

	if err := r.client.ContainerRestart(ctx, containerID, apicontainer.StopOptions{}); err != nil {
		return fmt.Errorf("restart container %q: %w", containerID, err)
	}

	return nil
}

func (r *Runtime) Kill(ctx context.Context, containerID string, signal string) error {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("container id is required")
	}
	if r == nil || r.client == nil {
		return fmt.Errorf("docker runtime client is not initialized")
	}

	if err := r.client.ContainerKill(ctx, containerID, signal); err != nil {
		return fmt.Errorf("kill container %q with signal %q: %w", containerID, signal, err)
	}

	return nil
}
