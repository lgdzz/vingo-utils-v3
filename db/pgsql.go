// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：pgsql数据
// *****************************************************************************

package db

import (
	"fmt"
	"github.com/lgdzz/vingo-utils-v3/db/book"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

func NewPgSql(config Config) *DatabaseApi {
	config.StringValue(&config.Host, "127.0.0.1")
	config.StringValue(&config.Port, "54321")
	config.StringValue(&config.Username, "system")
	config.StringValue(&config.Password, "123456")
	config.StringValue(&config.Charset, "utf8mb4")
	config.IntValue(&config.MaxIdleConns, 10)
	config.IntValue(&config.MaxOpenConns, 100)

	var dbApi = DatabaseApi{
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

// Book 数据库字典
func (s *PgsqlAdapter) Book() string {
	return book.BuildPgsqlBook(s.db)
}

// QueryWhereFindInSet 在字符串[1,2,3...]集合中查找
func (s *PgsqlAdapter) QueryWhereFindInSet(db *gorm.DB, input TextSlice, column string) *gorm.DB {
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
