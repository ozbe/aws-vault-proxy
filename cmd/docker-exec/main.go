package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/ozbe/aws-vault-proxy/proxy"
)

const (
	defaultNetwork = "tcp"
	defaultHost    = "host.docker.internal"
	defaultPort    = "7654"
)

const (
	hostEnvKey = "AWS_VAULT_PROXY_HOST"
	portEnvKey = "AWS_VAULT_PROXY_PORT"
)

type config struct {
	network string
	host    string
	port    string
}

func newConfig() config {
	host := defaultHost
	if val, ok := os.LookupEnv(hostEnvKey); ok {
		host = val
	}

	port := defaultPort
	if val, ok := os.LookupEnv(portEnvKey); ok {
		port = val
	}

	return config{
		network: defaultNetwork,
		host:    host,
		port:    port,
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing command.")
	}

	c := newConfig()

	var err error
	switch os.Args[1] {
	case "exec":
		err = execCmd(c, os.Args[2:])
	default:
		log.Fatal("Unknown command.")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func execCmd(conf config, args []string) error {
	if len(args) < 3 || args[1] != "--" {
		return errors.New("invalid arguments length or format")
	}

	env, err := getAwsEnvVars(conf, args[0])
	if err != nil {
		return err
	}

	return runCmd(env, args[2:])
}

func getAwsEnvVars(conf config, profile string) ([]string, error) {
	w := NewFilterAwsEnvWriter(os.Stdout)

	p := proxy.NewClient(
		conf.network,
		net.JoinHostPort(conf.host, conf.port),
	)
	cmd := p.Cmd("exec", profile, "--", "env")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = w

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return w.Env(), nil
}

func runCmd(env []string, args []string) error {
	ec := exec.Command(args[0], args[1:]...)
	ec.Env = append(os.Environ(), env...)
	ec.Stderr, ec.Stdout = os.Stderr, os.Stdout

	return ec.Run()
}
