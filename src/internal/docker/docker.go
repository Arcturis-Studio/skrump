package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

// https://docs.docker.com/reference/api/engine/sdk/examples

type DockerClient struct {
	*client.Client
	ctx context.Context
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer func() { cli.Close() }()

	return &DockerClient{
		Client: cli,
		ctx:    context.Background(),
	}, nil
}

func (d *DockerClient) ListImages() ([]image.Summary, error) {
	images, err := d.ImageList(d.ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	d.ImagePull(d.ctx, "nginxdemos/hello", image.PullOptions{})

	return images, nil
}

func (d *DockerClient) SpawnDockerContainer(arg ...string) error {

	reader, err := d.ImagePull(d.ctx, "nginxdemos/hello", image.PullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)

	exposedPorts, boundPort, err := nat.ParsePortSpecs([]string{"55002:80"})
	if err != nil {
		return err
	}

	containerConfig := container.Config{
		Image:        "nginxdemos/hello",
		ExposedPorts: exposedPorts,
	}

	hostConfig := container.HostConfig{
		PortBindings: boundPort,
		AutoRemove:   true,
		Privileged:   false,
	}

	ctrResp, err := d.ContainerCreate(d.ctx, &containerConfig, &hostConfig, nil, nil, "")

	if err := d.ContainerStart(d.ctx, ctrResp.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := d.ContainerWait(d.ctx, ctrResp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	out, err := d.ContainerLogs(d.ctx, ctrResp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	// log.Println("%s %s (status: %s)\n", ctr.ID, ctr.Image, ctr.Status)

	return nil
}

// NOTE: This relies on HOSTNAME not being set on the containers.
// If hostnames are customized, there is a chance there will be a collision here.
func (d *DockerClient) CleanUpContainers() error {
	fmt.Println("Cleaning up existing containers")
	selfID, err := os.Hostname()
	if err != nil {
		// TODO: Either panic here or log hostname error
		return err
	}
	fmt.Printf("Found selfID (hostname) as: %s\n", selfID)
	// TODO: Logging
	containers, err := d.GetContainers()
	if err != nil {
		return nil
	}

	for _, c := range containers {
		if !strings.HasPrefix(c.ID, selfID) {
			fmt.Printf("Stopping container with ID: %s\n", c.ID)
			if err := d.StopContainer(c.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DockerClient) GetContainers() ([]types.Container, error) {
	containers, err := d.ContainerList(d.ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	// TODO: Log container get

	return containers, nil
}

func (d *DockerClient) StopContainer(id string) error {
	// TODO: Log container termination
	if err := d.ContainerStop(d.ctx, id, container.StopOptions{Timeout: nil}); err != nil {
		return err
	}

	return nil
}
