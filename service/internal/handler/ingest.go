package handler

import (
	"net/http"

	"service/internal/server/response"
)

// Upload 上传文件。
// TODO 解析 multipart 表单，校验后交给 service 保存并创建处理任务。
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	response.NotImplemented(w)
}

// List 查询文件列表。
// TODO 调用 service 查询文件及处理状态。
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response.NotImplemented(w)
}

// Progress 查询处理进度。
// TODO 从路径参数取 id，返回状态与已处理数量。
func (h *Handler) Progress(w http.ResponseWriter, r *http.Request) {
	response.NotImplemented(w)
}

// Retry 重试处理失败的文件。
// TODO 重置进度并重新提交处理任务。
func (h *Handler) Retry(w http.ResponseWriter, r *http.Request) {
	response.NotImplemented(w)
}

// Delete 删除文件及其派生数据。
// TODO 删除 MinIO 对象、数据库记录与向量数据。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	response.NotImplemented(w)
}
