// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/7/23
// 描述：
// *****************************************************************************

package db

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewSqlite(config Config) *Api {
	config.IntValue(&config.MaxIdleConns, 1)
	config.IntValue(&config.MaxOpenConns, 1)
	var dbApi = Api{
		Config: config,
	}

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("./%v.db", config.Dbname)), &gorm.Config{})
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
		config.InitAfter()
	}

	dbApi.DB = db
	return &dbApi
}
