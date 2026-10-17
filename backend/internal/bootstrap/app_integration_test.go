//go:build integration

package bootstrap_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/bootstrap"
	"github.com/GersonLujan97/danber/backend/internal/platform/config"
)

func TestServeAndGracefulShutdown(t *testing.T) {
	url := os.Getenv("DANBER_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("isolated test URL is required")
	}
	slot, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := slot.Addr().String()
	slot.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- bootstrap.Run(ctx, config.Config{DatabaseURL: url, HTTPAddr: address}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()
	client := &http.Client{Timeout: time.Second}
	ready := false
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		response, requestErr := client.Get("http://" + address + "/health/ready")
		if requestErr == nil {
			response.Body.Close()
			if response.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server never became ready")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("shutdown exceeded deadline")
	}
}
