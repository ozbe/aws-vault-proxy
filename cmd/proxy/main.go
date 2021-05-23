package main

import (
	"log"
	"net"
	"os"

	"github.com/ozbe/aws-vault-proxy/proxy"
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
	s := newServer()
	err := s.Listen()

	if err != nil {
		log.Fatal(err)
	}
}

func newServer() proxy.Server {
	var command string
	if val, ok := os.LookupEnv(commandEnvKey); ok {
		command = val
	}

	port := defaultPort
	if val, ok := os.LookupEnv(portEnvKey); ok {
		port = val
	}

	return proxy.Server{
		Command: command,
		Network: defaultNetwork,
		Address: net.JoinHostPort("", port),
	}
}
