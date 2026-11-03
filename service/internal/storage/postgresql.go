package storage

import (
	"context"
	"fmt"
	"service/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL PostgreSQL 数据库。
type PostgreSQL struct {
	pool *pgxpool.Pool
}

// Create 创建 PostgreSQL 数据库连接。
func Create() (*PostgreSQL, error) {
	pool, err := pgxpool.New(context.Background(), config.Config.PostgreSQL.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败，详细错误信息：%w", err)
	}
	return &PostgreSQL{pool: pool}, nil
}

// QueryEmpty 执行查询不返回值。
func (postgresql *PostgreSQL) QueryEmpty(ctx context.Context, sql string, arguments ...any) error {
	_, err := postgresql.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return err
	}
	return nil
}

// QuerySlice 执行查询返回切片。
func (postgresql *PostgreSQL) QuerySlice[T any](ctx context.Context, sql string, arguments ...any) ([]T, error) {
	rows, err := postgresql.pool.Query(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}
	return result, nil
}

// QuerySingle 执行查询返回记录。
func (postgresql *PostgreSQL) QuerySingle[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	rows, err := postgresql.pool.Query(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QuerySimple 执行查询返回单值。
func (postgresql *PostgreSQL) QuerySimple[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	var result T
	err := postgresql.pool.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Close 关闭连接池。
func (postgresql *PostgreSQL) Close() {
	postgresql.pool.Close()
}
