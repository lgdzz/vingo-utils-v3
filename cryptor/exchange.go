// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/11/13
// 描述：
// *****************************************************************************

package cryptor

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// 反向两两交换，= 不参与
func unswapPairsWithPadding(s string) string {
	// 分离末尾 = 号
	padding := ""
	for strings.HasSuffix(s, "=") {
		s = s[:len(s)-1]
		padding += "="
	}

	r := []rune(s)
	n := len(r)
	for i := 0; i+1 < n; i += 2 {
		r[i], r[i+1] = r[i+1], r[i]
	}

	return string(r) + padding
}

// 解密函数
func unswapBase64Pairs(enc string, times int) (string, error) {
	res := enc

	for t := 0; t < times; t++ {
		// 先逆转两两交换，保留末尾 =
		res = unswapPairsWithPadding(res)

		// Base64 解码
		decoded, err := base64.StdEncoding.DecodeString(res)
		if err != nil {
			return "", fmt.Errorf("base64 decode failed at loop %d: %v", t, err)
		}
		res = string(decoded)
	}

	// 去掉前缀：第一个字符数字表示随机字符长度
	if len(res) < 1 {
		return "", fmt.Errorf("invalid encoded string")
	}
	first := res[0]
	if first < '1' || first > '9' {
		return "", fmt.Errorf("invalid prefix first character")
	}
	n := int(first - '0')
	if len(res) < 1+n {
		return "", fmt.Errorf("encoded string too short")
	}
	original := res[1+n:] // 去掉第一个数字 + n 个随机字符

	return original, nil
}

func LoginDecode(text string, times int) string {
	original, err := unswapBase64Pairs(text, times)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return original
}
