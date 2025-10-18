// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/9
// 描述：
// *****************************************************************************

package book

import (
	"database/sql"
	"github.com/duke-git/lancet/v2/strutil"
	"gorm.io/gorm"
)

type TableItem struct {
	Name    string
	Comment string
	Columns []Column
}

type Database struct {
	Name        string
	ReleaseTime string
	Tables      []TableItem
}

// 获取字段顺序
func getFieldOrder(db *gorm.DB, dbName string, tableName string) []string {
	var fields []string

	rows, err := db.Raw(`SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION`, dbName, tableName).Rows()

	if err != nil {
		return fields
	}
	defer rows.Close()

	for rows.Next() {
		var field string
		if err := rows.Scan(&field); err != nil {
			continue
		}
		fields = append(fields, field)
	}

	return fields
}

func getTableColumnsOfMysql(db *gorm.DB, dbName string, tableName string) ([]Column, error) {
	var columns []Column
	rows, err := db.Raw(`SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA, COLUMN_COMMENT FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`, dbName, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var column Column
		if err := rows.Scan(&column.Field, &column.Type, &column.Null, &column.Key, &column.Default, &column.Extra, &column.Comment); err != nil {
			return nil, err
		}

		column.Field2 = strutil.CamelCase(column.Field)
		columns = append(columns, column)
	}

	return columns, nil
}

func getTableColumnsOfPgsql(db *gorm.DB, tableName string) ([]Column, error) {
	var columns []Column

	rows, err := db.Raw(`
		SELECT
			a.attname AS column_name,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			NOT a.attnotnull AS is_nullable,
			CASE WHEN ct.contype = 'p' THEN 'PRI' ELSE '' END AS column_key,
			COALESCE(pg_get_expr(d.adbin, d.adrelid), '') AS column_default,
			col_description(a.attrelid, a.attnum) AS column_comment
		FROM
			pg_attribute a
		JOIN pg_class c ON a.attrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = a.attnum
		LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid AND a.attnum = ANY(ct.conkey)
		WHERE
			a.attnum > 0
			AND NOT a.attisdropped
			AND c.relname = ?
			AND n.nspname = 'public'
		ORDER BY a.attnum
	`, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col Column
		var nullable bool
		var comment sql.NullString

		if err := rows.Scan(&col.Field, &col.Type, &nullable, &col.Key, &col.Default, &comment); err != nil {
			return nil, err
		}

		col.Null = map[bool]string{true: "YES", false: "NO"}[nullable]
		col.Extra = "" // PostgreSQL 没有 extra
		col.Comment = ""
		if comment.Valid {
			col.Comment = comment.String
		}

		col.Field2 = strutil.CamelCase(col.Field)
		columns = append(columns, col)
	}

	return columns, nil
}
