package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		cfg          container.Config
		wantLogEntry []LogEntry
		wantError    string
	}{
		{
			container.Config{},
			nil,
			"failed to ensure image exists: failed to pull image: invalid reference format",
		},
		{
			container.Config{
				Image: "alpine",
				Cmd:   []string{"sh", "-c", "echo hello; sleep 1; echo world >&2; sleep 1; echo hello"},
			},
			[]LogEntry{{Stream: "stdout", Log: "hello\n"}, {Stream: "stderr", Log: "world\n"}, {Stream: "stdout", Log: "hello\n"}},
			"",
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			out, err := Run(tc.cfg)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
			require.Equal(t, tc.wantLogEntry, out)

		})
	}
}
