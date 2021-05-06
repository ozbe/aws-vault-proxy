package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"regexp"
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

var envVarRegExp *regexp.Regexp
var awsEnvVarRegExp *regexp.Regexp

func init() {
	// envVarRegExp = regexp.MustCompile(`(?m)^\w+\=\w+$`)
	awsEnvVarRegExp = regexp.MustCompile(`(?m)^AWS_\w+\=\w+$`)
}

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
		err = ec(c, os.Args[2:])
	default:
		log.Fatal("Unknown command.")
	}

	if err != nil {
		log.Fatal(err)
	}
}

type Cmd struct {
	Args []string
}

func ec(conf config, args []string) error {
	if len(args) < 3 || args[1] != "--" {
		return errors.New("invalid arguments length or format")
	}

	// PROFILE
	profile := args[0]

	// make tcp connection
	conn, err := net.Dial(conf.network, fmt.Sprintf("%s:%s", conf.host, conf.port))
	if err != nil {
		return err
	}

	// request `exec PROFILE -- env`
	req := Cmd{
		Args: []string{"exec", profile, "--", "env"},
	}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(req)

	go func() {
		io.Copy(conn, os.Stdin)
	}()

	var buf bytes.Buffer
	w := io.MultiWriter(
		os.Stdout,
		&buf,
	)

	io.Copy(w, conn)

	// Filter out ^AWS
	env := awsEnvVarRegExp.FindAllString(buf.String(), -1)

	// Run command
	ec := exec.Command(args[2], args[3:]...)
	ec.Env = append(os.Environ(), env...)
	ec.Stderr, ec.Stdout = os.Stderr, os.Stdout

	return ec.Run()
}

func execCmd(conf config, args []string) error {
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
