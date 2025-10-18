// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：配置文件
// *****************************************************************************

package oss

type Config struct {
	Endpoint  string `yaml:"endpoint" json:"endpoint"`   // 接口节点
	AccessKey string `yaml:"accessKey" json:"accessKey"` // 接口key
	SecretKey string `yaml:"secretKey" json:"secretKey"` // 接口密钥
	Bucket    string `yaml:"bucket" json:"bucket"`       // 存储桶
	Domain    string `yaml:"domain" json:"domain"`       // 访问域名
	Region    string `yaml:"region" json:"region"`       // 存储区域（阿里云）
	UseSSL    bool   `yaml:"useSSL" json:"useSSL"`       // 是否使用SSL（MinIO）
	Driver    string `yaml:"driver" json:"driver"`       // 驱动：minio|qiniu|aliyun
	Private   bool   `yaml:"private" json:"private"`     // 私有空间
	Debug     bool   `yaml:"debug" json:"debug"`
}

func (s *Config) StringValue(value *string, defaultValue string) {
	if *value == "" {
		*value = defaultValue
	}
}

func (s *Config) IntValue(value *int, defaultValue int) {
	if *value == 0 {
		*value = defaultValue
	}
}
