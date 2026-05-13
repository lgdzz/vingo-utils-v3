// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：minio
// *****************************************************************************

package oss

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	options := minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	}
	if config.Transport != nil {
		options.Transport = *config.Transport
	}
	client, err := minio.New(config.Endpoint, &options)
	if err != nil {
		panic(fmt.Sprintf("MinIO初始化异常：%v", err.Error()))
	}

	return &MinIOAdapter{Config: config, client: client}
}

func (s MinIOAdapter) ObjectUrl(objectName string) string {
	if strings.HasPrefix(objectName, "http") {
		return objectName
	}
	if s.Config.Private {
		return s.privateUrl(objectName)
	}
	return strings.TrimRight(s.Config.Domain, "/") + "/" + objectName
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
		"domain": s.Config.Domain,
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

func (s MinIOAdapter) UploadBase64(objectName string, contentType string, fileBase64 string) {
	//var fileBase64Array = strings.Split(fileBase64, ",")
	//if len(fileBase64Array) > 1 {
	//	fileBase64 = fileBase64Array[1]
	//}
	//data, err := base64.StdEncoding.DecodeString(fileBase64)
	//if err != nil {
	//	panic(err.Error())
	//}
	//_, err = s.client.PutObject(context.Background(), s.Config.Bucket, objectName, strings.NewReader(string(data)), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	//if err != nil {
	//	panic(err.Error())
	//}

	fileBase64Array := strings.Split(fileBase64, ",")
	if len(fileBase64Array) > 1 {
		fileBase64 = fileBase64Array[1]
	}

	data, err := base64.StdEncoding.DecodeString(fileBase64)
	if err != nil {
		panic(err.Error())
	}

	// 创建 POST Policy
	policy := minio.NewPostPolicy()
	_ = policy.SetBucket(s.Config.Bucket)
	_ = policy.SetKey(objectName)
	_ = policy.SetExpires(time.Now().Add(time.Hour))
	_ = policy.SetContentType(contentType)

	uploadURL, formData, err := s.client.PresignedPostPolicy(context.Background(), policy)
	if err != nil {
		panic(err.Error())
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range formData {
		_ = writer.WriteField(k, v)
	}

	part, err := writer.CreateFormFile("file", objectName)
	if err != nil {
		panic(err.Error())
	}

	_, err = part.Write(data)
	if err != nil {
		panic(err.Error())
	}

	_ = writer.Close()

	req, err := http.NewRequest("POST", uploadURL.String(), body)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		panic("upload failed: " + resp.Status)
	}
}

func (s MinIOAdapter) objectUrl(objectName string) string {
	// 如果有内网地址，则使用内网地址访问（MinIO）
	if s.Config.Intranet != "" {
		return fmt.Sprintf("http://%v/%v/%v", s.Config.Intranet, s.Config.Bucket, objectName)
	}
	return s.ObjectUrl(objectName)
}

func (s MinIOAdapter) GetImageBase64(objectName string, timeout ...int) (string, string) {
	return GetImageBase64(s.ObjectUrl(objectName), timeout...)
}

func (s MinIOAdapter) GetBase64(objectName string, timeout ...int) (string, string) {
	return GetBase64(s.ObjectUrl(objectName), timeout...)
}

func (s MinIOAdapter) Client() any {
	return s.client
}
