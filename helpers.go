// summit backend/helpers.go - helper functions

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/msteinert/pam"
)

type (
	LogWriter struct{}
	Number    interface {
		~int | ~uint | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64
	}
)

// authenticates user with pam
func H_PamAuth(serviceName, userName, passwd string) error {
	t, err := pam.StartFunc(serviceName, userName, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return passwd, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", fmt.Errorf("unrecognized pam message style")
	})

	if err != nil {
		return err
	}
	if err = t.Authenticate(0); err != nil {
		return err
	}

	return nil
}

// generates random b64 str
func H_RandomBase64(length int) (string, error) {
	numBytes := (length * 3) / 4
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)[:length], nil
}

// executes a command and prints output if failed. returns complete output.
func H_Execute(args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no command")
	}

	// create the command
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	strout := string(out)

	if err != nil {
		lines := strings.Split(strings.TrimSpace(strout), "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]
			return strout, fmt.Errorf("%v: command execution failed: %v - last line: %v", strings.Join(args, " "), err, lastLine)
		}
		return strout, fmt.Errorf("command execution failed: %v", err)
	}

	return strout, nil
}

// logging format
func (lw *LogWriter) Write(bs []byte) (int, error) {
	if strings.Contains(string(bs), ": remote error: tls: unknown certificate") || strings.Contains(string(bs), "websocket: close 1001 (going away)") {
		return 0, nil
	}

	return fmt.Printf("[%s] %s", time.Now().Format("2006-01-02 15:04:05Z0700"), string(bs))
}

// gets user ip
func H_ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

/* "Go is such a nice language!"
 * Go:
 *
 * At this point, I REALLY hope that I'm just dumb and that there's a better way to do this.
 * There has to be a better way, right? RIGHT?
 */
func H_Cast[T Number](v any) (T, error) {
	var zero T
	switch n := v.(type) {
	case int, int8, int16, int32, int64:
		return T(reflect.ValueOf(n).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return T(reflect.ValueOf(n).Uint()), nil
	case float32, float64:
		return T(reflect.ValueOf(n).Float()), nil
	default:
		return zero, fmt.Errorf("not numeric: %T", v)
	}
}

// prints "{s}: {err}" to stdout, and gives HTTP 500 to w
func H_ISE(w http.ResponseWriter, s string, err error) {
	log.Println(s, err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// copies a file
func H_CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}
