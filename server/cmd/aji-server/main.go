// Command aji-server is the entry point for the Aji game server.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	ajinet "github.com/felixichters/Aji/server/internal/net"
	"github.com/felixichters/Aji/server/internal/world"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	boardW := flag.Int("board-w", 200, "board width")
	boardH := flag.Int("board-h", 200, "board height")
	radius := flag.Int("radius", 8, "engagement region radius")
	flag.Parse()

	logger := log.New(os.Stdout, "aji-server ", log.LstdFlags|log.Lmsgprefix)
	logger.Printf("booting on %s (board %dx%d, radius %d)", *addr, *boardW, *boardH, *radius)

	w := world.New(*boardW, *boardH, *radius)
	hub := ajinet.New(w, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("/ws", hub)

	srv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
