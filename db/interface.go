// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：方法适配器
// *****************************************************************************

package db

import (
	"gorm.io/gorm"
)

type Adapter interface {
	GetDatabaseName() (string, error)
	GetTableComment(dbName, tableName string) (string, error)
	GetColumns(tableName string) ([]Column, error)

	Book() string                                  // 数据库字典
	ModelFiles(tableNames ...string) (bool, error) // 模型文件

	QueryWhereFindInSet(db *gorm.DB, query TextSlice, column string) *gorm.DB
	JsonExtract(column string, key string) string // 提取json字段做为字段
}
