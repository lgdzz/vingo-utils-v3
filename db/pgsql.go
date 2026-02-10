// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：pgsql数据
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

func NewPgSql(config Config) *Api {
	config.StringValue(&config.Host, "127.0.0.1")
	config.StringValue(&config.Port, "54321")
	config.StringValue(&config.Username, "system")
	config.StringValue(&config.Password, "123456")
	config.StringValue(&config.Charset, "utf8mb4")
	config.IntValue(&config.MaxIdleConns, 10)
	config.IntValue(&config.MaxOpenConns, 100)

	var dbApi = Api{
		Config: config,
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai", config.Host, config.Port, config.Username, config.Password, config.Dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
			logger.Config{
				SlowThreshold:             time.Second, // 慢 SQL 阈值
				LogLevel:                  logger.Warn, // 日志级别
				IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  true,        // 禁用彩色打印
			},
		),
		NowFunc: func() time.Time {
			loc, _ := time.LoadLocation("Asia/Shanghai")
			return time.Now().In(loc)
		},
	})
	if err != nil {
		panic("Error to Db connection, err: " + err.Error())
	}

	// 连接池配置
	sqlDB, _ := db.DB()
	// 最大空闲数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	// 最大连接数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	// 连接最大存活时长
	sqlDB.SetConnMaxLifetime(60 * time.Minute)

	dbApi.DB = db
	return &dbApi
}

type PgsqlAdapter struct {
	db *gorm.DB
}

func NewPgsqlAdapter(db *gorm.DB) *PgsqlAdapter {
	return &PgsqlAdapter{db: db}
}

func (s *PgsqlAdapter) GetDatabaseName() (string, error) {
	var dbName string
	err := s.db.Raw("SELECT current_database()").Scan(&dbName).Error
	return dbName, err
}

func (s *PgsqlAdapter) GetTableComment(dbName, tableName string) (string, error) {
	var tableComment sql.NullString
	queryTableComment := `SELECT obj_description(c.oid, 'pg_class') 
                          FROM pg_class c 
                          WHERE relname = ?`
	err := s.db.Raw(queryTableComment, tableName).Scan(&tableComment).Error
	if err != nil {
		return "", err
	}
	if tableComment.Valid {
		return tableComment.String, nil
	}
	return "", nil // 表注释为空
}

func (s *PgsqlAdapter) GetColumns(tableName string) ([]Column, error) {
	var columns []Column
	queryColumn := `
			SELECT 
				a.attname AS field,
				format_type(a.atttypid, a.atttypmod) AS type,
				col_description(a.attrelid, a.attnum) AS comment,
				CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS null
			FROM 
				pg_attribute a
			JOIN 
				pg_class c ON a.attrelid = c.oid
			JOIN 
				pg_namespace n ON c.relnamespace = n.oid
			WHERE 
				c.relname = ? 
				AND a.attnum > 0 
				AND NOT a.attisdropped
		`
	err := s.db.Raw(queryColumn, tableName).Scan(&columns).Error

	if err == nil {
		columns = slice.Map(columns, func(index int, item Column) Column {
			t := strings.ToLower(item.Type) // 统一小写
			switch {
			case strutil.ContainsAny(t, []string{"bool", "tinyint(1)"}):
				item.BusinessType = "bool"
			case strutil.ContainsAny(t, []string{"date", "datetime", "timestamp"}):
				item.BusinessType = "datetime"
			case strutil.ContainsAny(t, []string{"int", "bigint", "float", "double", "decimal"}):
				item.BusinessType = "number"
			default:
				item.BusinessType = "string"
			}
			return item
		})
	}

	return columns, err
}

// Book 数据库字典
func (s *PgsqlAdapter) Book() string {
	return book.BuildPgsqlBook(s.db)
}

// ModelFiles 生成模型文件
func (s *PgsqlAdapter) ModelFiles(tableNames ...string) (bool, error) {
	if err := os.MkdirAll("model", 0777); err != nil {
		return false, fmt.Errorf("创建 model 目录失败: %w", err)
	}

	for _, tableName := range tableNames {
		success, err := s.modelFile(tableName)
		if err != nil {
			fmt.Printf("生成表 [%s] 模型失败：%v\n", tableName, err)
			return false, err
		}
		if !success {
			return false, fmt.Errorf("生成表 [%s] 模型失败", tableName)
		}
	}
	return true, nil
}

