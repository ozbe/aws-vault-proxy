package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"regexp"
)

const (
	defaultCommand = "aws-vault"
	defaultNetwork = "tcp"
	defaultHost    = "host.docker.internal"
	defaultPort    = "7654"
)

const (
	commandEnvKey = "AWS_VAULT_PROXY_COMMAND"
	hostEnvKey    = "AWS_VAULT_PROXY_HOST"
	portEnvKey    = "AWS_VAULT_PROXY_PORT"
)

type config struct {
	command string
	network string
	host    string
	port    string
}

func newConfig() config {
	command := defaultCommand
	if val, ok := os.LookupEnv(commandEnvKey); ok {
		command = val
	}

	host := defaultHost
	if val, ok := os.LookupEnv(hostEnvKey); ok {
		host = val
	}

	port := defaultPort
	if val, ok := os.LookupEnv(portEnvKey); ok {
		port = val
	}

	return config{
		command: command,
		network: defaultNetwork,
		host:    host,
		port:    port,
	}
}

var awsEnvVarRegExp *regexp.Regexp

func init() {
	awsEnvVarRegExp = regexp.MustCompile(`(?m)^AWS_.*`)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing command.")
	}

	c := newConfig()

	var err error
	switch os.Args[1] {
	case "server":
		err = serve(c)
	case "exec":
		err = call(c, os.Args[2:])
	default:
		log.Fatal("Unknown command.")
	}

	if err != nil {
		log.Fatal(err)
	}
}

type Proxy struct {
	command string
}

func (p *Proxy) Env(profile *string, env *[]string) error {
	// FIXME - handle MFA (`Enter token for arn:aws:iam::ACCOUNTID:mfa/USER:`)`
	output, err := exec.Command(p.command, "exec", *profile, "--", "env").Output()

	if err != nil {
		return err
	}

	*env = awsEnvVarRegExp.FindAllString(string(output), -1)
	return nil
}

func serve(conf config) error {
	if err := os.RemoveAll(defaultHost); err != nil {
		return err
	}

	proxy := &Proxy{
		command: conf.command,
	}
	err := rpc.Register(proxy)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	l, err := net.Listen(conf.network, fmt.Sprintf(":%s", conf.port))
	if err != nil {
		return err
	}
	defer l.Close()

	log.Printf("Listening at %s\n", defaultPort)

	return http.Serve(l, nil)
}

func call(conf config, args []string) error {
	if len(args) < 3 || args[1] != "--" {
		return errors.New("invalid arguments length or format")
	}

	profile := args[0]

	client, err := rpc.DialHTTP(conf.network, fmt.Sprintf("%s:%s", conf.host, conf.port))
	if err != nil {
		return err
	}

	var env []string
	err = client.Call("Proxy.Env", &profile, &env)
	if err != nil {
		return err
	}

	ec := exec.Command(args[2], args[3:]...)
	ec.Env = append(os.Environ(), env...)
	ec.Stderr, ec.Stdout = os.Stderr, os.Stdout

	err = ec.Run()
	if err != nil {
		return err
	}

	return nil
}
