package handler

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"service/internal/server/response"
	"service/internal/service"
)

// maxUploadSize 单次上传大小上限。
const maxUploadSize = 50 << 20

// IngestHandler 文件上传与处理进度接口。
type IngestHandler struct {
	ingest *service.IngestService
}

// NewIngestHandler 创建文件接入接口。
func NewIngestHandler(ingest *service.IngestService) *IngestHandler {
	return &IngestHandler{ingest: ingest}
}

// Register 注册文件接入相关路由。
func (h *IngestHandler) Register(group *gin.RouterGroup) {
	group.POST("/files", h.Upload)
	group.GET("/files", h.List)
	group.GET("/files/:id/progress", h.Progress)
	group.POST("/files/:id/retry", h.Retry)
	group.DELETE("/files/:id", h.Delete)
}

// Upload 上传文件。
func (h *IngestHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.Error(c, http.StatusRequestEntityTooLarge, "文件过大，最大支持 50MB")
			return
		}
		response.Error(c, http.StatusBadRequest, "读取上传文件失败："+err.Error())
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "读取上传内容失败："+err.Error())
		return
	}

	origin, err := h.ingest.Upload(c.Request.Context(), filepath.Base(header.Filename), int64(len(data)), data)
	if errors.Is(err, service.ErrDuplicateFile) {
		response.Error(c, http.StatusConflict, "文件已存在，请勿重复上传")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "上传失败："+err.Error())
		return
	}
	response.OK(c, origin)
}

// List 查询文件列表。
// TODO 调用 service 查询文件及处理状态。
func (h *IngestHandler) List(c *gin.Context) {
	response.NotImplemented(c)
}

// Progress 查询处理进度。
// TODO 从路径参数取 id，返回状态与已处理数量。
func (h *IngestHandler) Progress(c *gin.Context) {
	response.NotImplemented(c)
}

// Retry 重试处理失败的文件。
// TODO 重置进度并重新提交处理任务。
func (h *IngestHandler) Retry(c *gin.Context) {
	response.NotImplemented(c)
}

// Delete 删除文件及其派生数据。
// TODO 删除 MinIO 对象、数据库记录与向量数据。
func (h *IngestHandler) Delete(c *gin.Context) {
	response.NotImplemented(c)
}
