package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ozbe/aws-vault-proxy/server"
)

const (
	defaultNetwork = "tcp"
	defaultPort    = "7654"
)

const (
	commandEnvKey = "AWS_VAULT_PROXY_COMMAND"
	portEnvKey    = "AWS_VAULT_PROXY_PORT"
)

func main() {
	c := newConfig()

	s := server.New(c)
	err := s.Listen()

	if err != nil {
		log.Fatal(err)
	}
}

func newConfig() server.Config {
	var command string
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
		Address: fmt.Sprintf(":%s", port),
	}
}
