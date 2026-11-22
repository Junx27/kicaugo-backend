package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestShutdownOnSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	httpServer := &http.Server{}
	done := make(chan struct{})

	go func() {
		shutdownOnSignal(ctx, httpServer)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdownOnSignal() tidak selesai setelah konteks dibatalkan")
	}
}
