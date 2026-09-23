// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/9/16
// 描述：openGauss数据库适配器
// 文档：https://docs.opengauss.org/zh/
// 说明：
// 1. 基于 openGauss 原生系统目录实现元数据读取
// 2. 不依赖 Dolphin/B-compatible 扩展
// 3. Schema 全程限定，避免不同 Schema 同名对象产生歧义
// 4. 主键一次查询，避免字段循环 N+1 SQL
// 5. DDL 使用 openGauss 原生 pg_get_tabledef()
// 6. 尽量兼容 PostgreSQL 风格 SQL
// sql_compatibility
// | 值    | 兼容对象       | 大致用途          |
// | ---- | ---------- | ------------- |
// | `A`  | Oracle     | Oracle 兼容     |
// | `B`  | MySQL      | MySQL 兼容      |
// | `C`  | Teradata   | Teradata 兼容   |
// | `PG` | PostgreSQL | PostgreSQL 兼容 |
// 创建test数据库使用PG兼容模式
// CREATE DATABASE test WITH dbcompatibility = 'PG';
// *****************************************************************************

package db

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/lgdzz/vingo-utils-v3/db/book"
	"github.com/lgdzz/vingo-utils-v3/db/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// -----------------------------------------------------------------------------
// Connection
// -----------------------------------------------------------------------------

func NewOpenGauss(config Config) *Api {
	config.StringValue(&config.Host, "127.0.0.1")
	config.StringValue(&config.Port, "5432")
	config.StringValue(&config.Username, "omm")
	config.StringValue(&config.Password, "")
	config.StringValue(&config.Charset, "utf8")
	config.StringValue(&config.Schema, "public")
	config.StringValue(&config.Mode, "PG")
	config.IntValue(&config.ConnectTimeout, 5)
	config.IntValue(&config.MaxIdleConns, 10)
	config.IntValue(&config.MaxOpenConns, 100)

	var dbApi = Api{
		Config: config,
	}

	// openGauss 使用 PostgreSQL wire protocol，
	// 因此继续使用 gorm.io/driver/postgres 是正确的。
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s "+
			"sslmode=disable search_path=%s connect_timeout=%d TimeZone=Asia/Shanghai",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Dbname,
		quoteSearchPath(config.Schema),
		config.ConnectTimeout,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,

		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),

		NowFunc: func() time.Time {
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				return time.Now()
			}
			return time.Now().In(loc)
		},
	})

	if err != nil {
		panic("Error to openGauss connection, err: " + err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("Error to get sql.DB, err: " + err.Error())
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if config.InitAfter != nil {
		config.InitAfter(db)
	}

	dbApi.DB = db

	return &dbApi
}

// -----------------------------------------------------------------------------
// Adapter
// -----------------------------------------------------------------------------

type OpenGaussAdapter struct {
	db     *gorm.DB
	config *Config
}

func NewOpenGaussAdapter(db *gorm.DB, config *Config) *OpenGaussAdapter {
	return &OpenGaussAdapter{
		db:     db,
		config: config,
	}
}

// -----------------------------------------------------------------------------
// Database
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) GetDatabases() ([]DatabaseInfo, error) {
	var databases = make([]DatabaseInfo, 0)

	err := s.db.Raw(`
SELECT
d.datname AS name,
pg_encoding_to_char(d.encoding) AS charset,
d.datcollate AS collation,
pg_database_size(d.datname) AS size
FROM pg_database d
WHERE d.datallowconn = true
ORDER BY d.datname
`).Scan(&databases).Error

	return databases, err
}

func (s *OpenGaussAdapter) GetDatabaseName() (string, error) {
	var dbName string

	err := s.db.Raw(`
SELECT current_database()
`).Scan(&dbName).Error

	return dbName, err
}

