// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/9
// 描述：
// *****************************************************************************

package book

import "database/sql"

type Column struct {
	Field    string
	Field2   string
	Type     string
	Null     string
	Key      string
	Default  sql.NullString
	Extra    string
	Comment  string
	DataName string
	DataType string
	JsonName string
}
