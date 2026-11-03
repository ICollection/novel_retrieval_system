package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"service/internal/config"
)

// Store PostgreSQL 存储，封装连接池。
type Store struct {
	pool *pgxpool.Pool
}

// New 创建 PostgreSQL 存储，不校验数据库连通性。
func New() (*Store, error) {
	cfg := config.Config.PostgreSQL
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     "/" + cfg.Database,
		RawQuery: "sslmode=" + sslMode,
	}
	pool, err := pgxpool.New(context.Background(), dsn.String())
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败，详细错误信息：%w", err)
	}
	return &Store{pool: pool}, nil
}

// Pool 返回底层连接池，供 service 层直接执行 SQL。
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Close 关闭连接池。
func (s *Store) Close() {
	s.pool.Close()
}
