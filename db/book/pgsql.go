// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/9
// 描述：
// *****************************************************************************

package book

import (
	"bytes"
	"database/sql"
	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/gorm"
	"html/template"
	"time"
)

func BuildPgsqlBook(db *gorm.DB) string {
	var tables []TableItem
	var dbName string

	// PostgreSQL 获取当前数据库名
	err := db.Raw("SELECT current_database()").Row().Scan(&dbName)
	if err != nil {
		panic(err)
	}

	// PostgreSQL 获取所有表名及注释
	rows, err := db.Raw(`
		SELECT c.relname AS table_name, obj_description(c.oid) AS comment
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r' AND n.nspname = 'public'
	`).Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var comment sql.NullString

		if err := rows.Scan(&tableName, &comment); err != nil {
			panic(err)
		}

		tableComment := ""
		if comment.Valid {
			tableComment = comment.String
		}

		// 获取列信息
		columns, err := getTableColumnsOfPgsql(db, tableName)
		if err != nil {
			panic(err)
		}

		// 字段排序
		fieldOrder := getFieldOrder(db, "public", tableName)
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

	_ = slice.SortByField(tables, "Name", "asc")

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