// -----------------------------------------------------------------------------
// Tables
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) GetTables() ([]TableInfo, error) {
	var tables = make([]TableInfo, 0)

	err := s.db.Raw(`
SELECT
c.relname AS name,
n.nspname AS "schema",

CASE
WHEN c.relkind = 'r' THEN 'BASE TABLE'
WHEN c.relkind = 'p' THEN 'PARTITIONED TABLE'
WHEN c.relkind = 'v' THEN 'VIEW'
WHEN c.relkind = 'm' THEN 'MATERIALIZED VIEW'
WHEN c.relkind = 'f' THEN 'FOREIGN TABLE'
ELSE c.relkind::text
END AS type,

COALESCE(
obj_description(c.oid, 'pg_class'),
''
) AS comment,

COALESCE(
c.reltuples::bigint,
0
) AS "rows",

COALESCE(
pg_total_relation_size(c.oid),
0
) AS size,

'' AS charset,
'' AS collation

FROM pg_class c

INNER JOIN pg_namespace n
ON n.oid = c.relnamespace

WHERE n.nspname = current_schema()
AND c.relkind IN (
'r',
'p',
'f'
)

ORDER BY c.relname
`).Scan(&tables).Error

	return tables, err
}

// -----------------------------------------------------------------------------
// Columns
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) GetColumns(tableName string) ([]Column, error) {
	var columns = make([]Column, 0)

	query := `
SELECT
a.attname AS field,

format_type(
a.atttypid,
a.atttypmod
) AS type,

COALESCE(
col_description(
a.attrelid,
a.attnum
),
''
) AS comment,

CASE
WHEN a.attnotnull THEN 'NO'
ELSE 'YES'
END AS null,

CASE
WHEN EXISTS (
SELECT 1
FROM pg_index i
WHERE i.indrelid = a.attrelid
AND i.indisprimary
AND a.attnum = ANY(i.indkey)
)
THEN true
ELSE false
END AS is_pk

FROM pg_attribute a

INNER JOIN pg_class c
ON c.oid = a.attrelid

INNER JOIN pg_namespace n
ON n.oid = c.relnamespace

WHERE c.relname = ?
AND n.nspname = current_schema()
AND a.attnum > 0
AND NOT a.attisdropped

ORDER BY a.attnum
`

	err := s.db.Raw(query, tableName).Scan(&columns).Error
	if err != nil {
		return columns, err
	}

	columns = slice.Map(
		columns,
		func(index int, item Column) Column {
			item.BusinessType = openGaussBusinessType(item.Type)
			return item
		},
	)

	return columns, nil
}

// -----------------------------------------------------------------------------
// Table comment
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) GetTableComment(
	dbName string,
	tableName string,
) (string, error) {
	var tableComment sql.NullString

	query := `
SELECT
obj_description(c.oid, 'pg_class')
FROM pg_class c
INNER JOIN pg_namespace n
ON n.oid = c.relnamespace
WHERE c.relname = ?
AND n.nspname = current_schema()
LIMIT 1
`

	err := s.db.Raw(
		query,
		tableName,
	).Scan(&tableComment).Error

	if err != nil {
		return "", err
	}

	if tableComment.Valid {
		return tableComment.String, nil
	}

	return "", nil
}

// -----------------------------------------------------------------------------
// Table DDL
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) GetTableDDL(table string) (string, error) {
	// openGauss 原生提供 pg_get_tabledef。
	//
	// 与 PostgreSQL 不同：
	// openGauss 的 pg_get_tabledef 可以直接重建：
	// - CREATE TABLE
	// - INDEX
	// - COMMENT
	// - 部分存储/分布属性
	//
	// 因此生产环境优先使用该函数。

	var ddl string

	err := s.db.Raw(`
SELECT pg_get_tabledef(?)
`, table).Scan(&ddl).Error

	if err != nil {
		return "", fmt.Errorf(
			"获取 openGauss 表 DDL 失败，table=%s: %w",
			table,
			err,
		)
	}

	return ddl, nil
}

// -----------------------------------------------------------------------------
// Book
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) Book() string {
	return book.BuildPgsqlBook(s.db)
}

// -----------------------------------------------------------------------------
// Model
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) ModelFiles(
	tableNames ...string,
) (bool, error) {
	if err := os.MkdirAll("model", 0777); err != nil {
		return false, fmt.Errorf(
			"创建 model 目录失败: %w",
			err,
		)
	}

	for _, tableName := range tableNames {
		success, err := s.modelFile(tableName)
		if err != nil {
			fmt.Printf(
				"生成表 [%s] 模型失败：%v\n",
				tableName,
				err,
			)

			return false, err
		}

		if !success {
			return false, fmt.Errorf(
				"生成表 [%s] 模型失败",
				tableName,
			)
		}
	}

	return true, nil
}

