package service

import (
	"mime/multipart"
	"path/filepath"
	"service/internal/server"
	"service/internal/storage"
	"strings"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
)

// OriginService 原始文件服务。
type OriginService struct {
	postgre *storage.Postgre
	minio   *storage.Minio
}

// createOriginService 创建原始文件服务。
func createOriginService() (*OriginService, error) {
	postgre, err := storage.CreatePostgre()
	if err != nil {
		return nil, err
	}
	minio, err := storage.CreateMinio()
	if err != nil {
		return nil, err
	}
	return &OriginService{postgre: postgre, minio: minio}, nil
}

// Upload 上传原始文件。
func (service *OriginService) Upload(ctx *gin.Context, file multipart.File, header *multipart.FileHeader) (*string, error) {
	cliam := ctx.MustGet("cliam").(server.Claims)
	// 获取文件信息。
	id := strings.ToUpper(uuid.NewV4().String())
	kind := filepath.Ext(header.Filename)
	name := strings.TrimSuffix(header.Filename, kind)
	kind = strings.ToLower(strings.TrimPrefix(kind, "."))
	size := header.Size
	// 获取文件内容。
	hash := ""
	// 插入系统用户。
	err := service.postgre.QueryNaught(ctx, `insert into origin(id,name,type,size,hash,create_user,create_time) values($1,$2,$3,$4,$5,$6,$7)`, id, name, kind, size, hash, cliam.Subject, time.Now())
	if err != nil {
		return nil, err
	}
	// 返沪用户编码。
	return &id, nil
}
