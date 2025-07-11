// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：mysql数据库
// *****************************************************************************

package db

import (
	"fmt"
	"github.com/lgdzz/vingo-utils-v3/db/book"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

// 新建一个数据库连接池
func NewMysql(config Config) *DatabaseApi {
	config.StringValue(&config.Host, "127.0.0.1")
	config.StringValue(&config.Port, "3306")
	config.StringValue(&config.Username, "root")
	config.StringValue(&config.Password, "123456789")
	config.StringValue(&config.Charset, "utf8mb4")
	config.IntValue(&config.MaxIdleConns, 10)
	config.IntValue(&config.MaxOpenConns, 100)

	var dbApi = DatabaseApi{
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

// Book 数据库字典
func (s *MysqlAdapter) Book() string {
	return book.BuildMysqlBook(s.db)
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
