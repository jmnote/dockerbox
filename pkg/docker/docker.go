package docker

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func ensureImageExists(cli *client.Client, ctx context.Context, imageName string) error {
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()

	// wait for pulling
	if _, err := io.ReadAll(out); err != nil {
		return fmt.Errorf("error reading pull output: %w", err)
	}
	return nil
}

func Run(cfg container.Config) ([]LogEntry, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	if err := ensureImageExists(cli, ctx, cfg.Image); err != nil {
		return nil, fmt.Errorf("failed to ensure image exists: %w", err)
	}

	resp, err := cli.ContainerCreate(ctx, &cfg, nil, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	defer func() {
		if rmErr := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); rmErr != nil {
			fmt.Printf("warning: failed to remove container: %v\n", rmErr)
		}
	}()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	if err := waitContainer(ctx, cli, resp.ID); err != nil {
		return nil, fmt.Errorf("failed to wait for container to stop: %w", err)
	}

	logEntries, err := getContainerLogs(ctx, cli, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	return logEntries, nil
}

func getContainerLogs(ctx context.Context, cli *client.Client, containerID string) ([]LogEntry, error) {
	out, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	logEntries := parseLogEntries(out)

	return logEntries, nil
}

func parseLogEntries(out io.Reader) []LogEntry {
	var logEntries []LogEntry
	data, err := io.ReadAll(out)
	if err != nil {
		fmt.Printf("error reading logs: %v\n", err)
		return logEntries
	}

	for len(data) > 0 {
		streamType := data[0]
		var stream string
		switch streamType {
		case 1:
			stream = "stdout"
		case 2:
			stream = "stderr"
		default:
			fmt.Printf("unknown stream type: %v\n", streamType)
			return logEntries
		}

		msgLength := binary.BigEndian.Uint32(data[4:])
		msg := data[8 : 8+msgLength]

		logEntry := LogEntry{
			Stream: stream,
			Log:    string(msg),
		}
		logEntries = append(logEntries, logEntry)

		data = data[8+msgLength:]
	}

	return logEntries
}

func waitContainer(ctx context.Context, cli *client.Client, containerID string) error {
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to wait for container: %w", err)
		}
	case <-statusCh:
	}
	return nil
}
