package storage

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"service/internal/config"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient MinIO 对象存储客户端。
type MinioClient struct {
	client  *minio.Client
	bucket  string
	timeout time.Duration
}

// newMinioClient 创建 MinIO 对象存储连接。
//
// param: cfg 对象存储配置。
// return: 对象存储连接，错误信息。
func newMinioClient(cfg config.Minio) (*MinioClient, error) {
	endpoint := net.JoinHostPort(cfg.Endpoint, strconv.Itoa(cfg.Port))
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""), Secure: cfg.UseSSL, Region: cfg.Region})
	if err != nil {
		slog.Error("[OSS]创建连接失败", "error", err)
		return nil, err
	}
	return &MinioClient{client: client, bucket: cfg.Bucket, timeout: time.Second * time.Duration(cfg.Timeout)}, nil
}

// Select 获取对象。
//
// param: ctx 上下文。
// param: path 对象路径。
// param: file 对象文件。
// return: 对象数据，错误信息。
func (mc *MinioClient) Select(ctx context.Context, path string, file string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()
	// 获取对象数据。
	object, err := mc.client.GetObject(ctx, mc.bucket, path+"/"+file, minio.GetObjectOptions{})
	if err != nil {
		slog.Error("[OSS]读取对象失败", "error", err)
		return nil, err
	}
	defer object.Close()
	// 解析对象数据。
	data, err := io.ReadAll(object)
	if err != nil {
		slog.Error("[OSS]读取对象失败", "error", err)
		return nil, err
	}
	// 返回对象数据。
	return data, nil
}

// Upload 上传对象。
//
// param: ctx 上下文。
// param: path 对象路径。
// param: file 对象文件。
// param: data 对象数据。
// param: contentType 内容类型。
// return: 错误信息。
func (mc *MinioClient) Upload(ctx context.Context, path string, file string, data []byte, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()
	_, err := mc.client.PutObject(ctx, mc.bucket, path+"/"+file, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		slog.Error("[OSS]上传对象失败", "error", err)
		return err
	}
	return nil
}

// Delete 删除对象。
//
// param: ctx 上下文。
// param: path 对象路径。
// param: file 对象文件。
// return: 错误信息。
func (mc *MinioClient) Delete(ctx context.Context, path string, file string) error {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()
	if err := mc.client.RemoveObject(ctx, mc.bucket, path+"/"+file, minio.RemoveObjectOptions{}); err != nil {
		slog.Error("[OSS]删除对象失败", "error", err)
		return err
	}
	return nil
}
