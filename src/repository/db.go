package repository

import (
	"context"
	"database/sql"
)

type DB interface {
	Close() error
	Get() (*sql.DB, error)
	Transaction(ctx context.Context, txOpt *sql.TxOptions, fn func(ctx context.Context) error) error
}
