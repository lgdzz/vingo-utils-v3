// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/7/29
// 描述：
// *****************************************************************************

package file

import (
	"os"
	"path/filepath"
)

// Mkdir 创建目录
func Mkdir(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

// Chmod 设置文件或目录权限
func Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// Copy 复制文件或文件夹
// 返回实际复制后的路径
func Copy(src, dst string) (string, error) {

	info, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	dst = uniquePath(dst)

	if info.IsDir() {
		err = copyDir(src, dst)
	} else {
		err = copyFile(src, dst)
	}

	if err != nil {
		return "", err
	}

	return dst, nil
}

// Move 剪切文件或文件夹
func Move(src, dst string) (string, error) {

	if _, err := os.Stat(src); err != nil {
		return "", err
	}

	dst = uniquePath(dst)

	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return "", err
	}

	err = os.Rename(src, dst)

	if err != nil {
		// 跨磁盘 rename 失败，降级复制删除
		newPath, err := Copy(src, dst)
		if err != nil {
			return "", err
		}

		err = Delete(src)

		return newPath, err
	}

	return dst, nil
}

// Rename 重命名
func Rename(path, newName string) (string, error) {

	dir := filepath.Dir(path)

	newPath := filepath.Join(dir, newName)

	newPath = uniquePath(newPath)

	err := os.Rename(path, newPath)

	return newPath, err
}

// Delete 删除文件或文件夹
func Delete(path string) error {

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return os.RemoveAll(path)
	}

	return os.Remove(path)
}

// ReadText 读取文本文件
func ReadText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// WriteText 写入文本文件
func WriteText(path string, content string) error {

	// 自动创建父目录
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(
		path,
		[]byte(content),
		0644,
	)
}
