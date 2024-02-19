package docker

import (
	"context"
	"fmt"
	"github.com/pterodactyl/wings/environment"
	"os/exec"
)

// Get Local IP from Docker Container Inspection
func (e *Environment) GetLocalIP(ctx context.Context) (string, error) {
	if c, err := e.ContainerInspect(ctx); err != nil {
		return "", err
	} else {
		if c.State.Running {
			e.SetState(environment.ProcessRunningState)

			return c.NetworkSettings.Networks["pterodactyl_nw"].IPAddress, nil
		}
	}

	return "", fmt.Errorf("serverNotRunningError")
}

// Remove rules from the server
func (e *Environment) RemoveRules() {
	err := exec.Command("/bin/sh", "-c", fmt.Sprintf("iptables-save | sed -r '/DOCKER.*comment.*%s/s/-A/iptables -D/e'", e.Id)).Run()
	if err != nil {
		fmt.Println("Failed to remove rules")
	}
}
