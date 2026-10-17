// Package bootstrap composes infrastructure and HTTP without defining business rules.
package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/health"
	"github.com/GersonLujan97/danber/backend/internal/platform/config"
	"github.com/GersonLujan97/danber/backend/internal/platform/database"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	db, err := database.Open(connectCtx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return err
	}
	defer db.Close()
	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return errors.New("no se pudo abrir el puerto HTTP configurado")
	}
	server := &http.Server{
		Handler:           health.Handler(db.Check),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	logger.Info("api_iniciada", "direccion", listener.Addr().String())
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.New("el servidor HTTP se detuvo inesperadamente")
	case <-ctx.Done():
		// A fresh context lets in-flight requests finish after the parent is cancelled.
		shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return errors.New("no se completó el cierre ordenado")
		}
		err := <-result
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return errors.New("falló el cierre del servidor HTTP")
		}
		logger.Info("api_detenida")
		return nil
	}
}
