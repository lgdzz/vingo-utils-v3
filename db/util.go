// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/9/16
// 描述：
// *****************************************************************************

package db

import "strings"

func quoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
