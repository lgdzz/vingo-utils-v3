// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/22
// 描述：
// *****************************************************************************

package pathutil

import (
	"fmt"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/lgdzz/vingo-utils-v3/db"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// JoinField 自动拼接字段配置
type JoinField struct {
	Target string // 要设置的目标字段（最终拼接数据）
	Self   string // 当前节点字段（具体产生数据）
	Sep    string // 分隔符，默认 "."
}

// Option 设置路径选项
type Option struct {
	Tx *gorm.DB

	FieldId   string // 默认 "Id"
	FieldPid  string // 默认 "Pid"
	FieldPath string // 默认 "Path"
	FieldLen  string // 默认 "Len"

	JoinFields []JoinField
}

func getFieldSafe(v reflect.Value, name string) (reflect.Value, bool) {
	f := v.FieldByName(name)
	return f, f.IsValid() && f.CanSet()
}

func setStringFieldSafe(v reflect.Value, name, val string) {
	if f, ok := getFieldSafe(v, name); ok && f.Kind() == reflect.String {
		f.SetString(val)
	}
}

func setIntFieldSafe(v reflect.Value, name string, val int64) {
	if f, ok := getFieldSafe(v, name); ok {
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(val)
		default:
			panic("unhandled default case")
		}
	}
}

func getIDString(v reflect.Value, field string) string {
	f := v.FieldByName(field)
	switch f.Kind() {
	case reflect.String:
		return f.String()
	case reflect.Int, reflect.Int64, reflect.Int32:
		return strconv.FormatInt(f.Int(), 10)
	case reflect.Uint, reflect.Uint64, reflect.Uint32:
		return strconv.FormatUint(f.Uint(), 10)
	default:
		return ""
	}
}

func hasParent(v reflect.Value, field string) bool {
	f := v.FieldByName(field)
	switch f.Kind() {
	case reflect.String:
		return f.String() != ""
	case reflect.Int, reflect.Int64:
		return f.Int() > 0
	case reflect.Uint, reflect.Uint64:
		return f.Uint() > 0
	default:
		return false
	}
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func setPathOfChild[T any](model *T, option *Option) {
	s := reflect.ValueOf(model).Elem()

	var children []T
	option.Tx.Find(&children, fmt.Sprintf("%v = ?", strutil.CamelCase(option.FieldPid)), getIDString(s, option.FieldId))
	for _, child := range children {
		SetPathWithCreate[T](&child, model, option)
		setPathOfChild[T](&child, option)
	}
}

// SetPathWithCreate 设置数据路径，上下级数据结构包含（path、len）字段使用
func SetPathWithCreate[T any](model *T, parent *T, option *Option) {
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		panic("SetPath: model must be a pointer")
	}
	s := reflect.ValueOf(model).Elem()

	// 设置默认字段名
	if option.FieldId == "" {
		option.FieldId = "Id"
	}
	if option.FieldPid == "" {
		option.FieldPid = "Pid"
	}
	if option.FieldPath == "" {
		option.FieldPath = "Path"
	}
	if option.FieldLen == "" {
		option.FieldLen = "Len"
	}

	if hasParent(s, option.FieldPid) {
		if parent == nil {
			pid := getIDString(s, option.FieldPid)
			parent = pointer.Of(db.Find[T](option.Tx, fmt.Sprintf("%v = ?", strutil.CamelCase(option.FieldId)), pid))
		}
		parentValue := reflect.ValueOf(parent).Elem()
		path := fmt.Sprintf("%v,%v", getIDString(parentValue, option.FieldPath), getIDString(s, option.FieldId))
		setStringFieldSafe(s, option.FieldPath, path)
		setIntFieldSafe(s, option.FieldLen, parentValue.FieldByName(option.FieldLen).Int()+1)
	} else {
		setStringFieldSafe(s, option.FieldPath, getIDString(s, option.FieldId))
		setIntFieldSafe(s, option.FieldLen, 1)
	}

	// 自动处理 JoinFields
	for _, jf := range option.JoinFields {
		sep := jf.Sep
		if sep == "" {
			sep = ","
		}
		selfField := s.FieldByName(jf.Self)
		var val string
		if hasParent(s, option.FieldPid) && parent != nil {
			parentValue := reflect.ValueOf(parent).Elem()
			parentField := parentValue.FieldByName(jf.Target)
			val = fmt.Sprintf("%v%v%v", parentField.String(), sep, selfField.String())
		} else {
			val = selfField.String()
		}
		setStringFieldSafe(s, jf.Target, val)
	}

	// 构建更新字段列表
	fields := []string{
		lowerFirst(option.FieldPath),
		lowerFirst(option.FieldLen),
	}
	for _, jf := range option.JoinFields {
		fields = append(fields, lowerFirst(jf.Target))
	}

	// 构造更新字段 map
	updateMap := map[string]interface{}{}
	for _, field := range fields {
		fv := s.FieldByName(upperFirst(field))
		if fv.IsValid() {
			updateMap[field] = fv.Interface()
		}
	}
	option.Tx.Model(model).Select(fields).UpdateColumns(updateMap)
}

// SetPathWithUpdate 设置路径，更新所有子级路径
func SetPathWithUpdate[T any](model *T, option Option) {
	SetPathWithCreate[T](model, nil, &option)
	setPathOfChild[T](model, &option)
}
