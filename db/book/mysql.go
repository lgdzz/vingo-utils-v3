// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/9
// 描述：
// *****************************************************************************

package book

import (
	"bytes"
	"gorm.io/gorm"
	"html/template"
	"time"
)

func BuildMysqlBook(db *gorm.DB) string {
	var tables []TableItem
	var dbName string
	err := db.Raw("SELECT DATABASE()").Row().Scan(&dbName)
	if err != nil {
		panic(err)
	}

	// 查询所有表的信息
	rows, err := db.Raw(`SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?`, dbName).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, tableComment string
		if err = rows.Scan(&tableName, &tableComment); err != nil {
			panic(err)
		}

		// 查询每张表的列信息并按字段顺序排序
		columns, err := getTableColumnsOfMysql(db, dbName, tableName)
		if err != nil {
			panic(err)
		}

		// 获取字段顺序
		fieldOrder := getFieldOrder(db, dbName, tableName)

		// 根据字段顺序排序
		sortedColumns := make([]Column, len(columns))
		for i, field := range fieldOrder {
			for _, col := range columns {
				if col.Field == field {
					sortedColumns[i] = col
					break
				}
			}
		}

		tables = append(tables, TableItem{
			Name:    tableName,
			Comment: tableComment,
			Columns: sortedColumns,
		})
	}

	// 构造 Database 对象
	database := Database{
		Name:        dbName,
		Tables:      tables,
		ReleaseTime: time.Now().Format("2006年01月02日"),
	}

	// 渲染模板到 bytes.Buffer
	var buf bytes.Buffer
	t, err := template.New("tpl").Parse(BookTpl)
	if err != nil {
		panic(err)
	}

	if err := t.Execute(&buf, database); err != nil {
		panic(err)
	}

	// 返回渲染结果的字符串
	return buf.String()
}
