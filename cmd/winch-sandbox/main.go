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
	child := newChildOwner(shippedFakeHarness(), "-run-id", "sandbox")
	if err := child.Start(); err != nil {
		log.Fatal("fake harness failed to start")
	}
	handler, err := attach.New(config.staticDir, attach.Session{
		Posture:           config.posture,
		ChildStatus:       attach.ChildDown,
		ChildStatusSource: child.Status,
	})
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
		child.Stop(shutdown)
		_ = server.Shutdown(shutdown)
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("sandbox server failed")
	}
}
