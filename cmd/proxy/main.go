package main

import (
	"log"
	"net"
	"os"
	"os/exec"

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
	err := s.Run()
	defer s.Close()

	if err != nil {
		log.Fatal(err)
	}
}

func newServer() proxy.Server {
	command := proxy.DefaultCommand
	if val, ok := os.LookupEnv(commandEnvKey); ok {
		command = customCommand(val)
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

func customCommand(name string) proxy.Command {
	return func(args ...string) *exec.Cmd {
		return exec.Command(name, args...)
	}
}
