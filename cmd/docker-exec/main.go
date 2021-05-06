package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"unicode"
	"unicode/utf8"
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
	envVarRegExp = regexp.MustCompile(`(?m)^\w+\=\w+$`)
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
		err = execCmd(c, os.Args[2:])
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

func execCmd(conf config, args []string) error {
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

	// Copy user input
	go func() {
		io.Copy(conn, os.Stdin)
	}()

	// Filter output
	var env []string
	scanner := bufio.NewScanner(conn)
	scanner.Split(scanWordsWithLeadingAndOneTrailingSpace)

	for scanner.Scan() {
		if awsEnvVarRegExp.Match(scanner.Bytes()) {
			env = append(env, string(scanner.Bytes()))
		} else if !envVarRegExp.Match(scanner.Bytes()) {
			os.Stdout.Write(scanner.Bytes())
		}
	}

	// Run command
	ec := exec.Command(args[2], args[3:]...)
	ec.Env = append(os.Environ(), env...)
	ec.Stderr, ec.Stdout = os.Stderr, os.Stdout

	return ec.Run()
}

// Adapted from bufio.ScanWords
func scanWordsWithLeadingAndOneTrailingSpace(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Scan leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}
	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if unicode.IsSpace(r) {
			return i + width, data[0 : i+width], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[0:], nil
	}
	// Request more data.
	return start, nil, nil
}
