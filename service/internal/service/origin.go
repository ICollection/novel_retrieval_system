package service

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"service/internal/common"
	"service/internal/server"
	"service/internal/storage"
	"strings"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
)

// originService 原始文件服务。
type originService struct {
}

// Upload 上传原始文件。
func (service *originService) Upload(ctx *gin.Context, file multipart.File, header *multipart.FileHeader) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 获取文件信息。
	moment := time.Now().UTC()
	origin_id := strings.ToUpper(uuid.NewV4().String())
	novel_id := strings.ToUpper(uuid.NewV4().String())
	mission_id := strings.ToUpper(uuid.NewV4().String())
	kind := filepath.Ext(header.Filename)
	name := strings.TrimSuffix(header.Filename, kind)
	kind = strings.ToLower(strings.TrimPrefix(kind, "."))
	size := header.Size
	// 校验文件扩展名。
	if common.ExistsFileType(kind) == false {
		return nil, errors.New("文件类型错误。")
	}
	// 获取文件内容。
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败，详细错误信息：%w", err)
	}
	sums := sha1.Sum(content)
	hash := base64.StdEncoding.EncodeToString(sums[:])
	// 验证文件重复。
	count, err := storage.RDB.QuerySimple[int](ctx, `select count(*) from origin where hash=$1`, hash)
	if err != nil {
		return nil, err
	}
	if *count > 0 {
		return nil, errors.New("相同文件已经存在。")
	}
	// 上传对象存储。
	if err := storage.OSS.Upload(ctx, "Origin", origin_id+"."+kind, content, header.Header.Get("Content-Type")); err != nil {
		return nil, fmt.Errorf("保存文件失败，详细错误信息：%w", err)
	}
	// 事务保存数据。
	var batch storage.Batch
	batch.Append("insert into origin(id,name,type,size,hash,create_user,create_time) values($1,$2,$3,$4,$5,$6,$7);", origin_id, name, kind, size, hash, claims.Subject, moment)
	batch.Append("insert into novel(id,origin,title,character_count,chapter_count,create_user,create_time,update_time) values($1,$2,$3,$4,$5,$6,$7,$8)", novel_id, origin_id, name, 0, 0, claims.Subject, moment, moment)
	batch.Append("insert into mission(id,novel,status,total_count,embed_count,launch_time,finish_time,update_time,retry_count,message) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)", mission_id, novel_id, "01", 0, 0, nil, nil, moment, 0, nil)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return nil, err
	}
	// 返回文件编码。
	return &origin_id, nil
}
