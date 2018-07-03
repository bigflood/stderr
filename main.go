package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	flags := Flags{}

	flag.StringVar(&flags.FilePath, "f", "", "file path")
	flag.BoolVar(&flags.MuteErr, "m", false, "mute stderr")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s [ options ] cmd args... \n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		os.Exit(-1)
	}

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	errReader, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-3)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-2)
	}

	defer errReader.Close()
	if err := processStderr(flags, errReader); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-4)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-5)
	}
}

type Flags struct {
	FilePath string
	MuteErr  bool
}

func processStderr(flags Flags, input io.Reader) error {
	if !flags.MuteErr {
		input = io.TeeReader(input, os.Stderr)
	}

	if flags.FilePath == "-" {
		input = io.TeeReader(input, os.Stdout)
	} else if flags.FilePath != "" {
		f, err := os.Create(flags.FilePath)
		if err != nil {
			return err
		}

		defer f.Close()

		input = io.TeeReader(input, f)
	}

	if _, err := io.Copy(nullWriter{}, input); err != nil {
		return err
	}

	return nil
}

type nullWriter struct{}

func (nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
