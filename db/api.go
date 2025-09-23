// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：数据库API
// *****************************************************************************

package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fatih/color"
	"github.com/lgdzz/vingo-utils-exception/exception"
	"github.com/lgdzz/vingo-utils-v3/ctype"
	"github.com/lgdzz/vingo-utils-v3/pool"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"gorm.io/gorm"
	"reflect"
)

type Api struct {
	*gorm.DB
	*Common
	Adapter
	Config    Config
	ChangeLog func(tx *gorm.DB, option ChangeLogOption)
}

type ChangeLogOption struct {
	Ctx             any
	TableName       string
	Description     *string
	PrimaryKeyValue any
}

func NewDatabase(config Config) *Api {
	var api *Api
	switch config.Driver {
	case "pgsql":
		api = NewPgSql(config)
		api.Adapter = NewPgsqlAdapter(api.DB)
	default:
		api = NewMysql(config)
		api.Adapter = NewMysqlAdapter(api.DB)
	}

	// 公共方法
	api.Common = NewCommon(api.DB)
	// 设置密文字段的key
	ctype.Secret = []byte(config.Secret)

	// 注册统一异常插件
	RegisterAfterQuery(api)
	RegisterAfterCreate(api)
	RegisterBeforeUpdate(api)
	RegisterAfterUpdate(api)
	RegisterAfterDelete(api)
	return api
}

// RegisterAfterQuery 注册统一查询异常插件
func RegisterAfterQuery(api *Api) {

	err := api.DB.Callback().Query().After("gorm:query").Register("vingo:after_query", func(db *gorm.DB) {
		if db.Error != nil && !errors.Is(db.Error, gorm.ErrRecordNotFound) {
			_, _ = color.New(color.FgRed).Printf("[DB ERROR] %T: %v\n", db.Error, db.Error)
			panic(&exception.DbException{Message: db.Error.Error()})
		}

		// 如果开启diff
		if need, ok := db.Get("diff"); ok && need.(bool) {
			db.InstanceSet("diff", false) // 下次不再触发
			setDiffOldValue(db.Statement.Dest)
		}
	})
	if err != nil {
		panic(fmt.Sprintf("插件注册失败: %v", err.Error()))
	}
}

// RegisterAfterCreate 注册统一创建异常插件
func RegisterAfterCreate(api *Api) {
	err := api.DB.Callback().Create().After("gorm:create").Register("vingo:after_create", func(db *gorm.DB) {
		if db.Error != nil {
			_, _ = color.New(color.FgRed).Printf("[DB ERROR] %T: %v\n", db.Error, db.Error)
			panic(&exception.DbException{Message: db.Error.Error()})
		}
	})
	if err != nil {
		panic(fmt.Sprintf("插件注册失败: %v", err.Error()))
	}
}

func RegisterBeforeUpdate(api *Api) {
	err := api.DB.Callback().Update().Before("gorm:before_update").Register("vingo:before_update", func(db *gorm.DB) {
		// 处理diff新值（更新前）
		description := setDiffNewValue(db.Statement.Dest)
		db.Set("diff_description", description)
	})
	if err != nil {
		panic(fmt.Sprintf("插件注册失败: %v", err.Error()))
	}
}

// RegisterAfterUpdate 注册统一更新异常插件
func RegisterAfterUpdate(api *Api) {
	err := api.DB.Callback().Update().After("gorm:update").Register("vingo:after_update", func(db *gorm.DB) {
		if db.Error != nil {
			_, _ = color.New(color.FgRed).Printf("[DB ERROR] %T: %v\n", db.Error, db.Error)
			panic(&exception.DbException{Message: db.Error.Error()})
		}

		// 处理diff新值（更新后，补充一些钩子中的赋值）
		description := setDiffNewValue(db.Statement.Dest)
		db.Set("diff_description", description)

		// 更新后写变更日志
		if api.ChangeLog != nil {
			if ctx, ok := db.Get("ctx"); ok {
				if description, ok := db.Get("diff_description"); ok {
					api.ChangeLog(db.Session(&gorm.Session{NewDB: true}), ChangeLogOption{
						Ctx:             ctx,
						TableName:       db.Statement.Table,
						Description:     description.(*string),
						PrimaryKeyValue: getPrimaryKeyValue(db),
					})
				}
			}
		}
	})
	if err != nil {
		panic(fmt.Sprintf("插件注册失败: %v", err.Error()))
	}
}

// RegisterAfterDelete 注册统一删除异常插件
func RegisterAfterDelete(api *Api) {
	err := api.DB.Callback().Delete().After("gorm:delete").Register("vingo:after_delete", func(db *gorm.DB) {
		if db.Error != nil {
			_, _ = color.New(color.FgRed).Printf("[DB ERROR] %T: %v\n", db.Error, db.Error)
			panic(&exception.DbException{Message: db.Error.Error()})
		}
	})
	if err != nil {
		panic(fmt.Sprintf("插件注册失败: %v", err.Error()))
	}
}

