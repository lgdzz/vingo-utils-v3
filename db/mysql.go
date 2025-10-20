// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：mysql数据库
// *****************************************************************************

package db

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/lgdzz/vingo-utils-v3/db/book"
	"github.com/lgdzz/vingo-utils-v3/db/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 新建一个数据库连接池
func NewMysql(config Config) *Api {
	config.StringValue(&config.Host, "127.0.0.1")
	config.StringValue(&config.Port, "3306")
	config.StringValue(&config.Username, "root")
	config.StringValue(&config.Password, "123456789")
	config.StringValue(&config.Charset, "utf8mb4")
	config.IntValue(&config.MaxIdleConns, 10)
	config.IntValue(&config.MaxOpenConns, 100)

	var dbApi = Api{
		Config: config,
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Dbname,
		config.Charset)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
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
			tmp := time.Now().Local().Format("2006-01-02 15:04:05")
			now, _ := time.ParseInLocation("2006-01-02 15:04:05", tmp, time.Local)
			return now
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

type MysqlAdapter struct {
	db *gorm.DB
}

func NewMysqlAdapter(db *gorm.DB) *MysqlAdapter {
	return &MysqlAdapter{db: db}
}

func (s *MysqlAdapter) GetDatabaseName() (string, error) {
	var dbName string
	err := s.db.Raw("SELECT DATABASE()").Scan(&dbName).Error
	return dbName, err
}

func (s *MysqlAdapter) GetTableComment(dbName, tableName string) (string, error) {
	var tableComment string
	err := s.db.Table("information_schema.tables").
		Where("table_schema=? AND table_name=?", dbName, tableName).
		Select("table_comment").
		Scan(&tableComment).Error
	return tableComment, err
}

func (s *MysqlAdapter) GetColumns(tableName string) ([]Column, error) {
	var columns []Column
	err := s.db.Raw(fmt.Sprintf("SHOW FULL COLUMNS FROM `%v`", tableName)).Scan(&columns).Error
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
func (s *MysqlAdapter) Book() string {
	return book.BuildMysqlBook(s.db)
}

// ModelFiles 生成模型文件
func (s *MysqlAdapter) ModelFiles(tableNames ...string) (bool, error) {
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

func (s *MysqlAdapter) modelFile(tableName string) (bool, error) {
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
		case strutil.HasPrefixAny(col.Type, []string{"int", "tinyint", "smallint"}):
			col.DataType = "int"
		case strutil.HasPrefixAny(col.Type, []string{"decimal"}):
			col.DataType = "float64"
		case strutil.HasPrefixAny(col.Field, []string{"deletedAt", "deleted_at"}):
			col.DataType = "gorm.DeletedAt"
		case strutil.HasPrefixAny(col.Type, []string{"timestamp", "datetime"}):
			col.DataType = "*moment.LocalTime"
		default:
			col.DataType = "string"
		}
		col.JsonName = strutil.CamelCase(col.Field)
		col.DataName = strutil.UpperFirst(col.JsonName)
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
func (s *MysqlAdapter) QueryWhereFindInSet(db *gorm.DB, input TextSlice, column string) *gorm.DB {
	if db == nil {
		db = s.db
	}
	if input != "" {
		var text []string
		list := input.ToSlice()
		for _, value := range list {
			switch value.(type) {
			case float64:
				text = append(text, fmt.Sprintf("FIND_IN_SET(%v,%v)", value, column))
			case string:
				text = append(text, fmt.Sprintf("FIND_IN_SET('%v',%v)", value, column))
			}
		}
		if len(text) > 0 {
			db = db.Where(strings.Join(text, " OR "))
		}
	}
	return db
}

func (s *MysqlAdapter) JsonExtract(column string, key string) string {
	return fmt.Sprintf("JSON_EXTRACT(%v,'$.%v')", column, key)
}
