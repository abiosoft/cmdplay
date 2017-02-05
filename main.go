package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

var file string
var rec bool
var shell string

func exitWithErr(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	flag.StringVar(&file, "f", file, "input file (if record mode) or output file")
	flag.BoolVar(&rec, "r", rec, "use cmdplay in record mode")
	flag.StringVar(&shell, "s", shell, "the shell session to use. Defaults to $SHELL")

	flag.Parse()
	if file == "" {
		flag.Usage()
		os.Exit(1)
	}
	if shell == "" {
		if shell = os.Getenv("SHELL"); shell == "" {
			fmt.Println("$SHELL not found, use -s flag to specify shell to use")
			flag.Usage()
			os.Exit(1)
		}
	}

	state, err := terminal.MakeRaw(0)
	if err != nil {
		exitWithErr(err)
	}
	defer terminal.Restore(0, state)

	if rec {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			exitWithErr(err)
		}
		if err := record(f); err != nil {
			exitWithErr(err)
		}
		fmt.Println("Session saved to", f.Name())
	} else {
		f, err := os.Open(file)
		if err != nil {
			exitWithErr(err)
		}
		if err := play(f); err != nil {
			exitWithErr(err)
		}
		fmt.Println("Play complete")
	}

}

func record(f *os.File) error {
	fmt.Println("Recording started. Exit shell session to stop.")

	s, err := NewScreen(shell)
	if err != nil {
		return err
	}

	state, err := terminal.MakeRaw(0)
	if err != nil {
		return err
	}
	defer terminal.Restore(0, state)

	r := NewRecorder()
	s.OnInput(func(ascii byte) {
		r.Input(ascii)
	})

	s.Start()
	s.Wait()

	return r.Save(f)
}

func play(f *os.File) error {
	fmt.Println("Attempting to play", f.Name())

	s, err := NewScreen(shell)
	if err != nil {
		return err
	}

	state, err := terminal.MakeRaw(0)
	if err != nil {
		return err
	}
	defer terminal.Restore(0, state)

	r := NewRecorder()
	if err := r.Load(f); err != nil {
		return err
	}

	s.Start()
	if err := r.Play(s); err != nil {
		return err
	}
	s.Wait()
	return nil
}
