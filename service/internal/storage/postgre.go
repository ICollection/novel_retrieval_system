package storage

import (
	"context"
	"fmt"
	"service/internal/config"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	postgreMutex    sync.Mutex
	postgreInstance *Postgre
)

// Postgre PostgreSQL 数据库。
type Postgre struct {
	pool *pgxpool.Pool
}

// CreatePostgre 创建 PostgreSQL 数据库连接。
func CreatePostgre() (*Postgre, error) {
	postgreMutex.Lock()
	defer postgreMutex.Unlock()
	// 已创建时复用。
	if postgreInstance != nil {
		return postgreInstance, nil
	}
	// 创建数据连接。
	pool, err := pgxpool.New(context.Background(), config.Config.Postgre.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败，详细错误信息：%w", err)
	}
	postgreInstance = &Postgre{pool: pool}
	// 返回数据连接。
	return postgreInstance, nil
}

// QueryNaught 执行查询不返回值。
func (postgre *Postgre) QueryNaught(ctx context.Context, sql string, arguments ...any) error {
	_, err := postgre.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return err
	}
	return nil
}

// QueryMatrix 执行查询返回切片。
func (postgre *Postgre) QueryMatrix[T any](ctx context.Context, sql string, arguments ...any) ([]T, error) {
	rows, err := postgre.pool.Query(ctx, sql, arguments...)
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
func (postgre *Postgre) QuerySingle[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	rows, err := postgre.pool.Query(ctx, sql, arguments...)
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
func (postgre *Postgre) QuerySimple[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	var result T
	err := postgre.pool.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BeginTransaction 开始事务。
func (postgre *Postgre) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return postgre.pool.BeginTx(ctx, pgx.TxOptions{})
}

// CommitTransaction 提交事务。
func (postgre *Postgre) CommitTransaction(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

// RollbackTransaction 回滚事务。
func (postgre *Postgre) RollbackTransaction(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

// Close 关闭连接池。
func (postgre *Postgre) Close() {
	postgre.pool.Close()
}
