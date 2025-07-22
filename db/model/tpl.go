// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/14
// 描述：
// *****************************************************************************

package model

const ModelTpl = `// *****************************************************************************
// 作者: lgdz
// 创建时间: {{ .Date }}
// 描述：{{ .TableComment }}
// *****************************************************************************

package model

import(
	"github.com/lgdzz/vingo-utils-v3/db"
	"github.com/lgdzz/vingo-utils-v3/moment"
	"gorm.io/gorm"
)

type {{ .ModelName }} struct {
	{{ range .TableColumns }}{{ .DataName }}   {{ .DataType }}  ` + "`gorm:\"{{ if eq .Key \"PRI\" }}primaryKey;{{ end }}column:{{ .Field }}\" json:\"{{ .JsonName }}\"`" + ` {{ if .Comment }}// {{ .Comment }}{{ end }}
	{{ end }}
	Diff db.DiffBox ` + "`gorm:\"-\" json:\"-\"`" + `
}

func (s *{{ .ModelName }}) TableName() string {
	return "{{ .TableName }}"
}

type {{ .ModelName }}Query struct {
	*db.PageQuery
}
`
