package storage

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"service/internal/config"
	"strconv"
	"time"

	minios "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Minio MinIO 对象存储连接。
type minio struct {
	client *minios.Client
}

// createMinio 创建 MinIO 对象存储连接。
func createMinio() (*minio, error) {
	endpoint := net.JoinHostPort(config.Config.MinIO.Endpoint, strconv.Itoa(config.Config.MinIO.Port))
	client, err := minios.New(endpoint, &minios.Options{Creds: credentials.NewStaticV4(config.Config.MinIO.AccessKey, config.Config.MinIO.SecretKey, ""), Secure: config.Config.MinIO.UseSSL, Region: config.Config.MinIO.Region})
	if err != nil {
		log.Println("[OSS]创建连接失败，详细错误信息：", err)
		return nil, err
	}
	return &minio{client: client}, nil
}

// Upload 上传对象。
func (minio *minio) Upload(ctx context.Context, path string, file string, data []byte, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.MinIO.Timeout))
	defer cancel()
	_, err := minio.client.PutObject(ctx, config.Config.MinIO.Bucket, path+"/"+file, bytes.NewReader(data), int64(len(data)), minios.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Println("[OSS]上传对象失败，详细错误信息：", err)
		return err
	}
	return nil
}

// Select 获取对象。
func (minio *minio) Select(ctx context.Context, path string, file string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.MinIO.Timeout))
	defer cancel()
	// 获取对象数据。
	object, err := minio.client.GetObject(ctx, config.Config.MinIO.Bucket, path+"/"+file, minios.GetObjectOptions{})
	if err != nil {
		log.Println("[OSS]读取对象失败，详细错误信息：", err)
		return nil, err
	}
	defer object.Close()
	// 解析对象数据。
	data, err := io.ReadAll(object)
	if err != nil {
		log.Println("[OSS]读取对象失败，详细错误信息：", err)
		return nil, err
	}
	// 返回对象数据。
	return data, nil
}

// Delete 删除对象。
func (minio *minio) Delete(ctx context.Context, path string, file string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.MinIO.Timeout))
	defer cancel()
	if err := minio.client.RemoveObject(ctx, config.Config.MinIO.Bucket, path+"/"+file, minios.RemoveObjectOptions{}); err != nil {
		log.Println("[OSS]删除对象失败，详细错误信息：", err)
		return err
	}
	return nil
}
