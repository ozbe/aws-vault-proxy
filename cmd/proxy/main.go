package main

import (
	"bufio"
	"bytes"
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
	defaultPort    = "7654"
)

const (
	commandEnvKey = "AWS_VAULT_PROXY_COMMAND"
	portEnvKey    = "AWS_VAULT_PROXY_PORT"
)

type config struct {
	command string
	network string
	port    string
}

var awsEnvVarRegExp *regexp.Regexp
var awsVaultMfaPromptRegExp *regexp.Regexp

func init() {
	awsEnvVarRegExp = regexp.MustCompile(`(?m)^AWS_.*`)
	awsVaultMfaPromptRegExp = regexp.MustCompile(`^Enter token for arn:aws:iam::\d+:mfa/.+:`)
}

func main() {
	c := newConfig()

	err := serve(c)

	if err != nil {
		log.Fatal(err)
	}
}

func newConfig() config {
	command := defaultCommand
	if val, ok := os.LookupEnv(commandEnvKey); ok {
		command = val
	}

	port := defaultPort
	if val, ok := os.LookupEnv(portEnvKey); ok {
		port = val
	}

	return config{
		command: command,
		network: defaultNetwork,
		port:    port,
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
			_, err = stdin.Write([]byte(mfa + "\n"))
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
