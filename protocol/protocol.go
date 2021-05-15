package protocol

import (
	"encoding/gob"
	"io"
)

/*
The gob encoder is needed across _all_ methods
maybe make a struct to wrap this all
*/

func init() {
	gob.Register(Cmd{})
	gob.Register(Stdin{})
	gob.Register(Stdout{})
	gob.Register(Stderr{})
	gob.Register(Exit{})
}

const (
	cmdMsgType MsgType = iota
	stdinMsgType
	stdoutMsgType
	stderrMsgType
	exitMsgType
)

type MsgType = int

type Msg interface {
	Type() MsgType
}

type Cmd struct {
	Args []string
}

var _ Msg = Cmd{}

func (_ Cmd) Type() MsgType {
	return cmdMsgType
}

type Stdin struct {
	Data []byte
}

var _ Msg = Cmd{}

func (_ Stdin) Type() MsgType {
	return stdinMsgType
}

type Stdout struct {
	Data []byte
}

var _ Msg = Stdout{}

func (_ Stdout) Type() MsgType {
	return stdoutMsgType
}

type Stderr struct {
	Data []byte
}

var _ Msg = Stderr{}

func (_ Stderr) Type() MsgType {
	return stderrMsgType
}

type Exit struct {
	ExitCode int
	Error    error
}

var _ Msg = Exit{}

func (_ Exit) Type() MsgType {
	return exitMsgType
}

type Writer struct {
	encoder *gob.Encoder
	wrap    func([]byte) Msg
}

func NewStdinWriter(w io.Writer) Writer {
	return newWriter(w, func(p []byte) Msg {
		return Stdin{p}
	})
}

func NewStdoutWriter(w io.Writer) Writer {
	return newWriter(w, func(p []byte) Msg {
		return Stdout{p}
	})
}

func NewStderrWriter(w io.Writer) Writer {
	return newWriter(w, func(p []byte) Msg {
		return Stderr{p}
	})
}

func newWriter(w io.Writer, wrap func([]byte) Msg) Writer {
	return Writer{
		encoder: gob.NewEncoder(w),
		wrap:    wrap,
	}
}

func (w Writer) Write(p []byte) (int, error) {
	msg := w.wrap(p)
	err := w.encoder.Encode(&msg)

	if err != nil {
		return 0, err
	}

	return len(p), err
}
