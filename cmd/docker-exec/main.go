package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"unicode"
	"unicode/utf8"

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
		err = runExec(c, os.Args[2:])
	default:
		log.Fatal("Unknown command.")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runExec(conf config, args []string) error {
	if len(args) < 3 || args[1] != "--" {
		return errors.New("invalid arguments length or format")
	}

	env, err := proxyExecEnv(conf, args[0])
	if err != nil {
		return err
	}

	aws := findAWS(env)

	return localExec(aws, args[2:])
}

func findAWS(env []string) []string {
	aws := []string{}

	for _, e := range env {
		if awsEnvVarRegExp.MatchString(e) {
			aws = append(aws, e)
		}
	}

	return aws
}

func proxyExecEnv(conf config, profile string) ([]string, error) {
	w := NewFilterEnvWriter(os.Stdout)

	p := proxy.NewClient(proxy.Config{
		Network: conf.network,
		Address: fmt.Sprintf("%s:%s", conf.host, conf.port),
	})
	pcmd := p.Cmd("exec", profile, "--", "env")
	pcmd.Stdin = os.Stdin
	pcmd.Stderr = os.Stderr
	pcmd.Stdout = &w

	if err := pcmd.Run(); err != nil {
		return nil, err
	}
	return w.Env(), nil
}

func localExec(env []string, args []string) error {
	ec := exec.Command(args[0], args[1:]...)
	ec.Env = append(os.Environ(), env...)
	ec.Stderr, ec.Stdout = os.Stderr, os.Stdout

	return ec.Run()
}

type FilterEnvWriter struct {
	inner     io.Writer
	env       []string
	remainder []byte
}

func NewFilterEnvWriter(w io.Writer) FilterEnvWriter {
	return FilterEnvWriter{
		inner: w,
	}
}

func (w *FilterEnvWriter) Write(p []byte) (int, error) {
	w.remainder = append(w.remainder, p...)
	return len(p), w.filter(false)
}

func (w *FilterEnvWriter) Close() error {
	return w.filter(true)
}

func (w *FilterEnvWriter) filter(atEOF bool) error {
	for {
		a, t, err := scanWordsWithLeadingAndOneTrailingSpace(w.remainder, atEOF)
		if err != nil || t == nil {
			return err
		}

		w.remainder = w.remainder[a:]
		if envVarRegExp.Match(t) {
			w.env = append(w.env, string(t))
		} else {
			_, err = w.inner.Write(t)

			if err != nil {
				return nil
			}
		}
	}
}

func (w FilterEnvWriter) Env() []string {
	return w.env
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
