package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed postgres_schema.sql
var postgresSchema string

// Database keeps the existing SQLite development path and binds the same
// parameterized application queries for PostgreSQL. Values are never interpolated.
type Database struct {
	*sql.DB
	postgres bool
}
type Transaction struct {
	*sql.Tx
	postgres bool
}

func (d *Database) Mode() string {
	if d.postgres {
		return "postgres (Supabase)"
	}
	return "sqlite (local)"
}
func openDB(path string) (*Database, error) {
	db, err := openSQLite(path)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}
func openConfiguredDB(c Config) (*Database, error) {
	if c.DatabaseURL == "" {
		return openDB(c.DBPath)
	}
	return openPostgres(c.DatabaseURL)
}
func openPostgres(connection string) (*Database, error) {
	return openPostgresSchema(connection, "researchbridge")
}

func openPostgresSchema(connection, schema string) (*Database, error) {
	if schema != "researchbridge" && !strings.HasPrefix(schema, "rb_test_") {
		return nil, errors.New("invalid application schema")
	}
	for _, r := range schema {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return nil, errors.New("invalid application schema")
		}
	}
	u, err := url.Parse(connection)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" {
		return nil, errors.New("DATABASE_URL must be a postgres:// or postgresql:// URI")
	}
	// Remote database connections require TLS; never fall back to cleartext.
	if host := u.Hostname(); host != "localhost" && host != "127.0.0.1" && host != "::1" {
		q := u.Query()
		mode := q.Get("sslmode")
		if mode == "" {
			q.Set("sslmode", "require")
			u.RawQuery = q.Encode()
		} else if mode != "require" && mode != "verify-ca" && mode != "verify-full" {
			return nil, errors.New("remote DATABASE_URL requires sslmode=require, verify-ca or verify-full")
		}
	}
	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, errors.New("invalid DATABASE_URL: copy the PostgreSQL Session pooler URI and URL-encode the password")
	}
	// An application-only schema avoids exposing custom auth tables through the
	// default Supabase public-schema Data API. Use the Go backend for all access.
	cfg.RuntimeParams["search_path"] = schema
	cfg.RuntimeParams["statement_timeout"] = "15000"
	cfg.ConnectTimeout = 15 * time.Second
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.New("PostgreSQL connection failed: check DATABASE_URL, database password, project status and network; no SQLite fallback was used")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.Close()
		return nil, errors.New("could not start PostgreSQL migration")
	}
	// Serialize startup migrations across laptops / instances.
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(73462101)"); err == nil {
		_, err = tx.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema)
	}
	if err == nil {
		// Individual statements work with pgx extended protocol as well.
		for _, statement := range strings.Split(postgresSchema, ";") {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if _, err = tx.ExecContext(ctx, statement); err != nil {
				break
			}
		}
	}
	if err == nil {
		_, err = tx.ExecContext(ctx, "REVOKE ALL ON SCHEMA "+schema+" FROM PUBLIC")
	}
	if err == nil {
		err = tx.Commit()
	} else {
		tx.Rollback()
	}
	if err != nil {
		db.Close()
		return nil, errors.New("PostgreSQL schema setup failed: check database CREATE privileges and existing researchbridge tables")
	}
	return &Database{DB: db, postgres: true}, nil
}

// bindSQL skips SQL string literals and quoted identifiers. Application SQL uses
// ? only for parameters; it does not use PostgreSQL JSON question-mark operators.
func bindSQL(query string, postgres bool) string {
	if !postgres {
		return query
	}
	var out strings.Builder
	var quote byte
	index := 0
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if quote != 0 {
			out.WriteByte(ch)
			if ch == quote {
				if i+1 < len(query) && query[i+1] == quote {
					i++
					out.WriteByte(query[i])
				} else {
					quote = 0
				}
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
		}
		if ch == '?' {
			index++
			out.WriteByte('$')
			out.WriteString(strconv.Itoa(index))
		} else {
			out.WriteByte(ch)
		}
	}
	return out.String()
}

// Existing schemas store flags as 0/1. database/sql scans these into Go bools.
func sqlArgs(args []any) []any {
	out := make([]any, len(args))
	for i, v := range args {
		if b, ok := v.(bool); ok {
			if b {
				out[i] = int64(1)
			} else {
				out[i] = int64(0)
			}
		} else {
			out[i] = v
		}
	}
	return out
}
func (d *Database) Exec(q string, args ...any) (sql.Result, error) {
	return d.DB.Exec(bindSQL(q, d.postgres), sqlArgs(args)...)
}
func (d *Database) Query(q string, args ...any) (*sql.Rows, error) {
	return d.DB.Query(bindSQL(q, d.postgres), sqlArgs(args)...)
}
func (d *Database) QueryRow(q string, args ...any) *sql.Row {
	return d.DB.QueryRow(bindSQL(q, d.postgres), sqlArgs(args)...)
}
func (d *Database) InsertID(q string, args ...any) (int64, error) {
	var id int64
	err := d.QueryRow(q+" RETURNING id", args...).Scan(&id)
	return id, err
}
func (d *Database) Begin() (*Transaction, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	if d.postgres {
		// Preserve the existing single-writer SQLite transaction semantics across
		// instances: invitation limits, meeting checks and one-time tokens are atomic.
		// This conservative lock can be replaced by scoped row locks as usage grows.
		if _, err = tx.Exec("SELECT pg_advisory_xact_lock(73462102)"); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("could not lock application transaction: %w", err)
		}
	}
	return &Transaction{Tx: tx, postgres: d.postgres}, nil
}
func (t *Transaction) Exec(q string, args ...any) (sql.Result, error) {
	return t.Tx.Exec(bindSQL(q, t.postgres), sqlArgs(args)...)
}
func (t *Transaction) Query(q string, args ...any) (*sql.Rows, error) {
	return t.Tx.Query(bindSQL(q, t.postgres), sqlArgs(args)...)
}
func (t *Transaction) QueryRow(q string, args ...any) *sql.Row {
	return t.Tx.QueryRow(bindSQL(q, t.postgres), sqlArgs(args)...)
}
func (t *Transaction) InsertID(q string, args ...any) (int64, error) {
	var id int64
	err := t.QueryRow(q+" RETURNING id", args...).Scan(&id)
	return id, err
}
