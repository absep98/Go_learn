// Package testdb gives each test package its own Postgres database.
//
// `go test ./...` runs packages as parallel processes. When the db and handlers
// test packages shared one database, one package's TRUNCATE could wipe rows the
// other was in the middle of using — an intermittent failure that only showed
// up when the two processes happened to overlap. Reproduced on demand by looping
// the db tests while repeatedly starting the handlers tests (6 failed cases);
// with a database per package the same run has 0 failures.
//
// Imported only from _test.go files, so it is never linked into the server.
package testdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/lib/pq"
)

// DSNs returns connection strings for the database "<DB_NAME>_<suffix>" on the
// server described by DB_HOST/DB_PORT/DB_USER/DB_PASSWORD, creating it on first
// use. ok is false when DB_HOST is unset (integration tests should skip).
func DSNs(suffix string) (dsn, migrateDSN string, ok bool, err error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		return "", "", false, nil
	}
	port, user, pass, base := os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")
	name := base + "_" + suffix

	admin, err := sql.Open("postgres", keyValueDSN(host, port, user, pass, base))
	if err != nil {
		return "", "", true, err
	}
	defer admin.Close()

	if _, err := admin.Exec("CREATE DATABASE " + pq.QuoteIdentifier(name)); err != nil {
		var pqErr *pq.Error
		if !errors.As(err, &pqErr) || pqErr.Code != "42P04" { // 42P04 = duplicate_database
			return "", "", true, fmt.Errorf("creating test database %s: %w", name, err)
		}
	}

	return keyValueDSN(host, port, user, pass, name),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name),
		true, nil
}

func keyValueDSN(host, port, user, pass, dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}
