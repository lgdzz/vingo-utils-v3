// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/10
// 描述：
// *****************************************************************************

package region

import "github.com/duke-git/lancet/v2/slice"

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