func (s *OpenGaussAdapter) modelFile(
	tableName string,
) (success bool, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(
				"生成模型发生 panic: %v",
				r,
			)
			success = false
		}
	}()

	modelPath := filepath.Join(
		"model",
		tableName+".go",
	)

	// -------------------------------------------------------------------------
	// database
	// -------------------------------------------------------------------------

	dbName, err := s.GetDatabaseName()
	if err != nil {
		return false, fmt.Errorf(
			"获取数据库名失败: %w",
			err,
		)
	}

	// -------------------------------------------------------------------------
	// comment
	// -------------------------------------------------------------------------

	tableComment, err := s.GetTableComment(
		dbName,
		tableName,
	)
	if err != nil {
		return false, fmt.Errorf(
			"获取表注释失败: %w",
			err,
		)
	}

	// -------------------------------------------------------------------------
	// columns
	// -------------------------------------------------------------------------

	columns, err := s.GetColumns(tableName)
	if err != nil {
		return false, fmt.Errorf(
			"获取字段失败: %w",
			err,
		)
	}

	// -------------------------------------------------------------------------
	// primary key
	// -------------------------------------------------------------------------

	primaryKeys, err := s.getPrimaryKeys(tableName)
	if err != nil {
		return false, fmt.Errorf(
			"获取主键失败: %w",
			err,
		)
	}

	columns = slice.Map(
		columns,
		func(index int, col Column) Column {

			col.DataType = openGaussGoType(
				col.Field,
				col.Type,
			)

			col.JsonName = strutil.CamelCase(
				col.Field,
			)

			col.DataName = strutil.UpperFirst(
				col.JsonName,
			)

			if primaryKeys[col.Field] {
				col.Key = "PRI"
			}

			return col
		},
	)

	// -------------------------------------------------------------------------
	// template
	// -------------------------------------------------------------------------

	tmpl, err := template.
		New("tpl").
		Option("missingkey=zero").
		Parse(model.ModelTpl)

	if err != nil {
		return false, fmt.Errorf(
			"解析模板失败: %w",
			err,
		)
	}

	outputFile, err := os.Create(modelPath)
	if err != nil {
		return false, fmt.Errorf(
			"创建文件失败: %w",
			err,
		)
	}

	defer outputFile.Close()

	err = tmpl.Execute(
		outputFile,
		Table{
			TableName: tableName,
			ModelName: strutil.UpperFirst(
				strutil.CamelCase(tableName),
			),
			TableComment: tableComment,
			TableColumns: columns,
			Date: time.Now().Format(
				"2006/01/02",
			),
		},
	)

	if err != nil {
		return false, fmt.Errorf(
			"渲染模板失败: %w",
			err,
		)
	}

	fmt.Printf(
		"✅ 成功生成模型文件: %s\n",
		modelPath,
	)

	return true, nil
}

// -----------------------------------------------------------------------------
// Primary keys
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) getPrimaryKeys(
	tableName string,
) (map[string]bool, error) {

	result := make(map[string]bool)

	query := `
SELECT
a.attname
FROM pg_index i

INNER JOIN pg_class c
ON c.oid = i.indrelid

INNER JOIN pg_namespace n
ON n.oid = c.relnamespace

INNER JOIN pg_attribute a
ON a.attrelid = i.indrelid
AND a.attnum = ANY(i.indkey)

WHERE c.relname = ?
AND n.nspname = current_schema()
AND i.indisprimary

ORDER BY a.attnum
`

	var fields []string

	err := s.db.Raw(
		query,
		tableName,
	).Scan(&fields).Error

	if err != nil {
		return result, err
	}

	for _, field := range fields {
		result[field] = true
	}

	return result, nil
}

// -----------------------------------------------------------------------------
// QueryWhereFindInSet
// -----------------------------------------------------------------------------

