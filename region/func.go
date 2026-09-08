// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/10
// 描述：
// *****************************************************************************

package region

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
)

func NewRegion(nodes []Node) Region {
	r := Region{
		Nodes: nodes,
	}
	return r
}

// InitCodeNameMap 生成以code为key，name为value的map
func (s *Region) InitCodeNameMap() *map[string]string {
	if s.CodeNameMap == nil {
		result := make(map[string]string)
		flattenNodeTreeCodeName(s.Nodes, result)
		s.CodeNameMap = &result
	}
	return s.CodeNameMap
}

// InitNameCodeMap 生成以name为key，code为value的map
func (s *Region) InitNameCodeMap() *map[string]string {
	if s.NameCodeMap == nil {
		result := make(map[string]string)
		flattenNodeTreeNameCode(s.Nodes, result, "")
		s.NameCodeMap = &result
	}
	return s.NameCodeMap
}

// GetNameByCode 按代码获取名称
func (s *Region) GetNameByCode(code string) string {
	s.InitCodeNameMap()
	if name, ok := (*s.CodeNameMap)[code]; ok {
		return name
	}
	return code
}

// GetNamesByCode 按代码获取完整名称
func (s *Region) GetNamesByCode(code string) []string {
	s.InitCodeNameMap()
	codes := parseAreaCode(code)
	names := make([]string, 0)
	for _, c := range codes {
		if name, ok := (*s.CodeNameMap)[c]; ok {
			names = append(names, name)
		}
	}
	return names
}

// GetCodeByName 按名称获取代码
func (s *Region) GetCodeByName(name string) string {
	s.InitNameCodeMap()
	// name格式：江西省/南昌市/...
	if code, ok := (*s.NameCodeMap)[name]; ok {
		return code
	}
	return name
}

// GetChildrenByCode 按代码获取子节点
func (s *Region) GetChildrenByCode(code string) []Node {
	node := s.findNodeByCode(code)
	if node == nil {
		return []Node{}
	}
	return node.Children
}

// GetNodeByCodeWithChildren 按代码获取节点及其子节点
func (s *Region) GetNodeByCodeWithChildren(code string) []Node {
	node := s.findNodeByCode(code)
	if node == nil {
		return []Node{}
	}
	return []Node{*node}
}

// GetSonCodes 按代码获取子节点的代码
func (s *Region) GetSonCodes(code string) []string {
	nodes := s.GetChildrenByCode(code)
	return slice.Map(nodes, func(index int, item Node) string {
		return item.Code
	})
}

// GetSonNames 按代码获取子节点的名称
func (s *Region) GetSonNames(code string) []string {
	nodes := s.GetChildrenByCode(code)
	return slice.Map(nodes, func(index int, item Node) string {
		return item.Name
	})
}

// GetSonNodes 按代码获取子节点
func (s *Region) GetSonNodes(code string) []NodeBase {
	nodes := s.GetChildrenByCode(code)
	return slice.Map(nodes, func(index int, item Node) NodeBase {
		return item.NodeBase
	})
}

func (s *Region) IsExist(nodes []Node, code string) bool {
	node := (&Region{Nodes: nodes}).findNodeByCode(code)
	return node != nil
}

// CodeUniqueCheck 唯一检测
// true: code 已存在
// false: code 不存在
func (s *Region) CodeUniqueCheck(code string) bool {
	return s.IsExist(s.Nodes, code)
}

// CreateWithAutoCode 在指定节点的children下面增加一个节点，code自动
// code规则：如果parentCode是6位数，则code从999开始每次-1，否则从99开始-1
// 使用生成的code前先验证code是否全局唯一
func (s *Region) CreateWithAutoCode(parentCode string, name string) {
	var code string

	if len(parentCode) == 6 {
		for i := 999; i >= 0; i-- {
			code = fmt.Sprintf("%s%03d", parentCode, i)

			if !s.CodeUniqueCheck(code) {
				break
			}

			if i == 0 {
				panic("没有可用的区域编码")
			}
		}
	} else {
		for i := 99; i >= 0; i-- {
			code = fmt.Sprintf("%s%02d", parentCode, i)

			if !s.CodeUniqueCheck(code) {
				break
			}

			if i == 0 {
				panic("没有可用的区域编码")
			}
		}
	}

	s.Create(parentCode, code, name)
}

// Create 在指定节点的children下面增加一个节点
func (s *Region) Create(parentCode string, code string, name string) {
	parentCode = strings.TrimSpace(parentCode)
	if parentCode == "" {
		panic("未知父节点区域编号")
	}
	if s.CodeUniqueCheck(code) {
		panic(fmt.Sprintf("区域编码已存在: %s", code))
	}

	node := Node{
		NodeBase: NodeBase{
			Code: code,
			Name: name,
		},
		Children: make([]Node, 0),
	}

	parent := s.findNodeByCode(parentCode)
	if parent == nil {
		panic(fmt.Sprintf("父区域不存在: %s", parentCode))
	}

	parent.Children = append(parent.Children, node)

	// 清除缓存
	s.CodeNameMap = nil
	s.NameCodeMap = nil
}

// Update 修改指定节点的名称
func (s *Region) Update(code string, name string) {
	node := s.findNodeByCode(code)
	if node == nil {
		panic(fmt.Sprintf("区域不存在: %s", code))
	}

	node.Name = name

	// 名称发生变化，需要清除 NameCodeMap
	s.NameCodeMap = nil
}

// Delete 删除指定节点
func (s *Region) Delete(code string) {
	var remove func(nodes []Node) ([]Node, bool)

	remove = func(nodes []Node) ([]Node, bool) {
		for i := range nodes {
			// 找到目标节点
			if nodes[i].Code == code {
				return append(nodes[:i], nodes[i+1:]...), true
			}

			// 从子节点继续查找
			if len(nodes[i].Children) > 0 {
				var deleted bool
				nodes[i].Children, deleted = remove(nodes[i].Children)

				if deleted {
					return nodes, true
				}
			}
		}

		return nodes, false
	}

	var deleted bool
	s.Nodes, deleted = remove(s.Nodes)

	if !deleted {
		panic(fmt.Sprintf("区域不存在: %s", code))
	}

	// 删除后清理缓存
	s.CodeNameMap = nil
	s.NameCodeMap = nil
}
