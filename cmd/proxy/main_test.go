package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultTestCommand = "./bin/fake-vault"
)

func TestDefault(t *testing.T) {
	proxy := &Proxy{defaultTestCommand}
	profile := "default"
	var env []string
	err := proxy.Env(&profile, &env)
	assert.NoError(t, err)
	assert.Equal(t, []string{"AWS_SUCCESS=true"}, env)
}

// func TestMfa(t *testing.T) {
// 	proxy := &Proxy{defaultTestCommand}
// 	profile := "mfa"
// 	var env []string
// 	err := proxy.Env(&profile, &env)
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{"AWS_SUCCESS=true"}, env)
// }
