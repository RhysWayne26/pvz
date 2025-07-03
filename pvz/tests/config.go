//go:build integration || e2e

package tests

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"pvz-cli/infrastructure/db"
	"time"
)

const (

	// MigrationsDir specifies the relative path to the directory containing database migration files.
	MigrationsDir = "../../../migrations"

	// TruncateOrderSql defines the SQL query to truncate the `orders` table and reset its identity sequence.
	TruncateOrderSql = `TRUNCATE orders RESTART IDENTITY CASCADE;`

	// TruncateHistorySQL defines the SQL query to truncate the `order_history` table and reset its identity sequence.
	TruncateHistorySQL = `TRUNCATE order_history RESTART IDENTITY CASCADE;`
)

// NewCommonDeps sets up and returns common dependencies including a PostgreSQL container, database client, and context.
func NewCommonDeps(t provider.T) CommonDeps {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("listening on IPv4 address").
			WithPollInterval(1 * time.Second).
			WithStartupTimeout(10 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)
	dsn := fmt.Sprintf("postgres://testuser:password@localhost:%s/testdb?sslmode=disable", port.Port())
	client, err := db.NewDefaultPGXClient(dsn, dsn)
	require.NoError(t, err)
	sqlDB := stdlib.OpenDBFromPool(client.WritePool)
	err = goose.SetDialect("postgres")
	require.NoError(t, err)
	err = goose.Up(sqlDB, MigrationsDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})
	return CommonDeps{
		Ctx:    ctx,
		Client: client,
		DSN:    dsn,
	}
}

// CommonDeps defines core dependencies including a context and a database client for executing queries and transactions.
type CommonDeps struct {
	Ctx    context.Context
	Client db.PGXClient
	DSN    string
}

func (d *CommonDeps) GetDSN() string {
	return d.DSN
}
