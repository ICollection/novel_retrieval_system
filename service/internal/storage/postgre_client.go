package storage

import (
	"context"
	"errors"
	"log/slog"
	"service/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreClient PostgreSQL 数据库连接。
type PostgreClient struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

// newPostgreClient 创建 PostgreSQL 数据库连接。
//
// param: cfg 数据库连接配置。
// return: 数据库连接，错误信息。
func newPostgreClient(cfg config.Postgres) (*PostgreClient, error) {
	pool, err := pgxpool.New(context.Background(), cfg.GetDSN())
	if err != nil {
		slog.Error("[RDB]创建连接失败", "error", err)
		return nil, err
	}
	return &PostgreClient{pool: pool, timeout: time.Second * time.Duration(cfg.Timeout)}, nil
}

// QueryNaught 执行查询不返回值。
//
// param: ctx 上下文。
// param: sql 查询语句。
// param: arguments 查询参数。
// return: 错误信息。
func (pc *PostgreClient) QueryNaught(ctx context.Context, sql string, arguments ...any) error {
	ctx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()
	_, err := pc.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		slog.Error("[RDB]执行查询失败", "error", err)
		return err
	}
	return nil
}

// QueryMatrix 执行查询返回切片。
//
// param: ctx 上下文。
// param: sql 查询语句。
// param: arguments 查询参数。
// return: 数据切片，错误信息。
func (pc *PostgreClient) QueryMatrix[T any](ctx context.Context, sql string, arguments ...any) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()
	rows, err := pc.pool.Query(ctx, sql, arguments...)
	if err != nil {
		slog.Error("[RDB]执行查询失败", "error", err)
		return nil, err
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		slog.Error("[RDB]执行查询失败", "error", err)
		return nil, err
	}
	return result, nil
}

// QuerySingle 执行查询返回记录。
//
// param: ctx 上下文。
// param: sql 查询语句。
// param: arguments 查询参数。
// return: 数据记录，错误信息。
func (pc *PostgreClient) QuerySingle[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()
	rows, err := pc.pool.Query(ctx, sql, arguments...)
	if err != nil {
		slog.Error("[RDB]执行查询失败", "error", err)
		return nil, err
	}
	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			slog.Error("[RDB]执行查询失败", "error", err)
			return nil, err
		}
	}
	return &result, nil
}

// QuerySimple 执行查询返回单值。
//
// param: ctx 上下文。
// param: sql 查询语句。
// param: arguments 查询参数。
// return: 数据单值，错误信息。
func (pc *PostgreClient) QuerySimple[T any](ctx context.Context, sql string, arguments ...any) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()
	var result *T = nil
	err := pc.pool.QueryRow(ctx, sql, arguments...).Scan(&result)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			slog.Error("[RDB]执行查询失败", "error", err)
			return nil, err
		}
	}
	return result, nil
}

// BatchNaught 执行事务不返回值。
//
// param: ctx 上下文。
// param: batch 事务批次。
// return: 错误信息。
func (pc *PostgreClient) BatchNaught(ctx context.Context, batch *Batch) error {
	ctx, cancel := context.WithTimeout(ctx, pc.timeout)
	defer cancel()
	// 打开查询事务。
	transaction, err := pc.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)
	// 执行批量查询。
	for _, command := range batch.commands {
		_, err := transaction.Exec(ctx, command.sql, command.arguments...)
		if err != nil {
			slog.Error("[RDB]执行事务失败", "error", err)
			return err
		}
	}
	// 提交查询事务。
	err = transaction.Commit(ctx)
	if err != nil {
		slog.Error("[RDB]提交事务失败", "error", err)
		return err
	}
	return nil
}

// Close 关闭连接池。
func (pc *PostgreClient) dispose() {
	pc.pool.Close()
}

// Batch 事务批次。
type Batch struct {
	commands []command
}

// command 查询命令。
type command struct {
	sql       string
	arguments []any
}

// Append 添加查询命令。
//
// param: sql 查询语句。
// param: arguments 查询参数。
func (batch *Batch) Append(sql string, arguments ...any) {
	batch.commands = append(batch.commands, command{sql: sql, arguments: arguments})
}
