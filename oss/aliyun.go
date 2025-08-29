// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：阿里云
// *****************************************************************************

package oss

import (
	"context"
	"fmt"
	aliyun "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"log"
	"strings"
	"time"
)

func NewAliYun(config Config) *Api {
	var api = Api{
		Config: config,
	}

	return &api
}

type AliYunAdapter struct {
	*Config
	client *aliyun.Client
}

func NewAliYunAdapter(config *Config) *AliYunAdapter {

	// 加载默认配置并设置凭证提供者和区域
	cfg := aliyun.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(config.Region)
	client := aliyun.NewClient(cfg)

	return &AliYunAdapter{Config: config, client: client}
}

func (s AliYunAdapter) ObjectUrl(objectName string) string {
	if s.Config.Private {

		result, err := s.client.Presign(context.TODO(), &aliyun.GetObjectRequest{
			Bucket: aliyun.Ptr(s.Config.Bucket),
			Key:    aliyun.Ptr(objectName),
		}, aliyun.PresignExpires(10*time.Minute))
		if err != nil {
			return fmt.Sprintf("failed to generate presign URL: %v", err)
		}

		return result.URL
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.Config.Domain, "/"), objectName)
}

func (s AliYunAdapter) UploadSign(objectName string) any {
	result, err := s.client.Presign(context.TODO(), &aliyun.PutObjectRequest{
		Bucket: aliyun.Ptr(s.Config.Bucket),
		Key:    aliyun.Ptr(objectName),
	},
		aliyun.PresignExpires(10*time.Minute),
	)
	if err != nil {
		panic(fmt.Sprintf("aliyun生成PutObject的预签名URL失败:%v", err))
	}
	return result
}

func (s AliYunAdapter) Delete(objectName string) error {
	request := &aliyun.DeleteObjectRequest{
		Bucket: aliyun.Ptr(s.Config.Bucket),
		Key:    aliyun.Ptr(objectName),
	}

	// 执行删除对象的操作并处理结果
	result, err := s.client.DeleteObject(context.TODO(), request)
	if err != nil {
		return err
	}
	if s.Config.Debug {
		// 打印删除对象的结果
		log.Printf("delete object result:%#v\n", result)
	}
	return nil
}

func (s *AliYunAdapter) UploadBase64(objectName string, contentType string, fileBase64 string) {}
