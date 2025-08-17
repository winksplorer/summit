// summit backend/helpers.go - helper functions

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/msteinert/pam"
)

// used for humanReadable functions
const (
	kib = 1024
	mib = kib * 1024
	gib = mib * 1024
	tib = gib * 1024
)

type (
	logWriter struct{}
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

// human readable byte sizes, split unit and value
func H_HumanReadableSplit(bytes uint64) (float64, string) {
	switch {
	case bytes >= tib:
		return float64(bytes) / float64(tib), "t"
	case bytes >= gib:
		return float64(bytes) / float64(gib), "g"
	case bytes >= mib:
		return float64(bytes) / float64(mib), "m"
	case bytes >= kib:
		return float64(bytes) / float64(kib), "k"
	default:
		return float64(bytes), "b"
	}
}

// human readable byte sizes, combined string
func H_HumanReadable(bytes uint64) string {
	value, unit := H_HumanReadableSplit(bytes)
	return fmt.Sprintf("%d%s", int(math.Round(value)), unit)
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
func (lw *logWriter) Write(bs []byte) (int, error) {
	if strings.Contains(string(bs), ": remote error: tls: unknown certificate") || strings.Contains(string(bs), "websocket: close 1001 (going away)") {
		return 0, nil
	}

	return fmt.Printf("[%s] %s", time.Now().Format("02 Jan 2006 15:04:05 MST"), string(bs))
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
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return T(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return T(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return T(rv.Float()), nil
	default:
		return zero, fmt.Errorf("not numeric: %T", v)
	}
}

// prints "{s}: {err}" to stdout, and gives HTTP 500 to w
func H_ISE(w http.ResponseWriter, s string, err error) {
	log.Println(s, err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// sets a value in m to val based on key. basically, key="x.y.z" will set m["x"]["y"]["z"] to val.
func H_SetValue(m map[string]interface{}, key string, val interface{}) error {
	var interf interface{} = m
	for i, k := range strings.Split(key, ".") {
		nested, ok := interf.(map[string]interface{})
		if !ok {
			return fmt.Errorf("not a map at %s", key)
		}

		if i == len(strings.Split(key, "."))-1 {
			nested[k] = val
			return nil
		}

		if _, ok := nested[k]; !ok {
			nested[k] = make(map[string]interface{})
		}
		interf = nested[k]
	}

	return nil
}

// returns a value in m based on key. basically, key="x.y.z" will return m["x"]["y"]["z"].
func H_GetRawValue(m map[string]interface{}, key string) (interface{}, error) {
	var interf interface{} = m
	for _, k := range strings.Split(key, ".") {
		nested, ok := interf.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("not a map at %q", k)
		}
		interf, ok = nested[k]
		if !ok {
			return nil, fmt.Errorf("couldn't find %q", key)
		}
	}

	return interf, nil
}

// returns T in m based on key, using value.(T). basically, key="x.y.z" will return m["x"]["y"]["z"].
func H_GetValue[T any](m map[string]interface{}, key string) (T, error) {
	var zero T

	interf, err := H_GetRawValue(m, key)
	if err != nil {
		return zero, err
	}

	val, ok := interf.(T)
	if !ok {
		return zero, fmt.Errorf("value at %q is not %T, but instead %T", key, zero, interf)
	}
	return val, nil
}

// returns T in m based on key, using H_Cast[T]. basically, key="x.y.z" will return m["x"]["y"]["z"].
func H_GetNumericalValue[T Number](m map[string]interface{}, key string) (T, error) {
	var zero T

	interf, err := H_GetRawValue(m, key)
	if err != nil {
		return zero, err
	}

	casted, err := H_Cast[T](interf)
	if err != nil {
		return zero, err
	}

	return casted, nil
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
