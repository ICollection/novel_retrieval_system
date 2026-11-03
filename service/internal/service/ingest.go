package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"

	"service/internal/model"
	"service/internal/storage/minio"
	"service/internal/storage/postgres"
)

// ErrDuplicateFile 表示相同内容的上传文件已存在。
var ErrDuplicateFile = errors.New("文件已存在")

// IngestService 文件接入与处理业务。
//
// 负责上传文件、创建处理任务、维护与查询处理进度。
type IngestService struct {
	postgres *postgres.Store
	objects  *minio.Store
}

// NewIngestService 创建文件接入业务。
func NewIngestService(pg *postgres.Store, objects *minio.Store) *IngestService {
	return &IngestService{postgres: pg, objects: objects}
}

// Upload 保存上传文件到对象存储与数据库，返回原始文件记录。
func (s *IngestService) Upload(ctx context.Context, name string, size int64, data []byte) (*model.Origin, error) {
	sum := sha1.Sum(data)
	hash := hex.EncodeToString(sum[:])

	// 1. 查重
	var exists bool
	err := s.postgres.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM origin WHERE hash = $1)`, hash).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("查询文件摘要失败，详细错误信息：%w", err)
	}
	if exists {
		return nil, ErrDuplicateFile
	}

	id := uuid.NewV7().String()
	origin := model.Origin{
		ID:         id,
		Name:       name,
		Type:       fileType(name),
		Size:       size,
		Hash:       hash,
		CreateTime: time.Now(),
	}

	object := objectName(id, origin.Type)
	if err = s.objects.Put(ctx, object, data); err != nil {
		return nil, err
	}
	// 3. 写 origin 表
	_, err = s.postgres.Pool().Exec(ctx,
		`INSERT INTO origin (id, name, type, size, hash, create_time) VALUES ($1, $2, $3, $4, $5, $6)`,
		origin.ID, origin.Name, origin.Type, origin.Size, origin.Hash, origin.CreateTime)
	if err != nil {
		// 对象存储回滚
		if removeErr := s.objects.Remove(ctx, object); removeErr != nil {
			return nil, fmt.Errorf("%w；清理对象失败：%v", err, removeErr)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateFile
		}
		return nil, err
	}
	return &origin, nil
}

// fileType 取文件扩展名作为类型。
func fileType(name string) string {
	kind := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if len(kind) > 10 {
		kind = kind[:10]
	}
	return kind
}

// objectName 生成对象存储中的文件名。
func objectName(id string, kind string) string {
	if kind == "" {
		return "origin/" + id
	}
	return "origin/" + id + "." + kind
}
