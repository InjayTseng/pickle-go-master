package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBTX is a common interface for *sqlx.DB and *sqlx.Tx
type DBTX interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
}

// TxManager manages database transactions
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager creates a new TxManager
func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// DB returns the underlying database connection
func (m *TxManager) DB() *sqlx.DB {
	return m.db
}

// WithTx executes a function within a transaction
// If the function returns an error, the transaction is rolled back
// If the function panics, the transaction is rolled back and the panic is re-raised
func (m *TxManager) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// WithTxResult executes a function within a transaction and returns a result
func WithTxResult[T any](m *TxManager, ctx context.Context, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	var result T
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return result, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	result, err = fn(tx)
	if err != nil {
		tx.Rollback()
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}

	return result, nil
}
