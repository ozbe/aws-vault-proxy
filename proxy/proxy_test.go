package proxy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	conf := Config{"tcp", ":7654"}
	args := []string{"exec", "--", "env"}

	c := NewClient(conf)
	cmd := c.Cmd(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	assert.NoError(t, cmd.Run())
}