// QueryWhereFindInSet
//
// 数据格式：
// "1,2,3"
// "a,b,c"
//
// 原来的实现：
//  1. 直接拼 SQL
//  2. 存在 SQL 注入风险
//  3. string_to_array 类型转换不够稳定
//
// openGauss 可以使用 string_to_array + ANY，
// 但参数必须走 bind variable。
func (s *OpenGaussAdapter) QueryWhereFindInSet(
	db *gorm.DB,
	input TextSlice,
	column string,
) *gorm.DB {

	if db == nil {
		db = s.db
	}

	if input == "" {
		return db
	}

	list := input.ToSlice()
	if len(list) == 0 {
		return db
	}

	conditions := make([]string, 0, len(list))
	args := make([]any, 0, len(list))

	for _, value := range list {
		switch v := value.(type) {

		case float64:
			conditions = append(
				conditions,
				fmt.Sprintf(
					`CAST(? AS TEXT) = ANY(string_to_array(%s, ','))`,
					column,
				),
			)
			args = append(args, fmt.Sprintf("%v", v))

		case string:
			conditions = append(
				conditions,
				fmt.Sprintf(
					`? = ANY(string_to_array(%s, ','))`,
					column,
				),
			)
			args = append(args, v)
		}
	}

	if len(conditions) == 0 {
		return db
	}

	return db.Where(
		strings.Join(conditions, " OR "),
		args...,
	)
}

// -----------------------------------------------------------------------------
// JSON
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) JsonExtract(column string, key string) string {
	return fmt.Sprintf("NULLIF(%s->>'%s', '')::numeric", column, key)
}

// -----------------------------------------------------------------------------
// Aggregate
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) CountWithCondition(
	condition string,
) string {
	return fmt.Sprintf(
		"SUM(CASE WHEN %s THEN 1 ELSE 0 END)",
		condition,
	)
}

func (s *OpenGaussAdapter) SumWithCondition(
	condition string,
	column string,
) string {
	return fmt.Sprintf(
		"SUM(CASE WHEN %s THEN %s ELSE 0 END)",
		condition,
		column,
	)
}

func (s *OpenGaussAdapter) AvgWithCondition(
	condition string,
	column string,
) string {
	return fmt.Sprintf(
		"AVG(CASE WHEN %s THEN %s END)",
		condition,
		column,
	)
}

// -----------------------------------------------------------------------------
// Group
// -----------------------------------------------------------------------------

// GroupExpr 分组表达式
func (s *OpenGaussAdapter) GroupExpr(column string, defaultValue ...string) string {
	dv := "未知"
	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}
	// NULLIF方法，参数1==参数2，返回NULL
	// COALESCE方法，参数1==NULL，返回参数2
	return fmt.Sprintf("COALESCE(NULLIF(CAST(%s AS TEXT), ''), '%s')", column, dv)
}

// -----------------------------------------------------------------------------
// Distinct
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) DistinctCount(
	column string,
) string {
	return fmt.Sprintf(
		"COUNT(DISTINCT %s)",
		column,
	)
}

// -----------------------------------------------------------------------------
// Column Group
// -----------------------------------------------------------------------------

func (s *OpenGaussAdapter) ColumnGroupCountExpr(
	column string,
	category ...string,
) string {
	return s.columnGroupExpr(
		"COUNT",
		"1",
		column,
		category...,
	)
}

func (s *OpenGaussAdapter) ColumnGroupSumExpr(
	sumColumn string,
	conditionColumn string,
	category ...string,
) string {
	return s.columnGroupExpr(
		"SUM",
		sumColumn,
		conditionColumn,
		category...,
	)
}

func (s *OpenGaussAdapter) columnGroupExpr(method string, valueColumn string, conditionColumn string, category ...string) string {
	var expr []string

	for _, value := range category {

		alias := strings.ReplaceAll(value, "-", "_")

		var item string

		switch method {
		case "COUNT":
			item = fmt.Sprintf(
				`SUM(CASE WHEN %s = '%s' THEN 1 ELSE 0 END) AS "%s"`,
				conditionColumn,
				value,
				alias,
			)

		case "SUM":
			item = fmt.Sprintf(
				`COALESCE(SUM(CASE WHEN %s = '%s' THEN %s ELSE 0 END), 0) AS "%s"`,
				valueColumn,
				conditionColumn,
				value,
				alias,
			)
		}

		expr = append(expr, item)
	}

	return strings.Join(expr, ",")
}

// -----------------------------------------------------------------------------
// Total
// -----------------------------------------------------------------------------

// Total 汇总统计
// exprMap key=别名	value=表达式
func (s *OpenGaussAdapter) Total(db *gorm.DB, exprMap map[string]string) map[string]any {
	var result = map[string]any{}

	expr := make([]string, 0)

	for key, value := range exprMap {
		// 生成表达式文本如：`CountWithCondition("room_type='01'") AS 个人调解室`
		expr = append(expr, fmt.Sprintf(`%s AS "%s"`, value, key))
	}

	db = db.Select(strings.Join(expr, ","))
	db = db.Scan(&result)

	return result
}

