package main

import (
	"log"
	"net/http"
	"os"
)

const portEnvVar = "PORT"
const defaultPort = "8080"
const dataDir = "data"

func main() {
	port := os.Getenv(portEnvVar)
	if port == "" {
		port = defaultPort
		log.Printf("warning: %s not specified; using default %s", portEnvVar, port)
	}

	addr := ":" + port
	log.Printf("listen addr %s (http://localhost:%s/); data dir=%s", addr, port, dataDir)

	handler := http.FileServer(http.Dir(dataDir))
	if err := http.ListenAndServe(addr, handler); err != nil {
		panic(err)
	}
}
