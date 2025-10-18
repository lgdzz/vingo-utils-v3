// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/10
// 描述：
// *****************************************************************************

package region

type Region struct {
	Nodes       []Node             `json:"nodes"`
	CodeNameMap *map[string]string `json:"codeNameMap"`
	NameCodeMap *map[string]string `json:"nameCodeMap"`
}

type NodeBase struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Node struct {
	NodeBase
	Children []Node `json:"children"`
}

// 将字符串区域编码转成数组
func parseAreaCode(code string) []string {
	splitPoints := []int{2, 4, 6, 9}
	var result []string

	for _, p := range splitPoints {
		if len(code) >= p {
			result = append(result, code[:p])
		}
	}
	return result
}

func flattenNodeTreeCodeName(nodes []Node, result map[string]string) {
	for _, node := range nodes {
		result[node.Code] = node.Name
		if len(node.Children) > 0 {
			flattenNodeTreeCodeName(node.Children, result)
		}
	}
}

func flattenNodeTreeNameCode(nodes []Node, result map[string]string, prefix string) {
	for _, node := range nodes {
		var fullName string
		if prefix != "" {
			fullName = prefix + "/" + node.Name
		} else {
			fullName = node.Name
		}

		result[fullName] = node.Code

		if len(node.Children) > 0 {
			flattenNodeTreeNameCode(node.Children, result, fullName)
		}
	}
}

// 私有方法：用于获取指定 code 的节点指针
func (s *Region) findNodeByCode(code string) *Node {
	path := parseAreaCode(code)
	current := s.Nodes

	for _, p := range path {
		found := false
		for _, v := range current {
			if v.Code == p {
				if p == code {
					return &v
				}
				current = v.Children
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return nil
}
