package service

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
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

// prepareService 准备资料服务。
type prepareService struct {
}

// UploadOrigin 上传原始文件。
//
// param: ctx 上下文。
// param: file 上传文件。
// param: header 文件头部。
// return: 原始编码，错误信息。
func (ps *prepareService) UploadOrigin(ctx *gin.Context, file multipart.File, header *multipart.FileHeader) (*string, error) {
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
		return nil, err
	}
	sums := sha1.Sum(content)
	hash := base64.StdEncoding.EncodeToString(sums[:])
	// 验证文件重复。
	count, err := storage.RDB.QuerySimple[int](ctx, `select count(*) from origin where hash=$1`, hash)
	if err != nil {
		return nil, err
	}
	if *count > 0 {
		return nil, errors.New("文件内容重复。")
	}
	// 上传对象存储。
	if err := storage.OSS.Upload(ctx, "Origin", origin_id+"."+kind, content, header.Header.Get("Content-Type")); err != nil {
		return nil, err
	}
	// 事务保存数据。
	var batch storage.Batch
	batch.Append("insert into origin(id,name,type,size,hash,create_user,create_time) values($1,$2,$3,$4,$5,$6,$7);", origin_id, name, kind, size, hash, claims.Subject, moment)
	batch.Append("insert into novel(id,origin,title,character_count,chapter_count,create_user,create_time,update_time) values($1,$2,$3,$4,$5,$6,$7,$8);", novel_id, origin_id, name, 0, 0, claims.Subject, moment, moment)
	batch.Append("insert into mission(id,novel,status,total_count,embed_count,launch_time,finish_time,update_time,retry_count,message) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);", mission_id, novel_id, "01", 0, 0, nil, nil, moment, 0, nil)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return nil, err
	}
	// 返回文件编码。
	return &origin_id, nil
}

