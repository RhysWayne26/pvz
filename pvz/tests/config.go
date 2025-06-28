package tests

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// TestDBName represents the name of the test database used during integration tests.
	TestDBName = "pvz_test"

	// TestDBUser specifies the username for the test database used during integration tests.
	TestDBUser = "test_user"

	// TestDBPassword specifies the password for the test database used during integration tests.
	TestDBPassword = "test_pass"

	// TestStandaloneDSN is used for standalone integration tests with local Docker DB
	TestStandaloneDSN = "postgres://test_user:test_pass@localhost:5455/pvz_test?sslmode=disable"

	TruncateOrderSql = `TRUNCATE orders RESTART IDENTITY CASCADE;`

	// TruncateHistorySQL defines the SQL query to truncate the `order_history` table and reset its identity sequence.
	TruncateHistorySQL = `TRUNCATE order_history RESTART IDENTITY CASCADE;`
)

// BuildDSN constructs a PostgreSQL DSN using the given host and port with test database credentials.
// For tests only, not intended to be used in a project manually
func BuildDSN(host, port string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		TestDBUser, TestDBPassword, host, port, TestDBName)
}
