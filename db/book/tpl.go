// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/7/9
// 描述：
// *****************************************************************************

package book

const BookTpl = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Name }} 数据字典</title>
  <style>
    body {
      margin: 0 50px;
      font-size: 14px;
      padding-bottom: 50px;
    }

	h2 {	
	  margin-top: 15px;
	  margin-bottom:5px;
	}

    table {
      border-collapse: collapse;
      width: 100%;
    }

    th,
    td {
      border: 1px solid #ddd;
      padding: 6px;
      text-align: left;
    }

    th {
      background-color: #f2f2f2;
    }

    .main {
      display: flex;
      height: 85vh;
    }

    .menu {
      margin-right: 50px;
      height: 100%;
      overflow: auto;
    }

    .menu a {
      display: flex;
      color: #2196f3;
      font-size: 12px;
      text-decoration: inherit;
    }

    .menu a div:nth-child(1) {
      flex: 1;
    }

    .menu a div:nth-child(2) {
      color: #ccc;
      margin: 0 10px;
    }

    .table {
      flex: 1;
      height: 100%;
      overflow: auto;
    }
  </style>
</head>
<body>
  <h1>{{ .Name }} 数据字典<span style="float:right">{{ .ReleaseTime }}</span></h1>
  <div class="main">
	  <div class="menu">
	  {{ range .Tables }}
	  <a href="#{{ .Name }}">
      	<div>{{ .Name }}</div>
        <div>{{ .Comment }}</div>
      </a>
	  {{ end }}
	  </div>

  	  <div class="table">
	  {{ range .Tables }}
	  <h2 id="{{ .Name }}">{{ .Name }} {{ .Comment }}</h2>
	
	  <table>
		<tr>
		  <th>字段名</th>
		  <th>数据类型</th>
		  <th>允许空值</th>
		  <th>键</th>
		  <th>默认值</th>
		  <th>备注</th>
		</tr>
		{{ range .Columns }}
		<tr>
		  <td><div>{{ .Field }}</div><div style="font-size:12px;color:#ccc">{{ .Field2 }}</div></td>
		  <td>{{ .Type }}</td>
		  <td>{{ .Null }}</td>
		  <td>{{ .Key }}</td>
		  <td>{{ .Default }}</td>
		  <td>{{ .Comment }}</td>
		</tr>
		{{ end }}
	  </table>
	
	  {{ end }}
	  </div>
  </div>
</body>
</html>
`
