package box

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

func (b *Box) precheck() error {
	ctx := context.Background()
	if err := b.ensureImageExists(ctx); err != nil {
		return err
	}

	resp, err := b.cli.ContainerCreate(ctx, &b.opts.Config, &b.opts.HostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	b.containerID = resp.ID
	return nil
}

func (b *Box) postcheck() {
	if err := b.cli.ContainerRemove(context.Background(), b.containerID, container.RemoveOptions{Force: true}); err != nil {
		fmt.Printf("failed to remove container: %v\n", err)
	}
}

func (b *Box) ensureImageExists(ctx context.Context) error {
	out, err := b.cli.ImagePull(ctx, b.opts.Config.Image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("imagePull err: %w", err)
	}
	defer out.Close()

	if _, err := io.ReadAll(out); err != nil {
		return fmt.Errorf("reading pull output err: %w", err)
	}
	return nil
}
