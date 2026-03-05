package ddl

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DBType string

const (
	MySQL DBType = "mysql"
	PGSQL DBType = "pgsql"
)

// -------------------------
// AST 解析字段注释
// -------------------------
func ParseStructFieldComments(filePath string, structName string) (map[string]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != structName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}
				fieldName := field.Names[0].Name
				if field.Comment != nil {
					comment := strings.TrimSpace(field.Comment.Text())
					comment = strings.TrimPrefix(comment, "//")
					comment = strings.TrimSpace(comment)
					result[fieldName] = comment
				}
			}
		}
	}
	return result, nil
}

// -------------------------
// 生成 CREATE TABLE SQL
// -------------------------
func GenerateCreateTableSQLList(model interface{}, filePath string, dbType DBType, tableName string, tableComment string) ([]string, error) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be struct")
	}

	fieldComments, err := ParseStructFieldComments(filePath, t.Name())
	if err != nil {
		return nil, err
	}

	var columns []string
	var sqlList []string
	seqName := fmt.Sprintf("%s_id_seq", tableName)

	// -------------------------
	// 表/序列 DROP
	// -------------------------
	if dbType == PGSQL {
		sqlList = append(sqlList, fmt.Sprintf(`
DO $$
BEGIN
   IF EXISTS (SELECT 1 FROM pg_class WHERE relkind='r' AND relname='%s') THEN
       DROP TABLE %s CASCADE;
   END IF;
END
$$;`, tableName, tableName))
		sqlList = append(sqlList, fmt.Sprintf(`
DO $$
BEGIN
   IF EXISTS (SELECT 1 FROM pg_class WHERE relkind='S' AND relname='%s') THEN
       DROP SEQUENCE %s;
   END IF;
END
$$;`, seqName, seqName))
		sqlList = append(sqlList, fmt.Sprintf("CREATE SEQUENCE %s;", seqName))
	}
	if dbType == MySQL {
		sqlList = append(sqlList, fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", tableName))
	}

	// -------------------------
	// 列定义
	// -------------------------
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("gorm") == "-" {
			continue
		}

		columnName, isPK, _ := parseGormTag(field)
		if columnName == "" {
			columnName = toSnakeCase(field.Name)
		}

		sqlType := mapGoType(field.Type, field.Tag, dbType)
		if sqlType == "" {
			continue
		}

		comment := fieldComments[field.Name]
		colDef := fmt.Sprintf("  %s %s", wrapName(columnName, dbType), sqlType)

		isDeletedAt := field.Type == reflect.TypeOf(gorm.DeletedAt{})
		isPointer := field.Type.Kind() == reflect.Ptr

		// NULL / NOT NULL
		if !isDeletedAt {
			if isPointer {
				colDef += " NULL"
			} else {
				colDef += " NOT NULL"
			}
		}

		// 主键自增
		if isPK {
			if dbType == PGSQL && (sqlType == "INTEGER" || sqlType == "BIGINT") {
				colDef += fmt.Sprintf(" DEFAULT nextval('%s')", seqName)
			}
			if dbType == MySQL && (sqlType == "INTEGER" || sqlType == "BIGINT") {
				colDef += " AUTO_INCREMENT"
			}
		}

		// MySQL 字段备注
		if dbType == MySQL && comment != "" {
			colDef += fmt.Sprintf(" COMMENT '%s'", escape(comment))
		}

		columns = append(columns, colDef)
	}

	// -------------------------
	// CREATE TABLE
	// -------------------------
	createSQL := fmt.Sprintf("CREATE TABLE %s (\n%s\n);", wrapName(tableName, dbType), strings.Join(columns, ",\n"))
	sqlList = append(sqlList, createSQL)

	// -------------------------
	// 表注释
	// -------------------------
	if tableComment != "" {
		if dbType == MySQL {
			sqlList[len(sqlList)-1] = strings.TrimSuffix(sqlList[len(sqlList)-1], ";") + fmt.Sprintf(" COMMENT='%s';", escape(tableComment))
		} else {
			sqlList = append(sqlList, fmt.Sprintf("COMMENT ON TABLE %s IS '%s';", wrapName(tableName, dbType), escape(tableComment)))
		}
	}

	// -------------------------
	// PG 字段注释
	// -------------------------
	if dbType == PGSQL {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Tag.Get("gorm") == "-" {
				continue
			}
			columnName, _, _ := parseGormTag(field)
			if columnName == "" {
				columnName = toSnakeCase(field.Name)
			}
			comment := fieldComments[field.Name]
			if comment != "" {
				sqlList = append(sqlList, fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';", wrapName(tableName, dbType), wrapName(columnName, dbType), escape(comment)))
			}
		}
	}

	// -------------------------
	// 生成索引与约束
	// -------------------------
	type IndexInfo struct {
		Name               string
		Columns            []string
		Unique             bool
		IsPK               bool
		IsUniqueConstraint bool
	}

	indexes := map[string]*IndexInfo{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("gorm")
		if tag == "-" {
			continue
		}

		columnName, isPK, _ := parseGormTag(field)
		if columnName == "" {
			columnName = toSnakeCase(field.Name)
		}

		if isPK {
			idxName := fmt.Sprintf("%s_pk", tableName)
			indexes[idxName] = &IndexInfo{
				Name:    idxName,
				Columns: []string{columnName},
				Unique:  true,
				IsPK:    true,
			}
		}

		parts := strings.Split(tag, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "index" {
				idxName := fmt.Sprintf("%s_%s_idx", tableName, columnName)
				indexes[idxName] = &IndexInfo{
					Name:    idxName,
					Columns: []string{columnName},
					Unique:  false,
				}
			}
			if p == "unique" || p == "uniqueIndex" {
				idxName := fmt.Sprintf("%s_%s_uindex", tableName, columnName)
				indexes[idxName] = &IndexInfo{
					Name:               idxName,
					Columns:            []string{columnName},
					Unique:             true,
					IsUniqueConstraint: true,
				}
			}
			if strings.HasPrefix(p, "index:") || strings.HasPrefix(p, "uniqueIndex:") {
				parts2 := strings.Split(p, ":")
				if len(parts2) == 2 {
					idxName := parts2[1]
					isUnique := strings.HasPrefix(p, "uniqueIndex:")
					indexes[idxName] = &IndexInfo{
						Name:    idxName,
						Columns: []string{columnName},
						Unique:  isUnique,
					}
				}
			}
		}
	}

	// 生成 SQL
	for _, idx := range indexes {
		colStr := make([]string, len(idx.Columns))
		for i, c := range idx.Columns {
			colStr[i] = wrapName(c, dbType)
		}

		if dbType == PGSQL {
			if idx.IsPK {
				sqlList = append(sqlList, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s);", wrapName(tableName, dbType), idx.Name, strings.Join(colStr, ",")))
				continue
			}
			if idx.IsUniqueConstraint {
				sqlList = append(sqlList, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s);", wrapName(tableName, dbType), idx.Name, strings.Join(colStr, ",")))
				continue
			}
			if idx.Unique {
				sqlList = append(sqlList, fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);", idx.Name, wrapName(tableName, dbType), strings.Join(colStr, ",")))
			} else {
				sqlList = append(sqlList, fmt.Sprintf("CREATE INDEX %s ON %s (%s);", idx.Name, wrapName(tableName, dbType), strings.Join(colStr, ",")))
			}
		} else {
			// MySQL 主键在列上已经定义 AUTO_INCREMENT + PRIMARY KEY
			if !idx.IsPK {
				if idx.Unique {
					sqlList = append(sqlList, fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s` (%s);", idx.Name, tableName, strings.Join(idx.Columns, ",")))
				} else {
					sqlList = append(sqlList, fmt.Sprintf("CREATE INDEX `%s` ON `%s` (%s);", idx.Name, tableName, strings.Join(idx.Columns, ",")))
				}
			}
		}
	}

	return sqlList, nil
}

// -------------------------
// 类型映射
// -------------------------
func mapGoType(t reflect.Type, tag reflect.StructTag, dbType DBType) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	gormTag := tag.Get("gorm")

	// JSON 序列化类型
	if strings.Contains(gormTag, "serializer:json") {
		return "VARCHAR(2000)"
	}

	typeName := t.String()

	// -------------------------
	// 自定义类型特殊处理
	// -------------------------
	switch typeName {

	case "moment.LocalTime":
		if dbType == MySQL {
			return "DATETIME"
		}
		return "TIMESTAMP"

	case "ctype.Bool":
		if dbType == MySQL {
			return "TINYINT"
		}
		return "BOOLEAN"

	case "ctype.Money":
		// 金额类型，保留两位小数
		if dbType == MySQL {
			return "DECIMAL(18,2)"
		}
		return "NUMERIC(18,2)"

	case "ctype.Text":
		return "TEXT"
	}

	// -------------------------
	// 标准类型
	// -------------------------
	switch t.Kind() {

	case reflect.Int, reflect.Int32:
		return "INTEGER"

	case reflect.Int64:
		return "BIGINT"

	case reflect.String:
		return "VARCHAR(255)"

	case reflect.Bool:
		if dbType == MySQL {
			return "TINYINT"
		}
		return "BOOLEAN"

	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			if dbType == MySQL {
				return "DATETIME"
			}
			return "TIMESTAMP"
		}

		if t == reflect.TypeOf(gorm.DeletedAt{}) {
			if dbType == MySQL {
				return "DATETIME"
			}
			return "TIMESTAMP"
		}

		// 其他 struct 统一 varchar
		return "VARCHAR(2000)"
	}

	// 默认 fallback
	return "VARCHAR(255)"
}

// -------------------------
// 工具函数
// -------------------------
func parseGormTag(field reflect.StructField) (column string, isPK bool, skip bool) {
	tag := field.Tag.Get("gorm")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if strings.HasPrefix(p, "column:") {
			column = strings.TrimPrefix(p, "column:")
		}
		if p == "primaryKey" {
			isPK = true
		}
	}
	return
}

func wrapName(name string, dbType DBType) string {
	if dbType == MySQL {
		return "`" + name + "`"
	}
	return `"` + name + `"`
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func escape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func CreateTable(tx *gorm.DB, model interface{}, tableName string, tableComment string, filePath string, dbType string) {
	sqlList, err := GenerateCreateTableSQLList(
		model,
		filePath,
		DBType(dbType),
		tableName,
		tableComment,
	)
	if err != nil {
		panic(err)
	}

	for _, stmt := range sqlList {
		if err := tx.Exec(stmt).Error; err != nil {
			panic(err)
		}
	}
}
