// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：七牛云
// *****************************************************************************

package oss

import (
	"context"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/objects"
	"github.com/qiniu/go-sdk/v7/storagev2/uptoken"
	"time"
)

func NewQiNiu(config Config) *Api {
	var api = Api{
		Config: config,
	}

	return &api
}

type QiNiuAdapter struct {
	*Config

	mac     *credentials.Credentials
	manager *objects.Bucket
}

func NewQiNiuAdapter(config *Config) *QiNiuAdapter {
	return &QiNiuAdapter{Config: config}
}

func (s QiNiuAdapter) ObjectUrl(objectName string) string {
	if s.Config.Private {
		return storage.MakePrivateURL(s.newMac(), s.Domain, objectName, 3600)
	}
	return storage.MakePublicURL(s.Config.Domain, objectName)
}

func (s QiNiuAdapter) UploadSign(objectName string) any {

	var expiry = 1 * time.Hour

	putPolicy, err := uptoken.NewPutPolicy(s.Config.Bucket, time.Now().Add(expiry))
	if err != nil {
		panic(err)
	}
	upToken, err := uptoken.NewSigner(putPolicy, s.newMac()).GetUpToken(context.Background())
	if err != nil {
		panic(err)
	}

	return upToken
}

func (s QiNiuAdapter) Delete(objectName string) error {
	err := s.bucketManager().Object(objectName).Delete().Call(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (s QiNiuAdapter) newMac() *credentials.Credentials {
	if s.mac == nil {
		s.mac = credentials.NewCredentials(s.Config.AccessKey, s.Config.SecretKey)
	}
	return s.mac
}

func (s QiNiuAdapter) bucketManager() *objects.Bucket {
	if s.manager == nil {
		objectsManager := objects.NewObjectsManager(&objects.ObjectsManagerOptions{
			Options: http_client.Options{Credentials: s.newMac()},
		})
		s.manager = objectsManager.Bucket(s.Config.Bucket)
	}
	return s.manager
}

func (s *QiNiuAdapter) UploadBase64(objectName string, contentType string, fileBase64 string) {}
