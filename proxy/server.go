package proxy

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
)

type Server struct {
	Command  Command
	Network  string
	Address  string
	listener net.Listener
}

func DefaultCommand(args ...string) *exec.Cmd {
	return exec.Command("aws-vault", args...)
}

type Command func(args ...string) *exec.Cmd

func NewServer(network string, address string) *Server {
	return &Server{
		Command: DefaultCommand,
		Network: network,
		Address: address,
	}
}

func (s *Server) Run() error {
	if s.listener != nil {
		return nil
	}

	l, err := net.Listen(s.Network, s.Address)
	if err != nil {
		return err
	}
	s.listener = l

	log.Printf("Listening at %s\n", s.Address)

	go s.handleConnections()

	return nil
}

func (s *Server) Close() error {
	if s.listener == nil {
		return nil
	}

	err := s.listener.Close()
	s.listener = nil
	return err
}

func (s *Server) handleConnections() {
	for {
		if s.listener == nil {
			return
		}

		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		go s.handleConnection(conn)
	}
}

func (s Server) handleConnection(conn net.Conn) {
	if err := s.handleRequest(conn, conn); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	conn.Close()
}

func (s Server) handleRequest(w io.Writer, r io.Reader) error {
	decoder := gob.NewDecoder(r)

	var args Args
	if err := decoder.Decode(&args); err != nil {
		return err
	}

	cmd := s.Command(args...)
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

	var error string
	if err = cmd.Run(); err != nil {
		error = err.Error()
	}

	// TODO - determine which errors should go back to the client
	encoder := gob.NewEncoder(w)
	var msg interface{} = Exit{
		ExitCode: cmd.ProcessState.ExitCode(),
		Error:    error,
	}
	err = encoder.Encode(&msg)

	return err
}
