package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/msteinert/pam"
)

// authenticates user with pam
func pamAuth(serviceName, userName, passwd string) error {
	t, err := pam.StartFunc(serviceName, userName, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return passwd, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", errors.New("unrecognized pam message style")
	})

	if err != nil {
		return err
	}
	if err = t.Authenticate(0); err != nil {
		return err
	}

	return nil
}

// human readable byte sizes
func humanReadable(bytes uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%dg", int(math.Round(float64(bytes)/float64(gb))))
	case bytes >= mb:
		return fmt.Sprintf("%dm", int(math.Round(float64(bytes)/float64(mb))))
	case bytes >= kb:
		return fmt.Sprintf("%dk", int(math.Round(float64(bytes)/float64(kb))))
	default:
		return fmt.Sprintf("%d", bytes)
	}
}

// generate random b64 str
func randomBase64String(length int) (string, error) {
	numBytes := (length * 3) / 4
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)[:length], nil
}

func execCmd(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	// Extract the command name and arguments
	cmdName := args[0]
	cmdArgs := args[1:]

	// Create the command with the provided arguments
	cmd := exec.Command(cmdName, cmdArgs...)

	// Run the command and capture the combined output
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	output := stdout.String()
	if output != "" {
		fmt.Println(output)
	}

	if err != nil {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]
			return fmt.Errorf("%v: command execution failed: %v - last line: %v", strings.Join(args, " "), err, lastLine)
		}
		return fmt.Errorf("command execution failed: %v", err)
	}

	return nil
}

type logWriter struct{}

// logging format
func (lw *logWriter) Write(bs []byte) (int, error) {
	if strings.Contains(string(bs), ": remote error: tls: unknown certificate") || strings.Contains(string(bs), "websocket: close 1001 (going away)") {
		return fmt.Print()
	}
	return fmt.Printf("[%s] %s", time.Now().Format(time.RFC1123), string(bs))
}
