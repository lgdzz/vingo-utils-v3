// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/7/29
// 描述：
// *****************************************************************************

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// 目标是目录
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	// 文件存在自动改名
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

	// 目标是目录
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	dst = uniquePath(dst)

	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return "", err
	}

	err = os.Rename(src, dst)

	if err == nil {
		return dst, nil
	}

	// 跨磁盘移动，降级复制删除
	newPath, err := Copy(src, dst)
	if err != nil {
		return "", err
	}

	err = Delete(src)

	return newPath, err
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

// Zip 压缩
func Zip(src []string, dst string, method ...string) (string, error) {

	if len(src) == 0 {
		return "", fmt.Errorf("没有需要压缩的文件")
	}

	zipMethod := "zip"

	if len(method) > 0 && method[0] != "" {
		zipMethod = strings.ToLower(method[0])
	}

	dst = buildArchivePath(
		src,
		dst,
		zipMethod,
	)

	switch zipMethod {

	case "zip":
		return zipFileCreate(src, dst)

	case "tar":
		return tarCreate(src, dst)

	case "tar.gz", "tgz":
		return tarGzCreate(src, dst)

	default:
		return "", fmt.Errorf(
			"不支持的压缩格式: %s",
			zipMethod,
		)
	}
}

// Unzip 解压文件
func Unzip(src string, dst ...string) error {
	ok, _, err := isArchiveFile(src)

	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf(
			"文件不是压缩文件: %s",
			src,
		)
	}

	//fmt.Println("压缩类型:", kind)

	// 默认解压到压缩文件同目录下的同名文件夹
	targetDir := ""

	if len(dst) > 0 && dst[0] != "" {
		targetDir = dst[0]
	} else {
		dir := filepath.Dir(src)
		name := filepath.Base(src)

		// 去除压缩扩展名
		switch {
		case strings.HasSuffix(name, ".tar.gz"):
			name = strings.TrimSuffix(name, ".tar.gz")

		case strings.HasSuffix(name, ".tgz"):
			name = strings.TrimSuffix(name, ".tgz")

		case strings.HasSuffix(name, ".tar"):
			name = strings.TrimSuffix(name, ".tar")

		case strings.HasSuffix(name, ".zip"):
			name = strings.TrimSuffix(name, ".zip")

		case strings.HasSuffix(name, ".gz"):
			name = strings.TrimSuffix(name, ".gz")

		case strings.HasSuffix(name, ".rar"):
			name = strings.TrimSuffix(name, ".rar")

		case strings.HasSuffix(name, ".7z"):
			name = strings.TrimSuffix(name, ".7z")
		}

		targetDir = filepath.Join(dir, name)
	}

	// 如果目录已存在，自动改名
	targetDir = uniquePath(targetDir)

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(src)

	switch {
	case strings.HasSuffix(ext, ".zip"):
		return unzipZip(src, targetDir)

	case strings.HasSuffix(ext, ".tar.gz"),
		strings.HasSuffix(ext, ".tgz"):
		return unzipTarGz(src, targetDir)

	case strings.HasSuffix(ext, ".tar"):
		return unzipTar(src, targetDir)

	case strings.HasSuffix(ext, ".gz"):
		return unzipGzip(src, targetDir)

	case strings.HasSuffix(ext, ".7z"):
		return commandUnzip([]string{
			"7z",
			"x",
			src,
			"-o" + targetDir,
		})

	case strings.HasSuffix(ext, ".rar"):
		return commandUnzip([]string{
			"unrar",
			"x",
			src,
			targetDir,
		})

	default:
		return fmt.Errorf("不支持的压缩格式: %s", src)
	}
}
