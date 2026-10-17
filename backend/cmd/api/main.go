package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/GersonLujan97/danber/backend/internal/bootstrap"
	"github.com/GersonLujan97/danber/backend/internal/platform/config"
)

func main() { os.Exit(run()) }

func run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		logger.Error("configuracion_invalida", "motivo", err.Error())
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := bootstrap.Run(ctx, cfg, logger); err != nil {
		// Database adapter errors are intentionally generic and never include DSNs.
		logger.Error("inicio_o_ejecucion_fallidos", "motivo", err.Error())
		return 1
	}
	return 0
}
