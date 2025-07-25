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

func NewOSS(config Config) *Api {
	var api *Api
	switch config.Driver {
	case "minio":
		api = NewMinIO(config)
		api.Adapter = NewMinIOAdapter(&config)
	case "qiniu":
		api = NewQiNiu(config)
		api.Adapter = NewQiNiuAdapter(&config)
	case "aliyun":
		api = NewAliYun(config)
		api.Adapter = NewAliYunAdapter(&config)
	default:
		panic("oss driver not found")
	}
	return api
}
