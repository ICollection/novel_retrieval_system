package storage

import (
	"context"
	"fmt"
	"service/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// postgres PostgreSQL 数据库连接。
type postgres struct {
	pool *pgxpool.Pool
}

// createPostgres 创建 PostgreSQL 数据库连接。
func createPostgres() (*postgres, error) {
	pool, err := pgxpool.New(context.Background(), config.Config.Postgres.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("创建 PostgreSQL 数据库连接失败，详细错误信息：%w", err)
	}
	return &postgres{pool: pool}, nil
}

// QueryNaught 执行查询不返回值。
func (postgres *postgres) QueryNaught(ctx context.Context, sql string, arguments ...any) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Postgres.Timeout))
	defer cancel()
	_, err := postgres.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return err
	}
	return nil
}

// QueryMatrix 执行查询返回切片。
func (postgres *postgres) QueryMatrix[T any](ctx context.Context, sql string, arguments ...any) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Postgres.Timeout))
	defer cancel()
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
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Postgres.Timeout))
	defer cancel()
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
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Postgres.Timeout))
	defer cancel()
	var result T
	err := postgres.pool.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchNaught 执行事务不返回值。
func (postgres *postgres) BatchNaught(ctx context.Context, batch *Batch) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Postgres.Timeout))
	defer cancel()
	// 打开查询事务。
	transaction, err := postgres.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)
	// 执行批量查询。
	for _, command := range batch.commands {
		_, err := transaction.Exec(ctx, command.sql, command.arguments...)
		if err != nil {
			return err
		}
	}
	// 提交查询事务。
	return transaction.Commit(ctx)
}

// Close 关闭连接池。
func (postgres *postgres) dispose() {
	postgres.pool.Close()
}

// command 查询命令。
type command struct {
	sql       string
	arguments []any
}

// Batch 查询批次。
type Batch struct {
	commands []command
}

// AppendCommand 添加批次命令。
func (batch *Batch) AppendCommand(sql string, arguments ...any) {
	batch.commands = append(batch.commands, command{sql: sql, arguments: arguments})
}

// CreateBatch 创建查询批次。
func (postgres *postgres) CreateBatch() *Batch {
	return &Batch{commands: []command{}}
}
