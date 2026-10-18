package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/user"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/identidad"
	"github.com/GersonLujan97/danber/backend/internal/platform/config"
	"github.com/GersonLujan97/danber/backend/internal/platform/database"
	"golang.org/x/term"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stderr, os.Getenv, term.IsTerminal(int(os.Stdin.Fd())), func() ([]byte, error) { return term.ReadPassword(int(os.Stdin.Fd())) }, provision); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type provisionFunc func(context.Context, string, []byte, string) error

func run(ctx context.Context, args []string, out io.Writer, env func(string) string, terminal bool, read func() ([]byte, error), create provisionFunc) error {
	flags := flag.NewFlagSet("crear-maestro", flag.ContinueOnError)
	flags.SetOutput(io.Discard) // Flag errors can include argument values: never echo them.
	username := flags.String("usuario", "", "nombre de acceso")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *username == "" {
		return errors.New("uso: go run ./cmd/crear-maestro --usuario <nombre>; la contraseña solo se ingresa en la terminal")
	}
	if !terminal {
		return errors.New("se requiere una terminal interactiva para ocultar la contraseña")
	}
	cfg, err := config.Load(env)
	if err != nil {
		return err
	}
	fmt.Fprint(out, "Contraseña (15–128 caracteres): ")
	secret, err := read()
	fmt.Fprintln(out)
	defer clear(secret)
	if err != nil {
		return errors.New("no se pudo leer la contraseña en la terminal")
	}
	fmt.Fprint(out, "Repetir contraseña: ")
	confirm, err := read()
	fmt.Fprintln(out)
	defer clear(confirm)
	if err != nil {
		return errors.New("no se pudo leer la confirmación")
	}
	if !bytes.Equal(secret, confirm) {
		return errors.New("las contraseñas no coinciden; no se creó ninguna cuenta")
	}
	if err := create(ctx, *username, secret, cfg.DatabaseURL); err != nil {
		return err
	}
	fmt.Fprintln(out, "Cuenta inicial creada con roles Maestro y Recepción. El inicio de sesión se implementará en la siguiente etapa.")
	return nil
}

func provision(ctx context.Context, username string, secret []byte, url string) error {
	operator, err := user.Current()
	if err != nil {
		return errors.New("no se pudo identificar al operador local")
	}
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	pool, err := database.Open(connectCtx, url)
	cancel()
	if err != nil {
		return err
	}
	defer pool.Close()
	operationCtx, cancelOperation := context.WithTimeout(ctx, 30*time.Second)
	defer cancelOperation()
	_, err = identidad.NewInitialAccount(pool).Create(operationCtx, username, secret, operator.Username)
	return err
}
