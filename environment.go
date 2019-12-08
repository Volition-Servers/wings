package main

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pterodactyl/wings/config"
	"go.uber.org/zap"
)

// Configures the required network for the docker environment.
func ConfigureDockerEnvironment(c *config.DockerConfiguration) error {
	// Ensure the required docker network exists on the system.
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	resource, err := cli.NetworkInspect(context.Background(), c.Network.Name, types.NetworkInspectOptions{})
	if err != nil && client.IsErrNotFound(err) {
		zap.S().Infow("creating missing pterodactyl0 interface, this could take a few seconds...")
		return createDockerNetwork(cli, c)
	} else if err != nil {
		zap.S().Fatalw("failed to create required docker network for containers", zap.Error(err))
	}

	switch resource.Driver {
	case "host":
		c.Network.Interface = "127.0.0.1"
		c.Network.ISPN = false
		return nil
	case "overlay":
	case "weavemesh":
		c.Network.Interface = ""
		c.Network.ISPN = true
		return nil
	default:
		c.Network.ISPN = false
	}

	return nil
}

// Creates a new network on the machine if one does not exist already.
func createDockerNetwork(cli *client.Client, c *config.DockerConfiguration) error {
	_, err := cli.NetworkCreate(context.Background(), c.Network.Name, types.NetworkCreate{
		Driver:     c.Network.Driver,
		EnableIPv6: true,
		Internal:   c.Network.IsInternal,
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet:  c.Network.Interfaces.V4.Subnet,
					Gateway: c.Network.Interfaces.V4.Gateway,
				},
				{
					Subnet:  c.Network.Interfaces.V6.Subnet,
					Gateway: c.Network.Interfaces.V6.Gateway,
				},
			},
		},
		Options: map[string]string{
			"encryption": "false",
			"com.docker.network.bridge.default_bridge":       "false",
			"com.docker.network.bridge.enable_icc":           "true",
			"com.docker.network.bridge.enable_ip_masquerade": "true",
			"com.docker.network.bridge.host_binding_ipv4":    "0.0.0.0",
			"com.docker.network.bridge.name":                 "pterodactyl0",
			"com.docker.network.driver.mtu":                  "1500",
		},
	})

	if err != nil {
		return err
	}

	switch c.Network.Driver {
	case "host":
		c.Network.Interface = "127.0.0.1"
		c.Network.ISPN = false
		break
	case "overlay":
	case "weavemesh":
		c.Network.Interface = ""
		c.Network.ISPN = true
		break
	default:
		c.Network.Interface = c.Network.Interfaces.V4.Gateway
		c.Network.ISPN = false
		break
	}

	return nil
}
