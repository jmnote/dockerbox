package docker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var docker *Docker = New()

func TestRun(t *testing.T) {
	testCases := []struct {
		name       string
		options    Options
		wantResult *Result
		wantError  string
	}{
		{
			"",
			Options{Image: "alpine", Command: ""},
			&Result{Logs: []Log{}, ExitCode: 0},
			"",
		},
		{
			"etc",
			Options{Image: "alpine", Command: "/etc"},
			&Result{Logs: []Log{}, ExitCode: 126},
			"",
		},
		// echo
		{
			"echo",
			Options{Image: "alpine", Command: "echo foo"},
			&Result{Logs: []Log{{Log: "foo\n", Stream: "stdout"}}, ExitCode: 0},
			"",
		},
		{
			"echo",
			Options{Image: "alpine", Command: "echo foo bar"},
			&Result{Logs: []Log{{Log: "foo bar\n", Stream: "stdout"}}, ExitCode: 0},
			"",
		},
		{
			"echo",
			Options{Image: "alpine", Command: "echo foo; echo bar"},
			&Result{Logs: []Log{{Log: "foo\n", Stream: "stdout"}}, ExitCode: 0},
			"",
		},
		{
			"sleep",
			Options{Image: "alpine", Command: "sleep 1"},
			&Result{Logs: []Log{}, ExitCode: 0},
			"",
		},
		{
			"sleep",
			Options{Image: "alpine", Command: "sleep 1; echo foo"},
			&Result{Logs: []Log{}, ExitCode: 0},
			"",
		},
		{
			"sleep",
			Options{Image: "alpine", Command: "sleep 20; echo foo"},
			&Result{Logs: []Log{}, ExitCode: 0},
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := docker.Run(tc.options)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
			require.Equal(t, tc.wantResult, result)
		})
	}
}
