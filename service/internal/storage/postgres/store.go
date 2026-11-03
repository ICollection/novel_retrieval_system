package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"service/internal/config"
	"service/internal/model"
)

// ErrDuplicate 表示记录违反唯一约束。
var ErrDuplicate = errors.New("记录已存在")

// Store PostgreSQL 存储，封装连接与原子读写。
type Store struct {
	pool *pgxpool.Pool
}

var (
	defaultOnce  sync.Once
	defaultStore *Store
	defaultErr   error
)

// Default 返回默认存储，首次调用时创建连接池，不校验连通性。
func Default() (*Store, error) {
	defaultOnce.Do(func() {
		defaultStore, defaultErr = New()
	})
	return defaultStore, defaultErr
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

// Close 关闭连接池。
func (s *Store) Close() {
	s.pool.Close()
}

// OriginExistsByHash 判断文件摘要是否已存在。
func (s *Store) OriginExistsByHash(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM origin WHERE hash = $1)`, hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("查询文件摘要失败，详细错误信息：%w", err)
	}
	return exists, nil
}

// CreateOrigin 写入原始文件记录。
func (s *Store) CreateOrigin(ctx context.Context, origin model.Origin) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO origin (id, name, type, size, hash, create_time) VALUES ($1, $2, $3, $4, $5, $6)`,
		origin.ID, origin.Name, origin.Type, origin.Size, origin.Hash, origin.CreateTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicate
		}
		return fmt.Errorf("写入原始文件记录失败，详细错误信息：%w", err)
	}
	return nil
}
