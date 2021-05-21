package main

import (
	"io"
	"regexp"
	"unicode"
	"unicode/utf8"
)

var envVarRegExp *regexp.Regexp = regexp.MustCompile(`(?m)^\w+\=\w+$`)

type FilterAwsEnvWriter struct {
	inner     io.Writer
	env       []string
	remainder []byte
}

func NewFilterAwsEnvWriter(w io.Writer) FilterAwsEnvWriter {
	return FilterAwsEnvWriter{
		inner: w,
	}
}

func (w *FilterAwsEnvWriter) Write(p []byte) (int, error) {
	w.remainder = append(w.remainder, p...)
	return len(p), w.filter(false)
}

func (w *FilterAwsEnvWriter) Close() error {
	return w.filter(true)
}

func (w *FilterAwsEnvWriter) filter(atEOF bool) error {
	for {
		a, t, err := scanWordsWithLeadingAndOneTrailingSpace(w.remainder, atEOF)
		if err != nil || t == nil {
			return err
		}

		w.remainder = w.remainder[a:]
		if awsEnvVarRegExp.Match(t) {
			w.env = append(w.env, string(t))
		} else {
			_, err = w.inner.Write(t)

			if err != nil {
				return err
			}
		}
	}
}

func (w FilterAwsEnvWriter) Env() []string {
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
