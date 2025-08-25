package adapter

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolAdapter struct{ *pgxpool.Pool }

func New(pool *pgxpool.Pool) *PoolAdapter { return &PoolAdapter{Pool: pool} }

// добавляем недостающие методы интерфейса:
func (p *PoolAdapter) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, p.Pool, f)
}
func (p *PoolAdapter) BeginTxFunc(ctx context.Context, opts pgx.TxOptions, f func(pgx.Tx) error) error {
	return pgx.BeginTxFunc(ctx, p.Pool, opts, f)
}

// остальные методы Pool уже есть, но интерфейс их требует явно:
func (p *PoolAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.Pool.Query(ctx, sql, args...)
}
func (p *PoolAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}
func (p *PoolAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, sql, args...)
}
