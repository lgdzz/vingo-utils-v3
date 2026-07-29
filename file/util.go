// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/7/29
// 描述：
// *****************************************************************************

package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ---------------- 内部方法 ----------------

func copyFile(src, dst string) error {

	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	return err
}

func copyDir(src, dst string) error {

	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
		} else {

			dstPath = uniquePath(dstPath)

			err = copyFile(srcPath, dstPath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// 如果存在自动改名
func uniquePath(path string) string {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)

	base := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {

		newPath := fmt.Sprintf(
			"%s_%d%s",
			base,
			i,
			ext,
		)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
