package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/ozbe/aws-vault-proxy/protocol"
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

func main() {
	c := newConfig()

	err := listen(c)

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

func listen(conf config) error {
	l, err := net.Listen(conf.network, fmt.Sprintf(":%s", conf.port))
	if err != nil {
		return err
	}
	defer l.Close()

	log.Printf("Listening at %s\n", defaultPort)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conf, conn)
	}
}

func handleConnection(conf config, conn net.Conn) {
	if err := handleRequest(conf, conn); err != nil {
		log.Println(err)
	}
	conn.Close()
}

func handleRequest(conf config, conn net.Conn) error {
	decoder := gob.NewDecoder(conn)

	var req protocol.Cmd
	if err := decoder.Decode(&req); err != nil {
		return err
	}

	cmd := exec.Command(conf.command, req.Args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdinWriter := protocol.NewStdinWriter(stdin)
	go func(w io.Writer, r io.Reader) {
		io.Copy(w, r)
	}(stdinWriter, conn)

	cmd.Stdout = protocol.NewStdoutWriter(conn)
	cmd.Stderr = protocol.NewStderrWriter(conn)

	err = cmd.Run()

	// TODO - determine which errors should go back to the client
	encoder := gob.NewEncoder(conn)
	var msg protocol.Msg = protocol.Exit{
		ExitCode: cmd.ProcessState.ExitCode(),
		Error:    nil,
	}
	err = encoder.Encode(&msg)

	return err
}
