package proxy_test

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"

	"github.com/ozbe/aws-vault-proxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	if os.Getenv("GO_TEST_MOCK_COMMAND") == "1" {
		mockCommand()
		return
	}

	network := "tcp"
	host := ""
	port := "7654"
	args := []string{"exec", "--", "env"}

	s := &proxy.Server{
		Command: serverCommand(t),
		Network: network,
		Address: net.JoinHostPort(host, port),
	}
	s.Run()
	defer s.Close()

	var stdout, stderr, stdin bytes.Buffer
	c := proxy.NewClient(network, net.JoinHostPort(host, port))
	cmd := c.Cmd(args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = &stdin
	assert.NoError(t, cmd.Run())
	assert.Equal(t, 0, *cmd.ExitCode)
	assert.Equal(t, "ENV=debug\nAWS_SUCCESS=true\n", stdout.String())
	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "", stderr.String())
}

func serverCommand(t *testing.T) proxy.Command {
	return func(args ...string) *exec.Cmd {
		args = append([]string{fmt.Sprintf("-test.run=%s", t.Name()), "--"}, args...)
		cmd := exec.Command(os.Args[0], args...)
		cmd.Env = append(os.Environ(), "GO_TEST_MOCK_COMMAND=1")
		return cmd
	}
}

func mockCommand() {
	args := os.Args[3:]
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Invalid number of arguments %d\n", len(args))
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "exec":
		if args[0] == "--" && args[1] == "env" {
			fmt.Println("ENV=debug\nAWS_SUCCESS=true")
		} else {
			fmt.Fprintf(os.Stderr, "Unknown exec args %q\n", args)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown cmd %q\n", cmd)
		os.Exit(2)
	}

	os.Exit(0)
}
