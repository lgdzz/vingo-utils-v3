// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：方法适配器
// *****************************************************************************

package oss

type Adapter interface {
	ObjectUrl(objectName string) string // 访问地址
	UploadSign(objectName string) any   // 上传签名
	Delete(objectName string) error     // 删除文件
}
