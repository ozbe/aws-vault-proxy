package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
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
var awsVaultMfaPromptRegExp *regexp.Regexp

func init() {
	awsEnvVarRegExp = regexp.MustCompile(`(?m)^AWS_.*`)
	awsVaultMfaPromptRegExp = regexp.MustCompile(`^Enter token for arn:aws:iam::\d+:mfa/.+:`)
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
	cmd := exec.Command(p.command, "exec", *profile, "--", "env")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	buf := bufio.NewReader(stdout)
	var line []byte
	for {
		segment, err := buf.ReadString(' ')
		if err != nil && err != io.EOF {
			break
		}

		line = append(line, segment...)
		line = bytes.Trim(line, "\n")

		if awsVaultMfaPromptRegExp.Match([]byte(line)) {
			mfa, err := promptForMFA(string(line))
			if err != nil {
				return err
			}
			line = nil
			go func(mfa string) {
				_, err = stdin.Write([]byte(mfa + "\n"))
			}(mfa)
		} else if awsEnvVarRegExp.Match([]byte(line)) {
			*env = append(*env, string(line))
		}

		if err == io.EOF {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func promptForMFA(prompt string) (string, error) {
	var mfa string
	fmt.Println(prompt)
	_, err := fmt.Scanln(&mfa)
	return mfa, err
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
