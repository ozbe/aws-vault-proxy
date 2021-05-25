package proxy_test

import (
	"net"
	"os"
	"testing"

	"github.com/ozbe/aws-vault-proxy/proxy"
	"github.com/stretchr/testify/assert"
)

const (
	defaultTestCommand = "../bin/fake-vault"
)

func TestAll(t *testing.T) {
	network := "tcp"
	host := ""
	port := "7654"
	args := []string{"exec", "--", "env"}

	s := &proxy.Server{
		Command: defaultTestCommand,
		Network: network,
		Address: net.JoinHostPort(host, port),
	}
	go s.Listen()
	defer s.Close()

	c := proxy.NewClient(network, net.JoinHostPort(host, port))
	cmd := c.Cmd(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	assert.NoError(t, cmd.Run())
}
