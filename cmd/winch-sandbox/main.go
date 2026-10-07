package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaielc/code-winch/internal/adapters/transport/attach"
)

func main() {
	log.SetFlags(0)
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	handler, err := attach.New(config.staticDir, config.posture)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: config.addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("sandbox server failed")
	}
}
