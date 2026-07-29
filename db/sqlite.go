// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/7/23
// 描述：
// *****************************************************************************

package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSqlite(config Config) *Api {
	config.IntValue(&config.MaxIdleConns, 1)
	config.IntValue(&config.MaxOpenConns, 1)
	var dbApi = Api{
		Config: config,
	}

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("./%v.db", config.Dbname)), &gorm.Config{
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

	db.Exec("PRAGMA busy_timeout=5000")

	if config.InitAfter != nil {
		config.InitAfter(db)
	}

	dbApi.DB = db
	return &dbApi
}
