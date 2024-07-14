package box

import (
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func TestRun_error(t *testing.T) {
	testCases := []struct {
		BoxOpts   BoxOpts
		wantError string
	}{
		{
			BoxOpts{},
			"precheck err: imagePull err: invalid reference format",
		},
		{
			BoxOpts{Config: container.Config{Image: "."}},
			"precheck err: imagePull err: invalid reference format",
		},
		{
			BoxOpts{Config: container.Config{Image: "a"}},
			"precheck err: imagePull err: Error response from daemon: pull access denied for a, repository does not exist or may require 'docker login': denied: requested access to the resource is denied",
		},
	}

	for _, tc := range testCases {
		t.Run("image:"+strings.ReplaceAll(tc.BoxOpts.Config.Image, "/", "%"), func(t *testing.T) {
			got, err := Run(tc.BoxOpts)
			require.EqualError(t, err, tc.wantError)
			require.Nil(t, got)
		})
	}
}
func TestRun_ok(t *testing.T) {
	testCases := []struct {
		name    string
		BoxOpts BoxOpts
		want    *RunResult
	}{
		{
			"hello-world",
			BoxOpts{Config: container.Config{Image: "hello-world"}},
			&RunResult{
				IsTimedOut: false,
				CPU:        22404000,
				MEM:        366496,
				Time:       600,
				Logs: []LogEntry{
					{Stream: "stdout", Log: "\n"},
					{Stream: "stdout", Log: "Hello from Docker!\n"},
					{Stream: "stdout", Log: "This message shows that your installation appears to be working correctly.\n"},
					{Stream: "stdout", Log: "\n"},
					{Stream: "stdout", Log: "To generate this message, Docker took the following steps:\n"},
					{Stream: "stdout", Log: " 1. The Docker client contacted the Docker daemon.\n"},
					{Stream: "stdout", Log: " 2. The Docker daemon pulled the \"hello-world\" image from the Docker Hub.\n"},
					{Stream: "stdout", Log: "    (amd64)\n"},
					{Stream: "stdout", Log: " 3. The Docker daemon created a new container from that image which runs the\n"},
					{Stream: "stdout", Log: "    executable that produces the output you are currently reading.\n"},
					{Stream: "stdout", Log: " 4. The Docker daemon streamed that output to the Docker client, which sent it\n"},
					{Stream: "stdout", Log: "    to your terminal.\n"},
					{Stream: "stdout", Log: "\n"},
					{Stream: "stdout", Log: "To try something more ambitious, you can run an Ubuntu container with:\n"},
					{Stream: "stdout", Log: " $ docker run -it ubuntu bash\n"},
					{Stream: "stdout", Log: "\n"},
					{Stream: "stdout", Log: "Share images, automate workflows, and more with a free Docker ID:\n"},
					{Stream: "stdout", Log: " https://hub.docker.com/\n"},
					{Stream: "stdout", Log: "\n"},
					{Stream: "stdout", Log: "For more examples and ideas, visit:\n"},
					{Stream: "stdout", Log: " https://docs.docker.com/get-started/\n"},
					{Stream: "stdout", Log: "\n"},
				},
			},
		},
		{
			"echo & sleep",
			BoxOpts{Config: container.Config{
				Image: "alpine",
				Cmd:   []string{"sh", "-c", "echo hello; sleep 1; echo world >&2; sleep 1; echo hello"},
			}},
			&RunResult{
				IsTimedOut: false,
				CPU:        24337000,
				MEM:        499712,
				Time:       1500,
				Logs: []LogEntry{
					{Stream: "stdout", Log: "hello\n"},
					{Stream: "stderr", Log: "world\n"},
					{Stream: "stdout", Log: "hello\n"},
				},
			},
		},
		{
			"echo a",
			BoxOpts{Config: container.Config{
				Image: "alpine",
				Cmd:   []string{"sh", "-c", "echo a"},
			}},
			&RunResult{
				IsTimedOut: false,
				CPU:        24337000,
				MEM:        114688,
				Time:       1500,
				Logs: []LogEntry{
					{Stream: "stdout", Log: "a\n"},
				},
			},
		},
		{
			"stress-ng",
			BoxOpts{
				Config: container.Config{
					Image: "litmuschaos/stress-ng",
					Cmd:   []string{"--matrix", "1", "-t", "10m"},
				},
				Timeout: 2000,
			},
			&RunResult{
				IsTimedOut: true,
				CPU:        2612692000,
				MEM:        3043328,
				Time:       1000,
				Logs: []LogEntry{
					{Stream: "stderr", Log: "stress-ng: info:  [1] dispatching hogs: 1 matrix\n"},
				},
			},
		},
		{
			"sleep",
			BoxOpts{
				Config: container.Config{
					Image: "alpine",
					Cmd:   []string{"sleep", "infinity"},
				},
				Timeout: 1000,
			},
			&RunResult{
				IsTimedOut: true,
				CPU:        22115000,
				MEM:        327680,
				Time:       1000,
				Logs:       nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Run(tc.BoxOpts)
			require.NoError(t, err)

			require.LessOrEqual(t, got.Time, tc.want.Time*4, "time:"+tc.name)
			require.GreaterOrEqual(t, got.Time, tc.want.Time/4, "time:"+tc.name)
			require.LessOrEqual(t, got.CPU, tc.want.CPU*4, "cpu:"+tc.name)
			require.GreaterOrEqual(t, got.CPU, tc.want.CPU/4, "cpu:"+tc.name)
			require.LessOrEqual(t, got.MEM, tc.want.MEM*4, "mem:"+tc.name)
			require.GreaterOrEqual(t, got.MEM, tc.want.MEM/4, "mem:"+tc.name)
			tc.want.Time = got.Time
			tc.want.CPU = got.CPU
			tc.want.MEM = got.MEM

			require.Equal(t, tc.want, got)
		})
	}
}