func getPrimaryKeyValue(tx *gorm.DB) any {
	if tx.Statement == nil || tx.Statement.Schema == nil {
		return nil
	}

	ctx := context.Background()
	rv := tx.Statement.ReflectValue

	for _, field := range tx.Statement.Schema.PrimaryFields {
		val, _ := field.ValueOf(ctx, rv)
		return val // 默认只取第一个主键字段
	}
	return nil
}

func setPtrField(field reflect.Value, val interface{}) {
	v := reflect.ValueOf(val)
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	field.Set(ptr)
}

// setDiffOldValue 设置Diff旧值
func setDiffOldValue(dest any) {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return // 必须是指针
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return // 必须是结构体
	}

	field := rv.FieldByName("Diff")
	if !field.IsValid() || field.Kind() != reflect.Struct || !field.CanSet() {
		return // 不存在或不是结构体字段或不可设置
	}

	oldField := field.FieldByName("Old")
	if oldField.IsNil() { // ✅ 只有在 Old 没有值时才设置
		setPtrField(oldField, rv.Interface())
	}

}

// setDiffNewValue 设置Diff新值
func setDiffNewValue(dest any) *string {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return nil // 必须是指针
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil // 必须是结构体
	}

	field := rv.FieldByName("Diff")
	if !field.IsValid() || field.Kind() != reflect.Struct || !field.CanSet() {
		return nil // 不存在或不是指针字段或不可设置
	}

	old := field.FieldByName("Old")
	if !old.IsNil() {
		newField := field.FieldByName("New") // 访问 New 字段
		if newField.IsValid() && newField.CanSet() {

			setPtrField(newField, rv.Interface())

			// 调用 Compare 方法
			if diffBox, ok := field.Addr().Interface().(DiffBoxInterface); ok {
				diffBox.Compare()
				if diffBox.HasChange() {
					j := diffBox.ResultJson()
					if j == nil {
						return nil
					}
					return pointer.Of(j.String())
				}
			}
		}

	}
	return nil
}

type QueryListOption[T any] struct {
	Iteratee func(i int, item T) any

	IterateePool func(i int, item *T) // 映射函数（协程池）
	PoolResult   *[]pool.Result       // 协程池结果
	MaxWorkers   int                  // 最大协程数

	IsTree       bool // 是否返回树结构，只支持id,pid为number类型的主键，其他情况使用ListCallback自定义
	ListCallback func(list []T) any
}

// QueryList 列表查询
// 如果传入的PageQuery.Limit.Page为nil，则为不分页查询，否则为分页查询
func QueryList[T any](db *gorm.DB, pq PageQuery, option *QueryListOption[T]) any {
	// 不分页模式
	if pq.Limit.Page == nil {
		var result = make([]T, 0)
		if pq.Limit.Size > 0 {
			db = db.Limit(pq.Limit.Size)
		}
		if pq.OrderRaw != nil {
			db = db.Order(*pq.OrderRaw)
		}
		db.Scan(&result)
		if option != nil {
			if option.Iteratee != nil {
				data := slice.Map(result, func(index int, item T) any {
					return option.Iteratee(index, item)
				})
				if option.IsTree {
					return vingo.FastTree[float64](data)
				}
				return data
			}

			if option.IsTree {
				return vingo.FastTree[float64](result)
			}

			if option.ListCallback != nil {
				return option.ListCallback(result)
			}
		}

		return result
	}

	if option == nil {
		option = &QueryListOption[T]{}
	}

	// 分页模式
	return NewPage(QueryOption[T]{
		Db:           db,
		Query:        pq,
		Iteratee:     option.Iteratee,
		IterateePool: option.IterateePool,
		PoolResult:   option.PoolResult,
		MaxWorkers:   option.MaxWorkers,
	})
}

func mustFind[T any](db *gorm.DB, enableDiff bool, condition ...any) (row T) {
	query := db
	if enableDiff {
		query = db.Set("diff", true)
	}
	if err := query.First(&row, condition...).Error; err != nil {
		typeName := reflect.TypeOf(row).Name()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//_, _ = color.New(color.FgRed).Println(db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			//	return tx.First(&row, condition...)
			//}))
			panic(fmt.Sprintf("Model[%s]记录不存在", typeName))
		}
		panic(fmt.Sprintf("Model[%s]查询失败，错误:%v", typeName, err.Error()))
	}
	return
}

// Find 根据任意条件查询
func Find[T any](db *gorm.DB, condition ...any) T {
	return mustFind[T](db, false, condition...)
}

// FindWithDiff 根据任意条件查询，并开启 diff 处理
func FindWithDiff[T any](db *gorm.DB, condition ...any) T {
	return mustFind[T](db, true, condition...)
}

// FindById 根据 ID 查询
func FindById[T any](db *gorm.DB, id int) T {
	return mustFind[T](db, false, id)
}

// FindByIdWithDiff 根据 ID 查询，并开启 diff 处理
func FindByIdWithDiff[T any](db *gorm.DB, id int) T {
	return mustFind[T](db, true, id)
}
