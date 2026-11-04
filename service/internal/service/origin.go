package service

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
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
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败，详细错误信息：%w", err)
	}
	hash := sha1.Sum(content)
	// 上传对象存储。
	if err := service.minio.Upload(ctx, "Origin", id+"."+kind, content, header.Header.Get("Content-Type")); err != nil {
		return nil, fmt.Errorf("保存文件失败，详细错误信息：%w", err)
	}
	// 插入原始文件。
	err = service.postgre.QueryNaught(ctx, `insert into origin(id,name,type,size,hash,create_user,create_time) values($1,$2,$3,$4,$5,$6,$7)`, id, name, kind, size, base64.StdEncoding.EncodeToString(hash[:]), cliam.Subject, time.Now())
	if err != nil {
		return nil, err
	}
	// 返回文件编码。
	return &id, nil
}
