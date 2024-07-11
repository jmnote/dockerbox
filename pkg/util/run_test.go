package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	testcases := []struct {
		command            string
		wantStdout         string
		wantStderrContains string
		wantExitCode       int
	}{
		{"echo hello", "hello\n", "", 0},
		{"cat /asdf/asdf/asdf", "", "No such file or directory\n", 1},
		{"docker run --rm hello-world", "\nHello from Docker!\nThis message shows that your installation appears to be working correctly.\n\nTo generate this message, Docker took the following steps:\n 1. The Docker client contacted the Docker daemon.\n 2. The Docker daemon pulled the \"hello-world\" image from the Docker Hub.\n    (amd64)\n 3. The Docker daemon created a new container from that image which runs the\n    executable that produces the output you are currently reading.\n 4. The Docker daemon streamed that output to the Docker client, which sent it\n    to your terminal.\n\nTo try something more ambitious, you can run an Ubuntu container with:\n $ docker run -it ubuntu bash\n\nShare images, automate workflows, and more with a free Docker ID:\n https://hub.docker.com/\n\nFor more examples and ideas, visit:\n https://docs.docker.com/get-started/\n\n", "", 0},
		{"docker run --rm not-exists", "", "Unable to find image 'not-exists:latest' locally\ndocker: Error response from daemon: pull access denied for not-exists, repository does not exist or may require 'docker login': denied: requested access to the resource is denied.\nSee 'docker run --help'.\n", 125},
	}
	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			stdout, stderr, exitCode := Run(tc.command)
			require.Equal(t, tc.wantStdout, stdout)
			require.Contains(t, stderr, tc.wantStderrContains)
			require.Equal(t, tc.wantExitCode, exitCode)
		})
	}
}
