package server

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"os/exec"

	"github.com/ozbe/aws-vault-proxy/internal/protocol"
)

type Config struct {
	Command string
	Network string
	Address string
}

type Server struct {
	Config
}

func New(conf Config) Server {
	if conf.Command == "" {
		conf.Command = "aws-vault"
	}
	return Server{conf}
}

// TODO - support deadline and close
func (s Server) Listen() error {
	l, err := net.Listen(s.Network, s.Address)
	if err != nil {
		return err
	}
	defer l.Close()

	log.Printf("Listening at %s\n", s.Address)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s Server) handleConnection(conn net.Conn) {
	if err := s.handleRequest(conn, conn); err != nil {
		log.Println(err)
	}
	conn.Close()
}

func (s Server) handleRequest(w io.Writer, r io.Reader) error {
	decoder := gob.NewDecoder(r)

	var req protocol.Cmd
	if err := decoder.Decode(&req); err != nil {
		return err
	}

	cmd := exec.Command(s.Command, req.Args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdinWriter := protocol.NewStdinWriter(stdin)
	go func(w io.Writer, r io.Reader) {
		io.Copy(w, r)
	}(stdinWriter, r)

	cmd.Stdout = protocol.NewStdoutWriter(w)
	cmd.Stderr = protocol.NewStderrWriter(w)

	err = cmd.Run()

	// TODO - determine which errors should go back to the client
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(&protocol.Exit{
		ExitCode: cmd.ProcessState.ExitCode(),
		Error:    nil,
	})

	return err
}
