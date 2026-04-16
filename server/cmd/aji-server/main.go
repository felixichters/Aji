// Command aji-server is the entry point for the Aji game server.
//
// v0 scaffold: boots an HTTP listener with a /healthz endpoint only.
// Game logic, WebSocket hub, and world wiring land in later steps.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	logger := log.New(os.Stdout, "aji-server ", log.LstdFlags|log.Lmsgprefix)
	logger.Printf("booting on %s", *addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	srv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
