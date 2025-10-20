package db

import (
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/gorm"
	"log"
	"reflect"
	"strings"
)

type DiffBoxInterface interface {
	Compare()
	HasChange() bool
	ResultContent() string
	ResultJson() *ChangeItems
}

type DiffBox[T any] struct {
	Old    *T
	New    *T
	Result *map[string]DiffItem
}

type DiffItem struct {
	Column   string
	OldValue any
	NewValue any
	Message  string
}

type ChangeItems struct {
	Old map[string]any `json:"old"`
	New map[string]any `json:"new"`
}

func (s ChangeItems) String() string {
	output, err := json.Marshal(s)
	if err != nil {
		log.Println(err.Error())
	}
	return string(output)
}

func (s *DiffItem) toJSONStr(v any) string {
	if v == nil {
		return "null"
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		b, err := json.Marshal(v)
		if err == nil {
			return string(b)
		}
	}
	return fmt.Sprintf("%v", v) // 非结构体直接用原格式
}

func (s *DiffItem) SetMessage() {
	oldStr := s.toJSONStr(s.OldValue)
	newStr := s.toJSONStr(s.NewValue)
	s.Message = fmt.Sprintf("将%v的值[%v]变更为[%v]；", s.Column, oldStr, newStr)
}

// SetNew 设置新值
func (s *DiffBox[T]) SetNew(newValue T) {
	s.New = &newValue
}

// SetNewAndCompare 设置新值并执行比对，支持回调
func (s *DiffBox[T]) SetNewAndCompare(newValue T, result func(diff *DiffBox[T])) {
	s.SetNew(newValue)
	s.Compare()
	if result != nil {
		result(s)
	}
}

// ensureCompared 保证 Compare 已执行
func (s *DiffBox[T]) ensureCompared() {
	if s.Result == nil {
		s.Compare()
	}
}

// Compare 比较 Old 和 New
func (s *DiffBox[T]) Compare() {
	result := map[string]DiffItem{}

	if s.Old == nil || s.New == nil {
		return
	}
	oldVal := reflect.ValueOf(s.Old)
	newVal := reflect.ValueOf(s.New)

	if oldVal.Kind() == reflect.Ptr {
		if oldVal.IsNil() {
			return
		}
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		if newVal.IsNil() {
			return
		}
		newVal = newVal.Elem()
	}

	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return
	}

	oldType := oldVal.Type()
	newType := newVal.Type()
	if oldType != newType {
		return
	}

	for i := 0; i < oldVal.NumField(); i++ {
		fieldType := oldType.Field(i)
		name := fieldType.Name

		if name == "Diff" || slice.Contain([]string{"CreatedAt", "UpdatedAt", "DeletedAt"}, name) {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			diffItem := DiffItem{
				Column:   name,
				OldValue: oldField.Interface(),
				NewValue: newField.Interface(),
			}
			diffItem.SetMessage()
			result[diffItem.Column] = diffItem
		}
	}
	s.Result = &result
}

// IsChange 判断某字段是否变更
func (s *DiffBox[T]) IsChange(column string) bool {
	s.ensureCompared()
	_, ok := (*s.Result)[column]
	return ok
}

// IsModify 如果字段变更则执行回调
func (s *DiffBox[T]) IsModify(column string, callback func()) {
	if s.IsChange(column) && callback != nil {
		callback()
	}
}

// IsChangeOr 任意字段变更则返回 true
func (s *DiffBox[T]) IsChangeOr(columns ...string) bool {
	for _, column := range columns {
		if s.IsChange(column) {
			return true
		}
	}
	return false
}

// IsChangeAnd 所有字段都变更才返回 true
func (s *DiffBox[T]) IsChangeAnd(columns ...string) bool {
	for _, column := range columns {
		if !s.IsChange(column) {
			return false
		}
	}
	return true
}

// ResultContent 生成变更摘要
func (s *DiffBox[T]) ResultContent() string {
	s.ensureCompared()
	if !s.HasChange() {
		return "无修改"
	}
	var builder strings.Builder
	for _, item := range *s.Result {
		builder.WriteString(item.Message)
	}
	return builder.String()
}

func (s *DiffBox[T]) ResultJson() *ChangeItems {
	s.ensureCompared()
	if !s.HasChange() {
		return nil
	}
	var result = ChangeItems{Old: map[string]any{}, New: map[string]any{}}
	for _, item := range *s.Result {
		result.Old[item.Column] = item.OldValue
		result.New[item.Column] = item.NewValue
	}
	return &result
}

func (s *DiffBox[T]) HasChange() bool {
	return len(*s.Result) > 0
}

func (s *DiffBox[T]) With(callback func(diff *DiffBox[T])) {
	if s.Old == nil || s.New == nil {
		return
	}

	callback(s)
}

func (s *DiffBox[T]) DiffLog(tx *gorm.DB, result any) {
	var data T
	ptrValue := reflect.ValueOf(&data)
	method := ptrValue.MethodByName("TableName")

	if !method.IsValid() {
		fmt.Println("未找到 TableName 方法")
		return
	}

	if method.Type().NumIn() != 0 {
		fmt.Println("TableName 方法不能带参数")
		return
	}

	out := method.Call(nil)
	if len(out) > 0 {
		tableName := out[0].Interface()
		pk := reflect.ValueOf(s.Old).Elem().FieldByName("Id").Interface()
		tx.Where("target=? AND `pk`=?", tableName, pk).Order("id desc").Find(result)
	}

}
