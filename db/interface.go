// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：方法适配器
// *****************************************************************************

package db

import "gorm.io/gorm"

type Adapter interface {
	QueryWhereFindInSet(db *gorm.DB, query TextSlice, column string) *gorm.DB

	Book() string // 数据库字典
}
