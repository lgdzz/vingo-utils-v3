// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：分页查询
// *****************************************************************************

package db

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

type PageResult struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
	Items any   `json:"items"` // 返回 []T 或 []any
}

type PageLimit struct {
	Page *int `form:"page"`
	Size int  `form:"size"`
}

func (s *PageLimit) GetPage() int {
	if *s.Page > 0 {
		return *s.Page
	}
	return 1
}

func (s *PageLimit) GetSize() int {
	if s.Size > 0 {
		return s.Size
	}
	return 10
}

func (s *PageLimit) Offset() int {
	return (s.GetPage() - 1) * s.GetSize()
}

type PageOrder struct {
	Column string `form:"sortField"`
	Sort   string `form:"sortOrder"` // asc 或 desc
}

func (s *PageOrder) HandleColumn() string {
	sort := strings.ToLower(strings.TrimSpace(s.Sort))
	if sort != "asc" && sort != "desc" {
		panic("sortOrder 不合法，只允许 asc 或 desc")
	}
	if strings.ContainsAny(s.Column, " ;--") {
		panic("sortField 存在 SQL 注入风险")
	}

	segments := strings.Split(s.Column, ".")
	for i, seg := range segments {
		if seg == "" {
			panic("字段名非法")
		}
		segments[i] = "`" + seg + "`"
	}
	return fmt.Sprintf("%s %s", strings.Join(segments, "."), sort)
}

type PageQuery struct {
	Limit    PageLimit
	Order    *PageOrder
	OrderRaw *string // 原始排序规则
}

type QueryOption[T any] struct {
	Db       *gorm.DB         // 查询数据库对象
	Query    PageQuery        // 分页参数
	DefOrder *PageOrder       // 默认排序
	Orders   *[]PageOrder     // 预设排序规则
	Iteratee func(int, T) any // 映射函数（可选）
}

func (s *QueryOption[T]) BuildOrderString() string {
	if s.Query.Order == nil && (s.Orders == nil || len(*s.Orders) == 0) {
		if s.Query.OrderRaw != nil {
			return *s.Query.OrderRaw
		}
		return "`id` desc"
	}
	if s.Query.Order != nil {
		s.Orders = &[]PageOrder{*s.Query.Order}
	}
	var orders []string
	for _, item := range *s.Orders {
		orders = append(orders, item.HandleColumn())
	}
	return strings.Join(orders, ", ")
}

func NewPage[T any](option QueryOption[T]) PageResult {
	var count int64
	query := option.Db
	query.Count(&count)

	result := PageResult{
		Page:  option.Query.Limit.GetPage(),
		Size:  option.Query.Limit.GetSize(),
		Total: count,
	}

	if count == 0 {
		result.Items = []any{}
		return result
	}

	// 查询数据
	var records = make([]T, 0, result.Size)
	query = query.Order(option.BuildOrderString()).
		Limit(result.Size).
		Offset(option.Query.Limit.Offset()).
		Find(&records)

	// 转换处理
	if option.Iteratee != nil {
		list := make([]any, 0, len(records))
		for index, item := range records {
			list = append(list, option.Iteratee(index, item))
		}
		result.Items = list
	} else {
		result.Items = records
	}

	return result
}
