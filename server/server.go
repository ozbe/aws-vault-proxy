package server

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"

	"github.com/ozbe/aws-vault-proxy/protocol"
)

type Config struct {
	Command string
	Network string
	Port    string
}

func Listen(conf Config) error {
	l, err := net.Listen(conf.Network, fmt.Sprintf(":%s", conf.Port))
	if err != nil {
		return err
	}
	defer l.Close()

	log.Printf("Listening at %s\n", conf.Port)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conf, conn)
	}
}

func handleConnection(conf Config, conn net.Conn) {
	if err := handleRequest(conf, conn); err != nil {
		log.Println(err)
	}
	conn.Close()
}

func handleRequest(conf Config, conn net.Conn) error {
	decoder := gob.NewDecoder(conn)

	var req protocol.Cmd
	if err := decoder.Decode(&req); err != nil {
		return err
	}

	cmd := exec.Command(conf.Command, req.Args...)
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
	err = encoder.Encode(&protocol.Exit{
		ExitCode: cmd.ProcessState.ExitCode(),
		Error:    nil,
	})

	return err
}
