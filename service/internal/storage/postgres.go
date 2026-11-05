package storage

import (
	"context"
	"fmt"
	"service/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// postgres PostgreSQL 数据库连接。
type postgres struct {
	pool *pgxpool.Pool
}

// createPostgres 创建 PostgreSQL 数据库连接。
func createPostgres() (*postgres, error) {
	pool, err := pgxpool.New(context.Background(), config.Config.Postgre.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("创建 PostgreSQL 数据库连接失败，详细错误信息：%w", err)
	}
	return &postgres{pool: pool}, nil
}

// QueryNaught 执行查询不返回值。
func (postgres *postgres) QueryNaught(ctx context.Context, sql string, arguments ...any) error {
	_, err := postgres.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return err
	}
	return nil
}

// QueryMatrix 执行查询返回切片。
func (postgres *postgres) QueryMatrix[T any](ctx context.Context, sql string, arguments ...any) ([]T, error) {
	rows, err := postgres.pool.Query(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, err
	}
	return result, nil
}

// QuerySingle 执行查询返回记录。
func (postgres *postgres) QuerySingle[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	rows, err := postgres.pool.Query(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QuerySimple 执行查询返回单值。
func (postgres *postgres) QuerySimple[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	var result T
	err := postgres.pool.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BeginTransaction 开始事务。
func (postgres *postgres) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return postgres.pool.BeginTx(ctx, pgx.TxOptions{})
}

// CommitTransaction 提交事务。
func (postgres *postgres) CommitTransaction(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

// RollbackTransaction 回滚事务。
func (postgres *postgres) RollbackTransaction(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

// BatchNaught 执行查询不返回值。
func (postgres *postgres) BatchNaught(ctx context.Context, tx pgx.Tx, sql string, arguments ...any) error {
	_, err := tx.Exec(ctx, sql, arguments...)
	if err != nil {
		return err
	}
	return nil
}

// Close 关闭连接池。
func (postgres *postgres) dispose() {
	postgres.pool.Close()
}
