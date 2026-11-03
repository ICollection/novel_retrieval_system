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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"service/internal/model"
	"service/internal/storage"
	"service/internal/storage/minio"
)

// ErrDuplicateFile 表示相同内容的上传文件已存在。
var ErrDuplicateFile = errors.New("文件已存在")

// IngestService 文件接入业务。
//
// 负责上传文件到对象存储、登记原始文件记录。
type IngestService struct {
	postgre *storage.Postgre
	objects *minio.Store
}

// NewIngestService 创建文件接入业务。
func NewIngestService(postgre *storage.Postgre, objects *minio.Store) *IngestService {
	return &IngestService{postgre: postgre, objects: objects}
}

// Upload 保存上传文件到对象存储与数据库，返回原始文件记录。
func (service *IngestService) Upload(ctx context.Context, name string, size int64, data []byte) (*model.Origin, error) {
	// 计算文件摘要。
	sum := sha1.Sum(data)
	hash := hex.EncodeToString(sum[:])

	// 判断摘要是否重复。
	exists, err := service.postgre.QuerySimple[bool](ctx, `select exists(select 1 from origin where hash=$1);`, hash)
	if err != nil {
		return nil, fmt.Errorf("查询文件摘要失败，详细错误信息：%w", err)
	}
	if *exists {
		return nil, ErrDuplicateFile
	}

	// 生成原始文件记录。
	code, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("生成文件编码失败，详细错误信息：%w", err)
	}
	id := strings.ToUpper(code.String())
	kind := fileType(name)
	origin := model.Origin{
		ID:         id,
		Name:       name,
		Type:       kind,
		Size:       size,
		Hash:       hash,
		CreateTime: time.Now(),
	}

	// 上传对象文件。
	object := objectName(id, kind)
	if err = service.objects.Put(ctx, object, data); err != nil {
		return nil, err
	}

	// 插入原始文件。
	err = service.postgre.QueryNaught(ctx, `insert into origin(id,name,type,size,hash,create_time) values($1,$2,$3,$4,$5,$6);`,
		origin.ID, origin.Name, origin.Type, origin.Size, origin.Hash, origin.CreateTime)
	if err != nil {
		// 清理对象文件。
		if removeErr := service.objects.Remove(ctx, object); removeErr != nil {
			return nil, fmt.Errorf("%w；清理对象文件失败：%v", err, removeErr)
		}
		// 判断摘要是否重复，并发上传同一文件时由唯一约束兜底。
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
