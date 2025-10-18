// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/11
// 描述：
// *****************************************************************************

package redis

type Config struct {
	Host         string `yaml:"host" json:"host"`
	Port         string `yaml:"port" json:"port"`
	Select       int    `yaml:"select" json:"select"`
	Password     string `yaml:"password" json:"password"`
	PoolSize     int    `yaml:"poolSize" json:"poolSize"`
	MinIdleConns int    `yaml:"minIdleConns" json:"minIdleConns"`
	Prefix       string `yaml:"prefix" json:"prefix"`
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
