// Package config validates explicit configuration without exposing connection secrets.
package config

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{DatabaseURL: strings.TrimSpace(getenv("DANBER_DATABASE_URL")), HTTPAddr: strings.TrimSpace(getenv("DANBER_HTTP_ADDR"))}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = "127.0.0.1:8090"
	}
	u, err := url.Parse(cfg.DatabaseURL)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" {
		return Config{}, errors.New("DANBER_DATABASE_URL debe ser una URL PostgreSQL válida")
	}
	if u.User == nil || u.User.Username() == "" {
		return Config{}, errors.New("la conexión requiere un usuario SQL de ejecución")
	}
	switch strings.ToLower(u.User.Username()) {
	case "root", "admin", "node":
		return Config{}, errors.New("usa un usuario SQL de ejecución, distinto del administrador")
	}
	if u.Path == "" || u.Path == "/" || u.Path == "/defaultdb" || u.Path == "/system" {
		return Config{}, errors.New("indica explícitamente la base de la aplicación")
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil || u.Fragment != "" {
		return Config{}, errors.New("opciones de conexión inválidas")
	}
	for key, values := range query {
		if len(values) != 1 {
			return Config{}, errors.New("no se permiten opciones de conexión repetidas")
		}
		switch key {
		case "sslmode", "sslrootcert", "sslcert", "sslkey":
		default:
			return Config{}, errors.New("la URL solo admite opciones TLS; usuario, host y base deben ser explícitos")
		}
	}
	// This increment is local-only. Never infer trust from a host suffix.
	if !loopback(u.Hostname()) {
		return Config{}, errors.New("esta entrega admite únicamente CockroachDB local")
	}
	modes := query["sslmode"]
	if len(modes) != 1 || (modes[0] != "disable" && modes[0] != "verify-full") {
		return Config{}, errors.New("define sslmode=disable para desarrollo local o verify-full")
	}
	host, port, err := net.SplitHostPort(cfg.HTTPAddr)
	n, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || n < 1 || n > 65535 || !loopback(host) {
		return Config{}, errors.New("DANBER_HTTP_ADDR requiere localhost o IP loopback y puerto válido")
	}
	return cfg, nil
}

func loopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
