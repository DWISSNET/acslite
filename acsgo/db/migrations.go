package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed schema.sql
var schemaSQLite string

// Migrate runs the SQL schema against the given database connection.
// For PostgreSQL, it adapts the SQLite-flavoured DDL automatically.
func Migrate(db *sql.DB, dbType string) error {
	schema := schemaSQLite
	if dbType == "postgres" {
		schema = adaptSchemaForPostgres(schema)
	}

	// Split on semicolons and run each statement
	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed for stmt [%.80s...]: %w", stmt, err)
		}
	}
	return nil
}

// adaptSchemaForPostgres converts SQLite-specific DDL to PostgreSQL-compatible DDL.
func adaptSchemaForPostgres(schema string) string {
	replacements := map[string]string{
		"INTEGER PRIMARY KEY AUTOINCREMENT": "SERIAL PRIMARY KEY",
		"CURRENT_TIMESTAMP":                 "NOW()",
	}
	for from, to := range replacements {
		schema = strings.ReplaceAll(schema, from, to)
	}
	return schema
}
