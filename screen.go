package main

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/kr/pty"
)

// Screen is a recorder screen.
type Screen interface {
	// Start starts and attaches shell to the terminal.
	Start()
	// Stop stopsscreen and detaches shell from the terminal.
	Stop()
	// Wait waits for the screen process to terminate.
	Wait()
	// OnInput calls f when there is an input.
	OnInput(f func(ascii byte))

	io.Writer
}

type screen struct {
	cmd *exec.Cmd
	f   *os.File
	io.Reader
	r Recorder

	actions  []func(byte)
	stopChan chan struct{}
}

// NewScreen creates a new screen.
func NewScreen(shell string) (Screen, error) {
	cmd := exec.Command(shell)

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &screen{
		cmd:    cmd,
		f:      f,
		Reader: os.Stdin,
		r:      &recorder{},
	}, nil
}

func (s *screen) Start() {
	s.stopChan = make(chan struct{})
	go handleResize(s)
	go handleStdout(s)
	go handleStdin(s)
	go handleExit(s)
}

func (s *screen) Stop() {
	s.cmd.Process.Kill()
}

func (s *screen) Wait() {
	<-s.stopChan
}

func (s *screen) Read(b []byte) (int, error) {
	for _, f := range s.actions {
		f(b[0])
	}
	return s.Reader.Read(b)
}

func (s *screen) Write(b []byte) (int, error) {
	return s.f.Write(b)
}

func (s *screen) OnInput(f func(ascii byte)) {
	s.actions = append(s.actions, f)
}

func handleStdin(s *screen) {
	io.Copy(s.f, s)
}

func handleStdout(s *screen) {
	io.Copy(os.Stdout, s.f)
}

func handleExit(s *screen) {
	s.cmd.Wait()
	close(s.stopChan)
}

func handleResize(s *screen) {
	resize := func() {
		InheritSize(os.Stdin, s.f)
	}

	// perform an immediate resize
	resize()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-c:
				resize()
			case <-s.stopChan:
				return
			}
		}
	}()
}
