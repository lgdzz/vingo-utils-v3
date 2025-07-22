package db

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"reflect"
	"strings"
)

//type DiffBox struct {
//	Old    any
//	New    any
//	Result *map[string]DiffItem
//}
//
//type DiffItem struct {
//	Column   string
//	OldValue any
//	NewValue any
//	Message  string
//}
//
//func (s *DiffItem) SetMessage() {
//	s.Message = fmt.Sprintf("将%v的值[%v]变更为[%v]；", s.Column, s.OldValue, s.NewValue)
//}
//
//// SetNew 设置新值
//func (s *DiffBox[T]) SetNew(newValue any) {
//	s.New = newValue
//}
//
//// SetNewAndCompare 设置新值并执行比对，支持回调
//func (s *DiffBox[T]) SetNewAndCompare(newValue any, result func(diff *DiffBox)) {
//	s.SetNew(newValue)
//	s.Compare()
//	if result != nil {
//		result(s)
//	}
//}
//
//// ensureCompared 保证 Compare 已执行
//func (s *DiffBox[T]) ensureCompared() {
//	if s.Result == nil {
//		s.Compare()
//	}
//}
//
//// Compare 比较 Old 和 New
//func (s *DiffBox[T]) Compare() {
//	if s.Old == nil || s.New == nil {
//		return
//	}
//	result := map[string]DiffItem{}
//	oldVal := reflect.ValueOf(s.Old)
//	newVal := reflect.ValueOf(s.New)
//	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
//		return
//	}
//	oldType := oldVal.Type()
//	newType := newVal.Type()
//	if oldType != newType {
//		return
//	}
//	for i := 0; i < oldVal.NumField(); i++ {
//		fieldType := oldType.Field(i)
//		name := fieldType.Name
//
//		if name == "Diff" || slice.Contain([]string{"CreatedAt", "UpdatedAt", "DeletedAt"}, name) {
//			continue
//		}
//
//		oldField := oldVal.Field(i)
//		newField := newVal.Field(i)
//
//		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
//			diffItem := DiffItem{
//				Column:   name,
//				OldValue: oldField.Interface(),
//				NewValue: newField.Interface(),
//			}
//			diffItem.SetMessage()
//			result[diffItem.Column] = diffItem
//		}
//	}
//	s.Result = &result
//}
//
//// IsChange 判断某字段是否变更
//func (s *DiffBox[T]) IsChange(column string) bool {
//	s.ensureCompared()
//	_, ok := (*s.Result)[column]
//	return ok
//}
//
//// IsModify 如果字段变更则执行回调
//func (s *DiffBox[T]) IsModify(column string, callback func()) {
//	if s.IsChange(column) && callback != nil {
//		callback()
//	}
//}
//
//// IsChangeOr 任意字段变更则返回 true
//func (s *DiffBox[T]) IsChangeOr(columns ...string) bool {
//	for _, column := range columns {
//		if s.IsChange(column) {
//			return true
//		}
//	}
//	return false
//}
//
//// IsChangeAnd 所有字段都变更才返回 true
//func (s *DiffBox[T]) IsChangeAnd(columns ...string) bool {
//	for _, column := range columns {
//		if !s.IsChange(column) {
//			return false
//		}
//	}
//	return true
//}
//
//// ResultContent 生成变更摘要
//func (s *DiffBox[T]) ResultContent() string {
//	s.ensureCompared()
//	if !s.HasChange() {
//		return "无修改"
//	}
//	var builder strings.Builder
//	for _, item := range *s.Result {
//		builder.WriteString(item.Message)
//	}
//	return builder.String()
//}
//
//func (s *DiffBox[T]) HasChange() bool {
//	return len(*s.Result) > 0
//}
//
//func (s *DiffBox[T]) With(callback func(diff *DiffBox)) {
//	if s.Old == nil || s.New == nil {
//		return
//	}
//
//	callback(s)
//}

type DiffBoxInterface interface {
	Compare()
	HasChange() bool
	ResultContent() string
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

func (s *DiffItem) SetMessage() {
	s.Message = fmt.Sprintf("将%v的值[%v]变更为[%v]；", s.Column, s.OldValue, s.NewValue)
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

func (s *DiffBox[T]) HasChange() bool {
	return len(*s.Result) > 0
}

func (s *DiffBox[T]) With(callback func(diff *DiffBox[T])) {
	if s.Old == nil || s.New == nil {
		return
	}

	callback(s)
}
