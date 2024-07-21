package docker

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Docker struct {
	client client.Client
	ctx    context.Context
}

func NewDocker() *Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &Docker{
		client: *cli,
		ctx:    context.Background(),
	}
}

func (d *Docker) Deploy(packageUrl, name, url string, port int) error {
	portStr := strconv.Itoa(port)
	containerPort := fmt.Sprintf("%s/tcp", portStr)

	config := &container.Config{
		Image: packageUrl,
		ExposedPorts: nat.PortSet{
			nat.Port(containerPort): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(containerPort): []nat.PortBinding{
				{
					HostIP:   url,
					HostPort: portStr,
				},
			},
		},
	}

	resp, err := d.client.ContainerCreate(d.ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return err
	}

	if err := d.client.ContainerStart(d.ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}

func (d *Docker) UnDeploy(name string) error {
	timeout := 10

	containerID, err := d.getContainerIDByName(name)
	if err != nil {
		return err
	}

	if err := d.client.ContainerStop(d.ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return err
	}

	if err := d.client.ContainerRemove(d.ctx, containerID, container.RemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func (d *Docker) getContainerIDByName(containerName string) (string, error) {
	filter := filters.NewArgs()
	filter.Add("name", containerName)
	var containerId string

	containers, err := d.client.ContainerList(d.ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})

	if err != nil {
		return containerId, err
	}

	if len(containers) == 0 {
		return containerId, fmt.Errorf("no container found with %s name", containerName)
	}

	return containers[0].ID, nil
}
