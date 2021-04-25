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
	network = "tcp"
	// TODO - Allow overriding this with env variable
	port = "7654"
	// TODO - Allow overriding this with env variable
	host = "host.docker.internal"
)

var awsEnvVarRegExp *regexp.Regexp

func init() {
	awsEnvVarRegExp = regexp.MustCompile(`(?m)^AWS_.*`)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing command.")
	}

	var err error
	switch os.Args[1] {
	case "server":
		err = serve()
	case "exec":
		err = call(os.Args[2:])
	default:
		log.Fatal("Unknown command.")
	}

	if err != nil {
		log.Fatal(err)
	}
}

type Proxy struct {
}

func (p *Proxy) Env(profile *string, env *[]string) error {
	// FIXME - handle MFA (`Enter token for arn:aws:iam::ACCOUNTID:mfa/USER:`)`
	output, err := exec.Command("aws-vault", "exec", *profile, "--", "env").Output()
	if err != nil {
		return err
	}

	*env = awsEnvVarRegExp.FindAllString(string(output), -1)
	return nil
}

func serve() error {
	if err := os.RemoveAll(host); err != nil {
		return err
	}

	proxy := new(Proxy)
	err := rpc.Register(proxy)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	l, err := net.Listen(network, fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}
	defer l.Close()

	log.Printf("Listening at %s\n", port)

	return http.Serve(l, nil)
}

func call(args []string) error {
	if len(args) < 3 || args[1] != "--" {
		return errors.New("invalid arguments length or format")
	}

	profile := args[0]

	client, err := rpc.DialHTTP(network, fmt.Sprintf("%s:%s", host, port))
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
