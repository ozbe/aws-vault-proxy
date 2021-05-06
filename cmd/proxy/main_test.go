package main

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultTestCommand = "../../bin/fake-vault"
)

// func TestDefault(t *testing.T) {
// 	proxy := &Proxy{defaultTestCommand}
// 	profile := "default"
// 	var env []string
// 	err := proxy.Env(&profile, &env)
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{"AWS_SUCCESS=true"}, env)
// }

func TestRequest(t *testing.T) {
	conf := config{
		command: defaultTestCommand,
		network: defaultNetwork,
		port:    defaultPort,
	}
	var r bytes.Buffer
	var w bytes.Buffer

	req := Cmd{
		Args: []string{"exec", "mfa", "--", "env"},
	}
	encoder := gob.NewEncoder(&r)
	encoder.Encode(req)
	r.WriteString("12141")

	err := handleRequest(conf, &r, &w)

	assert.NoError(t, err)
	assert.Equal(t, "Enter token for arn:aws:iam::1234564789:mfa/USER: AWS_SUCCESS=true\n", w.String())
}
