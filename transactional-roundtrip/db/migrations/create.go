package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/jmoiron/sqlx"
)

func CreateDB(ctx context.Context, dsn string) (bool, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return false, fmt.Errorf("DB URL is invalid: %w", err)
	}

	dbName := path.Base(u.Path)
	u.Path = path.Dir(u.Path)
	if dbName == "" {
		return false, fmt.Errorf("missing database name in URL path: %q", u.String())
	}

	db, err := sqlx.Open(dbDriver, u.String())
	if err != nil {
		return false, fmt.Errorf("could not connect to database server %v: %w", u, err)
	}
	defer func() {
		// will restart from a fresh pool
		_ = db.Close()
	}()

	cancellable, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := db.BeginTxx(cancellable, nil)
	if err != nil {
		return false, err
	}

	ok, err := dbExists(cancellable, tx, dbName)
	if err != nil {
		return false, err
	}

	if ok {
		return false, nil
	}

	_, err = db.ExecContext(cancellable, fmt.Sprintf(`CREATE DATABASE %s`, dbName))
	if err != nil {
		return false, fmt.Errorf("could not create database %s: %w", dbName, err)
	}

	err = tx.Commit()
	if err != nil {
		return false, fmt.Errorf("could not create database %s: %w", dbName, err)
	}

	return true, nil
}

func dbExists(ctx context.Context, tx *sqlx.Tx, dbName string) (bool, error) {
	var ignored sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT datname FROM pg_database WHERE datname = $1", dbName).Scan(&ignored)
	if err == nil {
		// already there
		return true, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return false, err
}
