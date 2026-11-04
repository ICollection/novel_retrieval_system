package service

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
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
	claims := ctx.MustGet("claims").(*server.Claims)
	// 获取文件信息。
	moment := time.Now()
	origin_id := strings.ToUpper(uuid.NewV4().String())
	novel_id := strings.ToUpper(uuid.NewV4().String())
	progress_id := strings.ToUpper(uuid.NewV4().String())
	kind := filepath.Ext(header.Filename)
	name := strings.TrimSuffix(header.Filename, kind)
	kind = strings.ToLower(strings.TrimPrefix(kind, "."))
	size := header.Size
	// 获取文件内容。
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败，详细错误信息：%w", err)
	}
	sums := sha1.Sum(content)
	hash := base64.StdEncoding.EncodeToString(sums[:])
	// 验证文件重复。
	count, err := service.postgre.QuerySimple[int](ctx, `select count(*) from origin where hash=$1`, hash)
	if err != nil {
		return nil, err
	}
	if *count > 0 {
		return nil, errors.New("相同文件已经存在。")
	}
	// 启动数据事务。
	transaction, err := service.postgre.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer service.postgre.RollbackTransaction(ctx, transaction)
	// 插入原始记录。
	err = service.postgre.BatchNaught(ctx, transaction, `insert into origin(id,name,type,size,hash,create_user,create_time) values($1,$2,$3,$4,$5,$6,$7)`, origin_id, name, kind, size, hash, claims.Subject, moment)
	if err != nil {
		return nil, err
	}
	// 插入小说记录。
	err = service.postgre.BatchNaught(ctx, transaction, `insert into novel(id,origin,title,character_count,chapter_count,create_user,create_time,update_time) values($1,$2,$3,$4,$5,$6,$7,$8)`, novel_id, origin_id, name, 0, 0, claims.Subject, moment, moment)
	if err != nil {
		return nil, err
	}
	// 插入进度记录。
	err = service.postgre.BatchNaught(ctx, transaction, `insert into progress(id,novel,status,total_count,embed_count,launch_time,finish_time) values($1,$2,$3,$4,$5,$6,$7)`, progress_id, novel_id, "01", 0, 0, nil, nil)
	if err != nil {
		return nil, err
	}
	// 上传对象存储。
	if err := service.minio.Upload(ctx, "Origin", origin_id+"."+kind, content, header.Header.Get("Content-Type")); err != nil {
		return nil, fmt.Errorf("保存文件失败，详细错误信息：%w", err)
	}
	// 提交数据事务。
	err = service.postgre.CommitTransaction(ctx, transaction)
	if err != nil {
		return nil, err
	}
	// 返回文件编码。
	return &origin_id, nil
}
