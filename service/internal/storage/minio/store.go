package minio

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"

	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"service/internal/config"
)

// Store MinIO 存储，封装文件对象的上传、下载与删除。
type Store struct {
	client *minioSDK.Client
	bucket string
	region string
}

// New 创建 MinIO 存储。
func New() (*Store, error) {
	cfg := config.Config.MinIO
	endpoint := net.JoinHostPort(cfg.Endpoint, strconv.Itoa(cfg.Port))
	client, err := minioSDK.New(endpoint, &minioSDK.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 MinIO 客户端失败，详细错误信息：%w", err)
	}
	return &Store{client: client, bucket: cfg.Bucket, region: cfg.Region}, nil
}

// Put 上传对象。
func (s *Store) Put(ctx context.Context, object string, data []byte) error {
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}
	_, err := s.client.PutObject(ctx, s.bucket, object, bytes.NewReader(data), int64(len(data)),
		minioSDK.PutObjectOptions{ContentType: "text/plain; charset=utf-8"})
	if err != nil {
		return fmt.Errorf("上传对象失败，详细错误信息：%w", err)
	}
	return nil
}

// Remove 删除对象。
func (s *Store) Remove(ctx context.Context, object string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, object, minioSDK.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除对象失败，详细错误信息：%w", err)
	}
	return nil
}

// ensureBucket 确保存储桶存在。
func (s *Store) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("检查存储桶失败，详细错误信息：%w", err)
	}
	if exists {
		return nil
	}
	if err = s.client.MakeBucket(ctx, s.bucket, minioSDK.MakeBucketOptions{Region: s.region}); err != nil {
		return fmt.Errorf("创建存储桶失败，详细错误信息：%w", err)
	}
	return nil
}