// FilterMission 查询处理任务。
//
// param: ctx 上下文。
// return: 处理任务，错误信息。
func (ps *prepareService) FilterMission(ctx *gin.Context) ([]mission, error) {
	result, err := storage.RDB.QueryMatrix[mission](ctx, "select mission.id,novel.id as novel_code,novel.title as novel_name,mission.status,mission.total_count,mission.embed_count,mission.launch_time,mission.finish_time,mission.update_time,mission.retry_count,mission.message from mission inner join novel on novel.id=mission.novel order by mission.update_time desc;")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LaunchMission 启动处理任务。
//
// param: ctx 上下文。
// param: id 任务编码。
// return: 任务编码，错误信息。
func (ps *prepareService) LaunchMission(ctx *gin.Context, id *string) (*string, error) {
	// 查询分析任务。
	mission, err := storage.RDB.QuerySingle[mission](ctx, "select mission.id,novel.id as novel_code,novel.title as novel_name,mission.status,mission.total_count,mission.embed_count,mission.launch_time,mission.finish_time,mission.update_time,mission.retry_count,mission.message from mission inner join novel on novel.id=mission.novel where mission.id=$1;", id)
	if err != nil {
		return nil, err
	}
	// 验证任务状态。
	if *mission.Status != common.MissionStatus01 && *mission.Status != common.MissionStatus07 && *mission.Status != common.MissionStatus08 {
		return nil, errors.New("任务状态错误。")
	}
	// 更新任务状态。
	err = storage.RDB.QueryNaught(ctx, "update mission set status=$1,launch_time=$2,update_time=$2, where id=$3;", common.MissionStatus02, time.Now().UTC(), id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// CancelMission 取消处理任务。
//
// param: ctx 上下文。
// param: id 任务编码。
// return: 任务编码，错误信息。
func (ps *prepareService) CancelMission(ctx *gin.Context, id *string) (*string, error) {
	// 查询分析任务。
	mission, err := storage.RDB.QuerySingle[mission](ctx, "select mission.id,novel.id as novel_code,novel.title as novel_name,mission.status,mission.total_count,mission.embed_count,mission.launch_time,mission.finish_time,mission.update_time,mission.retry_count,mission.message from mission inner join novel on novel.id=mission.novel where mission.id=$1;", id)
	if err != nil {
		return nil, err
	}
	// 验证任务状态。
	if *mission.Status != common.MissionStatus02 && *mission.Status != common.MissionStatus03 && *mission.Status != common.MissionStatus04 && *mission.Status != common.MissionStatus05 && *mission.Status != common.MissionStatus06 {
		return nil, errors.New("任务状态错误。")
	}
	// 更新任务状态。
	err = storage.RDB.QueryNaught(ctx, "update mission set status=$1 where id=$2;", common.MissionStatus08, id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// FilterNovel 查询参考小说。
//
// param: ctx 上下文。
// return: 参考小说，错误信息。
func (ps *prepareService) FilterNovel(ctx *gin.Context) ([]novel, error) {
	result, err := storage.RDB.QueryMatrix[novel](ctx, "select novel.id,novel.origin,origin.type,novel.title,mission.id as mission,mission.status,novel.character_count,novel.chapter_count,account.id as create_code,account.name as create_name,novel.create_time,novel.update_time from novel inner join origin on origin.id=novel.origin inner join account on account.id=novel.create_user inner join mission on mission.novel=novel.id order by novel.create_time desc;")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ModifyNovel 修改参考小说。
//
// param: ctx 上下文。
// param: id 小说编码。
// param: title 小说标题。
// return: 小说编码，错误信息。
func (ps *prepareService) ModifyNovel(ctx *gin.Context, id *string, title *string) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 验证小说状态。
	novel, err := storage.RDB.QuerySingle[novel](ctx, "select novel.id,novel.origin,origin.type,novel.title,mission.id as mission,mission.status,novel.character_count,novel.chapter_count,account.id as create_code,account.name as create_name,novel.create_time,novel.update_time from novel inner join origin on origin.id=novel.origin inner join account on account.id=novel.create_user inner join mission on mission.novel=novel.id where novel.id=$1;", id)
	if err != nil {
		return nil, err
	}
	if *novel.CreateCode != claims.Subject {
		return nil, errors.New("小说不属于此用户。")
	}
	// 修改参考小说。
	err = storage.RDB.QueryNaught(ctx, "update novel set title=$1,update_time=$2 where id=$3;", title, time.Now().UTC(), id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// DeleteNovel 删除参考小说。
//
// param: ctx 上下文。
// param: id 小说编码。
// return: 小说编码，错误信息。
func (ps *prepareService) DeleteNovel(ctx *gin.Context, id *string) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 验证小说状态。
	novel, err := storage.RDB.QuerySingle[novel](ctx, "select novel.id,novel.origin,origin.type,novel.title,mission.id as mission,mission.status,novel.character_count,novel.chapter_count,account.id as create_code,account.name as create_name,novel.create_time,novel.update_time from novel inner join origin on origin.id=novel.origin inner join account on account.id=novel.create_user inner join mission on mission.novel=novel.id where novel.id=$1", id)
	if err != nil {
		return nil, err
	}
	if *novel.CreateCode != claims.Subject {
		return nil, errors.New("小说不属于此用户。")
	}
	if *novel.Status != common.MissionStatus01 && *novel.Status != common.MissionStatus06 && *novel.Status != common.MissionStatus07 && *novel.Status != common.MissionStatus08 {
		return nil, errors.New("任务状态错误。")
	}
	// 删除关系数据。
	var batch storage.Batch
	batch.Append("delete from segment where novel=$1", novel.ID)
	batch.Append("delete from chapter where novel=$1", novel.ID)
	batch.Append("delete from mission where id=$1", novel.Mission)
	batch.Append("delete from novel where id=$1", novel.ID)
	batch.Append("delete from origin where id=$1", novel.Origin)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return nil, err
	}
	// 删除向量数据。
	err = storage.VDB.DeletePoints(ctx, "novel", *novel.ID)
	if err != nil {
		return nil, err
	}
	// 删除归一文件。
	storage.OSS.Delete(ctx, "Normal", *novel.Origin+".md")
	// 删除原始文件。
	storage.OSS.Delete(ctx, "Origin", *novel.Origin+"."+*novel.Type)
	// 返回小说编码。
	return novel.ID, nil
}

// mission 处理任务。
type mission struct {
	// ID 任务编码。
	ID *string `db:"id" json:"id"`
	// NovelCode 小说编码。
	NovelCode *string `db:"novel_code" json:"novel_code"`
	// NovelName 小说名称。
	NovelName *string `db:"novel_name" json:"novel_name"`
	// Status 任务状态。
	Status *string `db:"status" json:"status"`
	// TotalCount 切片总数。
	TotalCount *int32 `db:"total_count" json:"total_count"`
	// EmbedCount 入库数量。
	EmbedCount *int32 `db:"embed_count" json:"embed_count"`
	// LaunchTime 启动时间。
	LaunchTime *time.Time `db:"launch_time" json:"launch_time,omitempty"`
	// FinishTime 完成时间。
	FinishTime *time.Time `db:"finish_time" json:"finish_time,omitempty"`
	// UpdateTime 更新时间。
	UpdateTime *time.Time `db:"update_time" json:"update_time"`
	// RetryCount 重试次数。
	RetryCount *int32 `db:"retry_count" json:"retry_count"`
	// Message 错误消息。
	Message *string `db:"message" json:"message,omitempty"`
}

// novel 参考小说。
type novel struct {
	// ID 小说编码。
	ID *string `db:"id" json:"id"`
	// Origin 原始文件。
	Origin *string `db:"origin" json:"origin"`
	// Type 原始类型。
	Type *string `db:"type" json:"type"`
	// Title 小说名称。
	Title *string `db:"title" json:"title"`
	// Mission 处理任务。
	Mission *string `db:"mission" json:"mission"`
	// Status 任务状态。
	Status *string `db:"status" json:"status"`
	// CharacterCount 字符数量。
	CharacterCount *int32 `db:"character_count" json:"character_count"`
	// ChapterCount 章节数量。
	ChapterCount *int32 `db:"chapter_count" json:"chapter_count"`
	// CreateCode 创建用户编码。
	CreateCode *string `db:"create_code" json:"create_code"`
	// CreateName 创建用户名称。
	CreateName *string `db:"create_name" json:"create_name"`
	// CreateTime 创建时间。
	CreateTime *time.Time `db:"create_time" json:"create_time"`
	// UpdateTime 更新时间。
	UpdateTime *time.Time `db:"update_time" json:"update_time"`
}
