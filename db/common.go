// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：公共方法
//
// Where闭包Or条件示例
// QueryWhere(nil, "张三", "name1").Or(mysql.QueryWhereFindInSet(nil, "1,2,3", "name2"))
// QueryWhereFindInSet(nil, "1,2,3", "name1").Or(mysql.QueryWhereFindInSet(nil, "1,2,3", "name2"))
// QueryWhere的都支持闭包Or组装
//
//
// *****************************************************************************

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lgdzz/vingo-utils-v3/moment"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type Common struct {
	*gorm.DB
}

type Table struct {
	TableName    string
	ModelName    string
	TableComment string
	TableColumns []Column
	Date         string
}

type Column struct {
	Field    string
	Field2   string
	Type     string
	Null     string
	Key      string
	Default  sql.NullString
	Extra    string
	Comment  string
	DataName string
	DataType string
	JsonName string
}

func NewCommon(db *gorm.DB) *Common {
	return &Common{DB: db}
}

// Diff 开启Diff功能
// diff功能开启后，查询到的数据会自动设置到Diff.Old字段中，更新后的数据会自动设置到Diff.New字段中
// 并且自动执行比较
// 可以调用 DiffBox.ResultContent() 获取变更字段描述
func (s *Common) Diff() *gorm.DB {
	return s.DB.Set("diff", true)
}

// Operator diff操作人
func (s *Common) Operator(ctx *vingo.Context) *gorm.DB {
	return s.DB.Set("ctx", ctx)
}

// OperatorWithTx diff操作人
func (s *Common) OperatorWithTx(tx *gorm.DB, ctx *vingo.Context) *gorm.DB {
	return tx.Set("ctx", ctx)
}

// AutoCommit 自动提交事务
func (s *Common) AutoCommit(tx *gorm.DB, callback ...func()) {
	if r := recover(); r != nil {
		//fmt.Printf("%T\n%v\n", r, r)
		tx.Rollback()
		if len(callback) > 0 && callback[0] != nil {
			callback[0]()
		}
		panic(r)
	} else if err := tx.Statement.Error; err != nil {
		//fmt.Println("数据库异常事务回滚")
		tx.Rollback()
		if len(callback) > 0 && callback[0] != nil {
			callback[0]()
		}
		panic(err.Error())
	} else {
		//fmt.Println("事务提交")
		tx.Commit()
		if len(callback) > 1 && callback[1] != nil {
			callback[1]()
		}
	}
}

// FastCommit 快捷事务
func (s *Common) FastCommit(handler func(tx *gorm.DB)) {
	tx := s.DB.Begin()
	defer s.AutoCommit(tx)
	handler(tx)
}

// OrderWithTree 树结构数据排序
func (s *Common) OrderWithTree() string {
	return "len asc,sort asc,id asc"
}

// Exists 查询记录是否存在
func (s *Common) Exists(model any, condition ...any) bool {
	err := s.DB.First(model, condition...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else if err != nil {
		panic(err.Error())
	}
	return true
}

// TxExists 事务内查询记录是否存在
func (s *Common) TxExists(tx *gorm.DB, model any, condition ...any) bool {
	err := tx.First(model, condition...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else if err != nil {
		panic(err.Error())
	}
	return true
}

// NotExistsErr 记录不存在时抛出错误
func (s *Common) NotExistsErr(model any, condition ...any) {
	err := s.DB.First(model, condition...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err.Error())
	} else if err != nil {
		panic(err.Error())
	}
}

// NotExistsErrMsg 记录不存在时抛出错误
func (s *Common) NotExistsErrMsg(msg string, model any, condition ...any) {
	err := s.DB.First(model, condition...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic(msg)
	} else if err != nil {
		panic(err.Error())
	}
}

// TXNotExistsErr 记录不存在时抛出错误(事务内)
func (s *Common) TXNotExistsErr(tx *gorm.DB, model any, condition ...any) {
	err := tx.First(model, condition...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err.Error())
	} else if err != nil {
		panic(err.Error())
	}
}

// CheckHasChild 检查是否有子项，有则抛出异常
func (s *Common) CheckHasChild(model any, id int) {
	err := s.DB.First(model, "pid=?", id)
	if !errors.Is(err.Error, gorm.ErrRecordNotFound) {
		panic("记录有子项，删除失败")
	}
}

// QueryDb 数据库查询
func (s *Common) QueryDb(db *gorm.DB) *gorm.DB {
	if db == nil {
		db = s.DB
	}
	return db
}

