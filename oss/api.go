// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：
// *****************************************************************************

package oss

type Api struct {
	Config Config
	Adapter
}

const (
	MinioOss  = "minio"
	QiniuOss  = "qiniu"
	AliyunOss = "aliyun"
)

func NewOSS(config Config) *Api {
	var api *Api
	switch config.Driver {
	case MinioOss:
		api = NewMinIO(config)
		api.Adapter = NewMinIOAdapter(&config)
	case QiniuOss:
		api = NewQiNiu(config)
		api.Adapter = NewQiNiuAdapter(&config)
	case AliyunOss:
		api = NewAliYun(config)
		api.Adapter = NewAliYunAdapter(&config)
	default:
		panic("oss driver not found")
	}
	return api
}
