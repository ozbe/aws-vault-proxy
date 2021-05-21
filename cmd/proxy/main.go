package main

import (
	"log"
	"os"

	"github.com/ozbe/aws-vault-proxy/server"
)

const (
	defaultCommand = "aws-vault"
	defaultNetwork = "tcp"
	defaultPort    = "7654"
)

const (
	commandEnvKey = "AWS_VAULT_PROXY_COMMAND"
	portEnvKey    = "AWS_VAULT_PROXY_PORT"
)

func main() {
	c := newConfig()

	err := server.Listen(c)

	if err != nil {
		log.Fatal(err)
	}
}

func newConfig() server.Config {
	command := defaultCommand
	if val, ok := os.LookupEnv(commandEnvKey); ok {
		command = val
	}

	port := defaultPort
	if val, ok := os.LookupEnv(portEnvKey); ok {
		port = val
	}

	return server.Config{
		Command: command,
		Network: defaultNetwork,
		Port:    port,
	}
}
