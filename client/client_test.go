package client_test

import (
	"os"
	"testing"

	"github.com/ozbe/aws-vault-proxy/client"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	args := []string{"exec", "--", "env"}

	c := client.New("tcp", ":7654")
	cmd := c.Cmd(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	assert.NoError(t, cmd.Run())
}
