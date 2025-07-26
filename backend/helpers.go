// summit backend/helpers.go - helper functions

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/msteinert/pam"
)

// used for humanReadable functions
const (
	kb = 1024
	mb = kb * 1024
	gb = mb * 1024
	tb = gb * 1024
)

type logWriter struct{}

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

// human readable byte sizes, split unit and value
func humanReadableSplit(bytes uint64) (float64, string) {
	switch {
	case bytes >= tb:
		return float64(bytes) / float64(tb), "t"
	case bytes >= gb:
		return float64(bytes) / float64(gb), "g"
	case bytes >= mb:
		return float64(bytes) / float64(mb), "m"
	case bytes >= kb:
		return float64(bytes) / float64(kb), "k"
	default:
		return float64(bytes), "b"
	}
}

// human readable byte sizes, combined string
func humanReadable(bytes uint64) string {
	value, unit := humanReadableSplit(bytes)
	return fmt.Sprintf("%d%s", int(math.Round(value)), unit)
}

// generates random b64 str
func randomBase64String(length int) (string, error) {
	numBytes := (length * 3) / 4
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)[:length], nil
}

// executes a command and prints output
func execCmd(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	// extract the command name and arguments
	cmdName := args[0]
	cmdArgs := args[1:]

	// create the command
	cmd := exec.Command(cmdName, cmdArgs...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	// run the command and capture the combined output
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

// logging format
func (lw *logWriter) Write(bs []byte) (int, error) {
	if strings.Contains(string(bs), ": remote error: tls: unknown certificate") || strings.Contains(string(bs), "websocket: close 1001 (going away)") {
		return fmt.Print()
	}
	return fmt.Printf("[%s] %s", time.Now().Format(time.RFC1123), string(bs))
}

// gets user ip
func clientIP(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}

// why god
func asUint16(v any) uint16 {
	if u, ok := v.(uint8); ok {
		return uint16(u)
	}
	if u, ok := v.(int8); ok {
		return uint16(u)
	}
	return 0
}

// allows for slight differences if Lighthouse is involved
func userAgentMatches(storedUA, currentUA string) bool {
	if storedUA == currentUA {
		return true
	}
	if strings.Contains(currentUA, "Chrome-Lighthouse") {
		return extractAppleWebKitVersion(storedUA) == extractAppleWebKitVersion(currentUA)
	}
	return false
}

// returns the AppleWebKit version from a UA string
func extractAppleWebKitVersion(ua string) string {
	re := regexp.MustCompile(`AppleWebKit/([\d\.]+)`)
	match := re.FindStringSubmatch(ua)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// prints "{s}: {err}" to stdout, and gives HTTP 500 to w
func ise(w http.ResponseWriter, s string, err error) {
	log.Println(s, err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// returns a nested value
func getNested(m map[string]interface{}, keys ...string) (interface{}, error) {
	var interf interface{} = m
	for _, key := range keys {
		nested, ok := interf.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("couldn't find %s", strings.Join(keys, "."))
		}

		interf, ok = nested[key]
		if !ok {
			return nil, fmt.Errorf("couldn't find %s", strings.Join(keys, "."))
		}
	}
	return interf, nil
}
