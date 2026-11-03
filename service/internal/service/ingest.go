package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"uuid"

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
	once     sync.Once
	postgres *postgres.Store
	objects  *minio.Store
	err      error
}

// NewIngestService 创建文件接入业务，存储客户端在首次使用时创建。
func NewIngestService() *IngestService {
	return &IngestService{}
}

// stores 返回存储实例，首次调用时初始化。
func (s *IngestService) stores() (*postgres.Store, *minio.Store, error) {
	s.once.Do(func() {
		s.postgres, s.err = postgres.Default()
		if s.err != nil {
			return
		}
		s.objects, s.err = minio.Default()
	})
	return s.postgres, s.objects, s.err
}

// Upload 保存上传文件到对象存储与数据库，返回原始文件记录。
func (s *IngestService) Upload(ctx context.Context, name string, size int64, data []byte) (*model.Origin, error) {
	pg, objects, err := s.stores()
	if err != nil {
		return nil, err
	}

	sum := sha1.Sum(data)
	hash := hex.EncodeToString(sum[:])

	exists, err := pg.OriginExistsByHash(ctx, hash)
	if err != nil {
		return nil, err
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
	if err = objects.Put(ctx, object, data); err != nil {
		return nil, err
	}
	if err = pg.CreateOrigin(ctx, origin); err != nil {
		if removeErr := objects.Remove(ctx, object); removeErr != nil {
			return nil, fmt.Errorf("%w；清理对象失败：%v", err, removeErr)
		}
		if errors.Is(err, postgres.ErrDuplicate) {
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
