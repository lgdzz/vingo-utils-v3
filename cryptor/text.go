// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package cryptor

import (
	"encoding/base64"
	"github.com/duke-git/lancet/v2/cryptor"
)

// TextEncode 文本加密
func TextEncode(text string, secret []byte) string {
	cipherBytes := cryptor.AesGcmEncrypt([]byte(text), secret)
	return base64.StdEncoding.EncodeToString(cipherBytes)
}

// TextDecode 文本解密
func TextDecode(text string, secret []byte) string {
	cipherBytes, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		panic("decoding text: " + err.Error())
	}
	return string(cryptor.AesGcmDecrypt(cipherBytes, secret))
}

// TextBase64Encode 文本base64编码
func TextBase64Encode(text string) string {
	return cryptor.Base64StdEncode(text)
}

// TextBase64Decode 文本base64解码
func TextBase64Decode(text string) string {
	return cryptor.Base64StdDecode(text)
}
