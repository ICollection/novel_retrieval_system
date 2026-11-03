package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"service/internal/config"
	"strconv"
	"sync"

	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	minioMutex    sync.Mutex
	minioInstance *Minio
)

// Minio MinIO 对象存储。
type Minio struct {
	client *minioSDK.Client
}

// CreateMinio 创建 MinIO 对象存储。
func CreateMinio() (*Minio, error) {
	minioMutex.Lock()
	defer minioMutex.Unlock()
	// 已创建时复用。
	if minioInstance != nil {
		return minioInstance, nil
	}
	// 创建数据连接。
	endpoint := net.JoinHostPort(config.Config.MinIO.Endpoint, strconv.Itoa(config.Config.MinIO.Port))
	client, err := minioSDK.New(endpoint, &minioSDK.Options{Creds: credentials.NewStaticV4(config.Config.MinIO.AccessKey, config.Config.MinIO.SecretKey, ""), Secure: config.Config.MinIO.UseSSL, Region: config.Config.MinIO.Region})
	if err != nil {
		return nil, fmt.Errorf("创建 MinIO 客户端失败，详细错误信息：%w", err)
	}
	minioInstance = &Minio{client: client}
	// 返回数据连接。
	return minioInstance, nil
}

// Upload 上传对象。
func (minio *Minio) Upload(ctx context.Context, path string, file string, data []byte, contentType string) error {
	_, err := minio.client.PutObject(ctx, config.Config.MinIO.Bucket, path+"/"+file, bytes.NewReader(data), int64(len(data)), minioSDK.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("上传对象失败，详细错误信息：%w", err)
	}
	return nil
}

// Select 获取对象。
func (minio *Minio) Select(ctx context.Context, path string, file string) ([]byte, error) {
	// 获取对象数据。
	object, err := minio.client.GetObject(ctx, config.Config.MinIO.Bucket, path+"/"+file, minioSDK.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取对象失败，详细错误信息：%w", err)
	}
	defer object.Close()
	// 解析对象数据。
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("读取对象失败，详细错误信息：%w", err)
	}
	// 返回对象数据。
	return data, nil
}

// Delete 删除对象。
func (minio *Minio) Delete(ctx context.Context, object string) error {
	if err := minio.client.RemoveObject(ctx, config.Config.MinIO.Bucket, object, minioSDK.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除对象失败，详细错误信息：%w", err)
	}
	return nil
}
