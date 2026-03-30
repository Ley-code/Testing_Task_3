package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer bundles a running Postgres and its connection string (postgres://...).
type PostgresContainer struct {
	Container testcontainers.Container
	DSN       string
}

// StartPostgres starts postgres:15-alpine and returns DSN for user test, db orders_test.
func StartPostgres(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "orders_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	host, err := c.Host(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	port, err := c.MappedPort(ctx, nat.Port("5432/tcp"))
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/orders_test?sslmode=disable", host, port.Port())
	return &PostgresContainer{Container: c, DSN: dsn}, nil
}

// WireMockContainer exposes the mock HTTP base URL (e.g. http://host:port).
type WireMockContainer struct {
	Container testcontainers.Container
	BaseURL   string
}

// StartWireMock starts wiremock/wiremock and returns base URL for HTTP client.
func StartWireMock(ctx context.Context) (*WireMockContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "wiremock/wiremock:3.3.1",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	host, err := c.Host(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	port, err := c.MappedPort(ctx, nat.Port("8080/tcp"))
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	base := fmt.Sprintf("http://%s:%s", host, port.Port())
	return &WireMockContainer{Container: c, BaseURL: base}, nil
}
