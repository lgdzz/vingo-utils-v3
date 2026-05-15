// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/5/15
// 描述：
// *****************************************************************************

package onlyoffice

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
)

func CreateEmptyDocxBase64() (string, error) {
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
    <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
    <Default Extension="xml" ContentType="application/xml"/>
    <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`,

		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1"
        Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
        Target="word/document.xml"/>
</Relationships>`,

		"word/document.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body/>
</w:document>`,
	}

	for name, content := range files {
		f, err := zipWriter.Create(name)
		if err != nil {
			return "", err
		}

		_, err = f.Write([]byte(content))
		if err != nil {
			return "", err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
