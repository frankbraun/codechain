// Package terminal provides utility function to read from terminals.
package terminal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/frankbraun/codechain/util/bzero"
	"golang.org/x/crypto/ssh/terminal"
)

// ReadPassphrase reads a single line from fd without local echo and returns
// it (without trailing newline). When confirm is true it reads a second line
// and makes sure both passphrases match.
func ReadPassphrase(fd int, confirm bool) ([]byte, error) {
	var (
		pass   []byte
		pass2  []byte
		reader *bufio.Reader
		c      chan os.Signal
		stop   chan bool
		err    error
	)
	isTerminal := terminal.IsTerminal(fd)
	fmt.Printf("passphrase: ")
	if isTerminal {
		// Get terminal state to restore in case of interrupt.
		state, err := terminal.GetState(fd)
		if err != nil {
			return nil, err
		}
		// Create the necessary channels.
		c = make(chan os.Signal, 1)
		stop = make(chan bool, 1)
		// Register signal handler.
		signal.Notify(c, os.Interrupt)
		// Spawn goroutine to handle signal.
		go func() {
			select {
			case <-c:
				// Restore terminal and close goroutine.
				terminal.Restore(fd, state)
				fmt.Fprintln(os.Stderr, "cancelled")
				return
			case <-stop:
				return
			}
		}()
	}
	if isTerminal {
		pass, err = terminal.ReadPassword(fd)
		fmt.Println("")
		// Deregister signal handler.
		signal.Stop(c)
		// Stop signal handler goroutine to prevent goroutine leak.
		stop <- true
	} else {
		reader = bufio.NewReader(os.NewFile(uintptr(fd), "terminal"))
		pass, err = reader.ReadBytes('\n')
	}
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("unable to read passphrase")
		}
		return nil, err
	}
	if len(pass) == 0 {
		return nil, errors.New("please provide a passphrase")
	}
	pass = bytes.TrimRight(pass, "\n")
	if confirm {
		fmt.Printf("confirm passphrase: ")
		if isTerminal {
			pass2, err = terminal.ReadPassword(syscall.Stdin)
			fmt.Println("")
		} else {
			pass2, err = reader.ReadBytes('\n')
		}
		if err != nil {
			return nil, err
		}
		defer bzero.Bytes(pass2)
		pass2 = bytes.TrimRight(pass2, "\n")
		if !bytes.Equal(pass, pass2) {
			return nil, errors.New("passphrases don't match")
		}
	}
	return pass, nil
}

// ReadLine reads a single line from r it and returns it (without trailing
// newline).
func ReadLine(r io.Reader) ([]byte, error) {
	str, err := bufio.NewReader(r).ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("unable to read line")
		}
		return nil, err
	}
	return bytes.TrimSpace(str), nil
}

// Confirm asks the user to confirm the question with yes or no.
func Confirm(question string) error {
	for {
		fmt.Print(question + " [y/n]: ")
		answer, err := ReadLine(os.Stdin)
		if err != nil {
			return err
		}
		a := string(bytes.ToLower(answer))
		if strings.HasPrefix(a, "y") {
			return nil
		} else if strings.HasPrefix(a, "n") {
			return ErrAbort
		} else {
			fmt.Println("answer not recognized")
		}
	}
}
