package client_test

import (
	"os"
	"testing"

	"github.com/ozbe/aws-vault-proxy/client"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	conf := client.Config{"tcp", ":7654"}
	args := []string{"exec", "--", "env"}

	c := client.New(conf)
	cmd := c.Cmd(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	assert.NoError(t, cmd.Run())
}
