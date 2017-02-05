package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Recorder is command recorder.
type Recorder interface {
	// Input saves an input.
	Input(key byte)
	// Play replays the recorded inputs by writing them to the writer.
	Play(io.Writer) error
	// Save saves the session.
	Save(io.Writer) error
	// Load loads a previously saved session.
	Load(io.Reader) error
}

// NewRecorder returns a new Recorder. Ready for use.
func NewRecorder() Recorder {
	return &recorder{}
}

type inputEvent struct {
	key   byte
	delay time.Duration
}

type recorder struct {
	last   time.Time
	inputs []inputEvent
}

func (r *recorder) Input(key byte) {
	var ev inputEvent
	ev.key = key
	if !r.last.IsZero() {
		ev.delay = time.Since(r.last)
	}
	r.last = time.Now()
	r.inputs = append(r.inputs, ev)
}

func (r *recorder) Play(w io.Writer) error {
	for _, ev := range r.inputs {
		time.Sleep(ev.delay)
		_, err := w.Write([]byte{ev.key})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *recorder) Save(w io.Writer) error {
	for _, ev := range r.inputs {
		_, err := fmt.Fprintf(w, "%v %v\n", ev.key, int64(ev.delay))
		if err != nil {
			return err
		}
	}
	return nil
}

var errDecoding = errors.New("Error decoding")

func (r *recorder) Load(rd io.Reader) error {
	input := bufio.NewScanner(rd)
	for input.Scan() {
		str := strings.Fields(input.Text())
		if len(str) < 2 {
			return errDecoding
		}
		var err error
		var ev inputEvent
		n, err := strconv.ParseInt(str[0], 10, 8)
		if err != nil {
			return errDecoding
		}
		ev.key = byte(n)
		d, err := strconv.ParseInt(str[1], 10, 64)
		if err != nil {
			return errDecoding
		}
		ev.delay = time.Duration(d)
		r.inputs = append(r.inputs, ev)
	}
	return nil
}
