// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package cryptor

import (
	"github.com/duke-git/lancet/v2/cryptor"
)

// Md5 计算文本md5值
func Md5(str string) string {
	return cryptor.Md5String(str)
}

// Md5File 计算文件md5值
// Example: Md5File("./main.go")
func Md5File(filePath string) string {
	v, _ := cryptor.Md5File(filePath)
	return v
}
