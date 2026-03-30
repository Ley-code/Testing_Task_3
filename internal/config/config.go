package config

import "os"

// ListenAddr returns host:port for the HTTP server. Uses PORT if set (e.g. "8080" or ":8080"), else ":8080".
func ListenAddr() string {
	p := os.Getenv("PORT")
	if p == "" {
		return ":8080"
	}
	if p[0] == ':' {
		return p
	}
	return ":" + p
}
