package proxy

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"os/exec"
)

type Server struct {
	Command string
	Network string
	Address string
}

var DefaultCommand = "aws-vault"

func NewServer(network string, address string) Server {
	return Server{
		Command: DefaultCommand,
		Network: network,
		Address: address,
	}
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

	var args Args
	if err := decoder.Decode(&args); err != nil {
		return err
	}

	cmd := exec.Command(s.Command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdinWriter := NewStdinWriter(stdin)
	go func(w io.Writer, r io.Reader) {
		io.Copy(w, r)
	}(stdinWriter, r)

	cmd.Stdout = NewStdoutWriter(w)
	cmd.Stderr = NewStderrWriter(w)

	err = cmd.Run()

	// TODO - determine which errors should go back to the client
	encoder := gob.NewEncoder(w)
	var msg interface{} = Exit{
		ExitCode: cmd.ProcessState.ExitCode(),
		Error:    nil,
	}
	err = encoder.Encode(&msg)

	return err
}