package main

import (
	"log"
	"os"
)

// creates tls certs if they don't exist
func TLS_Init() error {
	_, err := os.Stat("/etc/ssl/private/summit.key")
	_, err2 := os.Stat("/etc/ssl/certs/summit.crt")

	// if cert doesn't exist
	if os.IsNotExist(err) || os.IsNotExist(err2) {
		log.Println("TLS_Init: Creating TLS certificate.")

		// create cert
		if _, err := H_Execute(
			"openssl",
			"req",
			"-x509",
			"-nodes",
			"-days", "365",
			"-newkey", "rsa:2048",
			"-keyout", "/etc/ssl/private/summit.key",
			"-out", "/etc/ssl/certs/summit.crt",
			"-subj", ("/C=US/ST=Washington/O=winksplorer & contributors/CN=summit (" + G_Hostname + ")"),
		); err != nil {
			return err
		}

		// change permissions
		if err := os.Chmod("/etc/ssl/private/summit.key", 0700); err != nil {
			return err
		}
	} else {
		log.Println("TLS_Init: No work needed.")
	}

	return nil
}
