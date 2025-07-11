// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：配置文件
// *****************************************************************************

package db

const (
	Asc  = "ASC"
	Desc = "DESC"
)

type Config struct {
	Host         string `yaml:"host" json:"host"`
	Port         string `yaml:"port" json:"port"`
	Dbname       string `yaml:"dbname" json:"dbname"`
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
	Charset      string `yaml:"charset" json:"charset"`
	MaxIdleConns int    `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns" json:"maxOpenConns"`
	Driver       string `yaml:"driver" json:"driver"`
	Secret       string `yaml:"secret" json:"secret"` // ciphertext类型字段key
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