// QueryWhere 条件查询
func (s *Common) QueryWhere(db *gorm.DB, input any, column ...string) *gorm.DB {
	db = s.QueryDb(db)
	valueOf := reflect.ValueOf(input)
	typeOf := valueOf.Type()
	if typeOf.Kind() == reflect.Ptr {
		if valueOf.IsNil() {
			//fmt.Println("空指针无条件")
			return db
		} else {
			input = valueOf.Elem().Interface()
		}
	} else {
		switch v := input.(type) {
		case string:
			if v == "" {
				//fmt.Println("string无条件")
				return db
			}
		}
		input = valueOf.Interface()
	}
	if input != nil {
		var text []string
		for _, item := range column {
			text = append(text, fmt.Sprintf("%v=@text", item))
		}
		db = db.Where(strings.Join(text, " OR "), sql.Named("text", input))
	}
	return db
}

// QueryWhereIn 包含查询
func (s *Common) QueryWhereIn(db *gorm.DB, input TextSlice, column string) *gorm.DB {
	db = s.QueryDb(db)
	if input != "" {
		db = db.Where(fmt.Sprintf("%v in(?)", column), input.ToSlice())
	}
	return db
}

// QueryWhereNotIn 排除查询
func (s *Common) QueryWhereNotIn(db *gorm.DB, input TextSlice, column string) *gorm.DB {
	db = s.QueryDb(db)
	if input != "" {
		db = db.Where(fmt.Sprintf("%v not in(?)", column), input.ToSlice())
	}
	return db
}

// QueryWhereBetween 范围查询
func (s *Common) QueryWhereBetween(db *gorm.DB, input Between[float64], column string) *gorm.DB {
	db = s.QueryDb(db)
	if input != "" {
		a, b := input.Between()
		db = db.Where(fmt.Sprintf("%v BETWEEN ? AND ?", column), a, b)
	}
	return db
}

// QueryWhereDate 时间范围查询
func (s *Common) QueryWhereDate(db *gorm.DB, input moment.DateTextRange, column string) *gorm.DB {
	db = s.QueryDb(db)
	if input != "" {
		a, b := input.ToStruct().BetweenText()
		db = db.Where(fmt.Sprintf("%v BETWEEN ? AND ?", column), a, b)
	}
	return db
}

// QueryWhereLike 模糊查询
func (s *Common) QueryWhereLike(db *gorm.DB, input TextSlice, column ...string) *gorm.DB {
	db = s.QueryDb(db)
	if !input.IsEmpty() {
		var text []string
		for _, value := range input.ToStringSlice() {
			value = fmt.Sprintf("%%%v%%", strings.TrimSpace(value))
			for _, item := range column {
				text = append(text, fmt.Sprintf("%v LIKE '%v'", item, value))
			}
		}
		fmt.Println(strings.Join(text, " OR "))
		db = db.Where(strings.Join(text, " OR "))
	}
	return db
}

// QueryWhereLikeRight 右模糊查询
func (s *Common) QueryWhereLikeRight(db *gorm.DB, input TextSlice, column ...string) *gorm.DB {
	db = s.QueryDb(db)
	if !input.IsEmpty() {
		var text []string
		for _, value := range input.ToStringSlice() {
			value = fmt.Sprintf("%v%%", strings.TrimSpace(value))
			for _, item := range column {
				text = append(text, fmt.Sprintf("%v LIKE '%v'", item, value))
			}
		}
		db = db.Where(strings.Join(text, " OR "))
	}
	return db
}

// QueryWherePath 查询路径数据
func (s *Common) QueryWherePath(db *gorm.DB, input any, column string) *gorm.DB {
	db = s.QueryDb(db)

	var queries []string
	switch v := input.(type) {
	case string:
		if v == "" {
			return db
		}
		queries = []string{v}
	case []string:
		if len(v) == 0 {
			return db
		}
		queries = v
	default:
		return db
	}

	var conditions []string
	for _, value := range queries {
		conditions = append(conditions, fmt.Sprintf("(%v='%v' OR %v LIKE '%v,%%')", column, value, column, value))
	}

	db = db.Where(strings.Join(conditions, " OR "))
	return db
}

// QueryWhereNotDeleted 查询未删除的数据
func (s *Common) QueryWhereNotDeleted(db *gorm.DB, column string) *gorm.DB {
	db = s.QueryDb(db)
	db = db.Where(fmt.Sprintf("%v IS NULL", column))
	return db
}

// ChineseSortString 指定字段第一个汉字按A-Z排序
func (s *Common) ChineseSortString(column string) string {
	return fmt.Sprintf("CONVERT(SUBSTR(%v, 1, 1) USING gbk)", column)
}