func (s *PgsqlAdapter) modelFile(tableName string) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("⚠️ 发生 panic，可能是数据库连接失败")
			fmt.Println("panic:", r)
		}
	}()

	modelPath := filepath.Join("model", tableName+".go")

	// 获取数据库名
	dbName, err := s.GetDatabaseName()
	if err != nil {
		return false, fmt.Errorf("获取数据库名失败: %w", err)
	}

	// 获取表注释
	tableComment, err := s.GetTableComment(dbName, tableName)
	if err != nil {
		return false, fmt.Errorf("获取表注释失败: %w", err)
	}

	// 获取字段信息
	columns, err := s.GetColumns(tableName)
	if err != nil {
		return false, fmt.Errorf("获取字段失败: %w", err)
	}

	columns = slice.Map(columns, func(index int, col Column) Column {
		switch {
		case strutil.HasPrefixAny(col.Type, []string{"integer", "int2", "int4", "int8", "bigint", "smallint", "int", "tinyint", "smallint"}):
			col.DataType = "int"
		case strutil.HasPrefixAny(col.Type, []string{"decimal", "numeric", "double precision", "real"}):
			col.DataType = "float64"
		case strutil.HasPrefixAny(col.Type, []string{"boolean", "bool"}):
			col.DataType = "ctype.Bool"
		case strutil.HasPrefixAny(col.Field, []string{"deletedAt", "deleted_at"}):
			col.DataType = "gorm.DeletedAt"
		case strutil.HasPrefixAny(col.Type, []string{"timestamp", "datetime"}):
			col.DataType = "*moment.LocalTime"
		default:
			col.DataType = "string"
		}
		col.JsonName = strutil.CamelCase(col.Field)
		col.DataName = strutil.UpperFirst(col.JsonName)

		// 判断主键
		var isPrimaryKey string
		s.db.Raw(`
				SELECT a.attname 
				FROM pg_index i 
				JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
				WHERE i.indrelid = ?::regclass AND i.indisprimary
			`, tableName).Scan(&isPrimaryKey)

		if isPrimaryKey == col.Field {
			col.Key = "PRI"
		}

		return col
	})

	// 渲染模板
	tmpl, err := template.New("tpl").Option("missingkey=zero").Parse(model.ModelTpl)
	if err != nil {
		return false, fmt.Errorf("解析模板失败: %w", err)
	}

	outputFile, err := os.Create(modelPath)
	if err != nil {
		return false, fmt.Errorf("创建文件失败: %w", err)
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, Table{
		TableName:    tableName,
		ModelName:    strutil.UpperFirst(strutil.CamelCase(tableName)),
		TableComment: tableComment,
		TableColumns: columns,
		Date:         time.Now().Format("2006/01/02"),
	})
	if err != nil {
		return false, fmt.Errorf("渲染模板失败: %w", err)
	}

	fmt.Printf("✅ 成功生成模型文件: %s\n", modelPath)
	return true, nil
}

// QueryWhereFindInSet 在字符串[1,2,3...]集合中查找
func (s *PgsqlAdapter) QueryWhereFindInSet(db *gorm.DB, input TextSlice, column string) *gorm.DB {
	if db == nil {
		db = s.db
	}
	if input != "" {
		var text []string
		list := input.ToSlice()
		for _, value := range list {
			switch value.(type) {
			case float64:
				text = append(text, fmt.Sprintf("%v=ANY(string_to_array(%v,','))", value, column))
			case string:
				text = append(text, fmt.Sprintf("'%v'=ANY(string_to_array(%v,','))", value, column))
			}
		}
		if len(text) > 0 {
			db = db.Where(strings.Join(text, " OR "))
		}
	}
	return db
}

func (s *PgsqlAdapter) JsonExtract(column string, key string) string {
	return fmt.Sprintf("(%v->>'%v')::numeric", column, key)
}