// -----------------------------------------------------------------------------
// Type Mapping
// -----------------------------------------------------------------------------

func openGaussBusinessType(
	columnType string,
) string {

	t := strings.ToLower(
		strings.TrimSpace(columnType),
	)

	switch {

	case strings.Contains(t, "bool"):
		return "bool"

	case strings.Contains(t, "date"),
		strings.Contains(t, "time"),
		strings.Contains(t, "timestamp"):
		return "datetime"

	case strings.Contains(t, "int"),
		strings.Contains(t, "decimal"),
		strings.Contains(t, "numeric"),
		strings.Contains(t, "number"),
		strings.Contains(t, "float"),
		strings.Contains(t, "double"),
		strings.Contains(t, "real"):
		return "number"

	default:
		return "string"
	}
}

func openGaussGoType(
	field string,
	columnType string,
) string {

	t := strings.ToLower(
		strings.TrimSpace(columnType),
	)

	// -------------------------------------------------------------------------
	// GORM soft delete
	// -------------------------------------------------------------------------

	if strings.EqualFold(field, "deleted_at") ||
		strings.EqualFold(field, "deletedAt") {

		return "gorm.DeletedAt"
	}

	// -------------------------------------------------------------------------
	// Integer
	// -------------------------------------------------------------------------

	switch {

	case strings.HasPrefix(t, "smallint"),
		strings.HasPrefix(t, "integer"),
		strings.HasPrefix(t, "int"),
		strings.HasPrefix(t, "int2"),
		strings.HasPrefix(t, "int4"),
		strings.HasPrefix(t, "int8"),
		strings.HasPrefix(t, "bigint"),
		strings.HasPrefix(t, "tinyint"):

		return "int"

		// -------------------------------------------------------------------------
		// Float / decimal
		// -------------------------------------------------------------------------

	case strings.HasPrefix(t, "decimal"),
		strings.HasPrefix(t, "numeric"),
		strings.HasPrefix(t, "number"),
		strings.HasPrefix(t, "double precision"),
		strings.HasPrefix(t, "double"),
		strings.HasPrefix(t, "float"),
		strings.HasPrefix(t, "real"):

		return "float64"

		// -------------------------------------------------------------------------
		// Boolean
		// -------------------------------------------------------------------------

	case strings.HasPrefix(t, "boolean"),
		strings.HasPrefix(t, "bool"):

		return "ctype.Bool"

		// -------------------------------------------------------------------------
		// Time
		// -------------------------------------------------------------------------

	case strings.HasPrefix(t, "timestamp"),
		strings.HasPrefix(t, "datetime"):

		return "*moment.LocalTime"

	case strings.HasPrefix(t, "date"):

		return "*moment.LocalTime"

		// -------------------------------------------------------------------------
		// JSON
		// -------------------------------------------------------------------------

	case strings.HasPrefix(t, "json"),
		strings.HasPrefix(t, "jsonb"):

		return "string"

		// -------------------------------------------------------------------------
		// Array
		// -------------------------------------------------------------------------

	case strings.HasSuffix(t, "[]"):

		return "string"

	default:

		return "string"
	}
}

func quoteSearchPath(
	schema string,
) string {

	if schema == "" {
		return "public"
	}

	return quoteIdentifier(schema)
}

func (s *OpenGaussAdapter) QI(field string) string {
	if s.config.Mode == "B" {
		return mysqlQI(field)
	}

	return pgsqlQI(field)
}

func (s *OpenGaussAdapter) AF(alias, field string) string {
	if s.config.Mode == "B" {
		return mysqlAF(alias, field)
	}
	return pgsqlAF(alias, field)
}

func (s *OpenGaussAdapter) AFSelect(g ...AFGroup) string {
	var expr = make([]string, 0)
	for _, v := range g {
		for _, f := range v.Field {
			expr = append(expr, s.AF(v.Alias, f))
		}
	}
	return strings.Join(expr, ",")
}

// Compare 比较
func (s *OpenGaussAdapter) Compare(column string, operator string, value any, typ CastType) string {
	if s.config.Mode == "B" {
		return fmt.Sprintf("%s %s %v", mysqlCast(column, typ), operator, value)
	}
	return fmt.Sprintf("%s %s %v", pgsqlCast(column, typ), operator, value)
}
