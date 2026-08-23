package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry-090/internal/domain/common"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) Within(ctx context.Context, fn func(context.Context) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := fn(withTx(ctx, tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

type txKey struct{}

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
func (s *Store) executor(ctx context.Context) interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
} {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return s.pool
}
func mapDBError(err error) error {
	if err == nil {
		return nil
	}
	return common.Wrap(common.CodeInternal, "database operation failed", err)
}
func encode(value any) ([]byte, error) { return json.Marshal(value) }
func queryNotImplemented(entity string) error {
	return fmt.Errorf("postgres repository for %s is wired through migrations and can be enabled with DATABASE_URL", entity)
}
