// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：密文字段
// *****************************************************************************

package ctype

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/cryptor"
)

var Secret []byte // 请确保为 16/24/32 字节的安全 key

type Ciphertext string

// =================== GORM 接口 ===================

func (s Ciphertext) Value() (driver.Value, error) {
	if s == "" {
		return "", nil
	}
	cipherBytes := cryptor.AesGcmEncrypt([]byte(s), Secret)
	cipherText := base64.StdEncoding.EncodeToString(cipherBytes)
	return []byte(cipherText), nil
}

func (s *Ciphertext) Scan(value any) error {
	if value == nil {
		*s = ""
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		panic(fmt.Sprintf("Ciphertext.Scan 类型不支持: %T", value))
	}

	if str == "" {
		*s = ""
		return nil
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		*s = ""
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			*s = ""
		}
	}()

	plainBytes := cryptor.AesGcmDecrypt(cipherBytes, Secret)
	*s = Ciphertext(plainBytes)
	return nil
}

// =================== JSON 接口 ===================

func (s Ciphertext) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s)) // 输出明文
}

func (s *Ciphertext) UnmarshalJSON(b []byte) error {
	var val string
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	*s = Ciphertext(val) // 接收明文
	return nil
}

func (s Ciphertext) String() string {
	return string(s)
}
