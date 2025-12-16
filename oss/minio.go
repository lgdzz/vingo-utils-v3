// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：minio
// *****************************************************************************

package oss

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strings"
	"time"
)

func NewMinIO(config Config) *Api {
	var api = Api{
		Config: config,
	}

	return &api
}

type MinIOAdapter struct {
	*Config
	client *minio.Client
}

func NewMinIOAdapter(config *Config) *MinIOAdapter {

	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("MinIO初始化异常：%v", err.Error()))
	}

	return &MinIOAdapter{Config: config, client: client}
}

func (s MinIOAdapter) ObjectUrl(objectName string) string {
	if s.Config.Private {
		return s.privateUrl(objectName)
	}
	return strings.TrimRight(s.Config.Domain, "/") + "/" + s.Config.Bucket + "/" + objectName
}

func (s MinIOAdapter) UploadSign(objectName string) any {
	policy := minio.NewPostPolicy()
	if err := policy.SetExpires(time.Now().Add(time.Minute * 10)); err != nil {
		panic(err.Error())
	}
	if err := policy.SetKey(objectName); err != nil {
		panic(err.Error())
	}
	if err := policy.SetBucket(s.Config.Bucket); err != nil {
		panic(err.Error())
	}
	url, formData, err := s.client.PresignedPostPolicy(context.Background(), policy)

	if err != nil {
		panic(err.Error())
	}
	return map[string]any{
		"policy": formData,
		"url":    url.String(),
	}
}

func (s MinIOAdapter) privateUrl(objectName string) string {
	expire := time.Hour * 2
	url, err := s.client.PresignedGetObject(context.Background(), s.Config.Bucket, objectName, expire, nil)
	if err != nil {
		panic(err.Error())
	}
	return url.String()
}

func (s MinIOAdapter) Delete(objectName string) error {
	err := s.client.RemoveObject(context.Background(), s.Config.Bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *MinIOAdapter) UploadBase64(objectName string, contentType string, fileBase64 string) {
	var fileBase64Array = strings.Split(fileBase64, ",")
	if len(fileBase64Array) > 1 {
		fileBase64 = fileBase64Array[1]
	}
	data, err := base64.StdEncoding.DecodeString(fileBase64)
	if err != nil {
		panic(err.Error())
	}
	_, err = s.client.PutObject(context.Background(), s.Config.Bucket, objectName, strings.NewReader(string(data)), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		panic(err.Error())
	}
}

func (s *MinIOAdapter) Client() any {
	return s.client
}
