// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/7/29
// 描述：
// *****************************************************************************

package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lgdzz/vingo-utils-v3/vingo"
)

func buildArchivePath(src []string, dst string, method string) string {

	ext := getCompressExt(method)

	// 指定目标
	if dst != "" {

		// 如果用户没有带扩展名，自动补
		if filepath.Ext(dst) == "" {
			dst += ext
		}
		
		return uniquePath(dst)
	}

	// 单个文件/目录
	if len(src) == 1 {

		path := src[0]

		name := filepath.Base(path)

		// 去掉原文件扩展名
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			name = strings.TrimSuffix(
				name,
				filepath.Ext(name),
			)
		}

		return uniquePath(
			filepath.Join(
				filepath.Dir(path),
				name+ext,
			),
		)
	}

	// 多个文件随机名称
	return filepath.Join(
		filepath.Dir(src[0]),
		vingo.RandomString(5)+ext,
	)
}

func getCompressExt(method string) string {

	switch strings.ToLower(method) {

	case "tar":
		return ".tar"

	case "tar.gz", "tgz":
		return ".tar.gz"

	case "zip", "":
		return ".zip"

	default:
		return ".zip"
	}
}

func zipFileCreate(src []string, dst string) (string, error) {

	dst = uniquePath(dst)

	file, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	for _, path := range src {

		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}

		base := filepath.Base(path)

		if info.IsDir() {

			err = zipDir(
				writer,
				path,
				base,
			)

		} else {

			err = writeZipFile(
				writer,
				path,
				base,
			)
		}

		if err != nil {
			return "", err
		}
	}

	return dst, nil
}

func writeZipFile(
	writer *zip.Writer,
	src string,
	name string,
) error {

	entry, err := writer.Create(name)

	if err != nil {
		return err
	}

	file, err := os.Open(src)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(
		entry,
		file,
	)

	return err
}

func zipDir(
	writer *zip.Writer,
	dir string,
	base string,
) error {

	return filepath.Walk(
		dir,
		func(
			path string,
			info os.FileInfo,
			err error,
		) error {

			if err != nil {
				return err
			}

			rel, err := filepath.Rel(
				dir,
				path,
			)

			if err != nil {
				return err
			}

			name := filepath.Join(
				base,
				rel,
			)

			if info.IsDir() {

				if rel == "." {
					return nil
				}

				_, err := writer.Create(
					name + "/",
				)

				return err
			}

			return writeZipFile(
				writer,
				path,
				name,
			)
		},
	)
}

func tarCreate(src []string, dst string) (string, error) {

	dst = uniquePath(dst)

	file, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := tar.NewWriter(file)
	defer writer.Close()

	for _, path := range src {

		err := writeTar(
			writer,
			path,
			filepath.Base(path),
		)

		if err != nil {
			return "", err
		}
	}

	return dst, nil
}

func tarGzCreate(src []string, dst string) (string, error) {

	dst = uniquePath(dst)

	file, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()

	writer := tar.NewWriter(gz)
	defer writer.Close()

	for _, path := range src {

		err := writeTar(
			writer,
			path,
			filepath.Base(path),
		)

		if err != nil {
			return "", err
		}
	}

	return dst, nil
}

func writeTar(
	writer *tar.Writer,
	src string,
	base string,
) error {

	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	if info.IsDir() {

		return filepath.Walk(
			src,
			func(
				path string,
				fileInfo os.FileInfo,
				err error,
			) error {

				if err != nil {
					return err
				}

				rel, err := filepath.Rel(
					src,
					path,
				)

				if err != nil {
					return err
				}

				name := filepath.Join(
					base,
					rel,
				)

				header, err := tar.FileInfoHeader(
					fileInfo,
					"",
				)

				if err != nil {
					return err
				}

				header.Name = filepath.ToSlash(name)

				err = writer.WriteHeader(header)

				if err != nil {
					return err
				}

				if fileInfo.IsDir() {
					return nil
				}

				return copyTarContent(
					writer,
					path,
				)
			},
		)
	}

	// 单文件

	header, err := tar.FileInfoHeader(
		info,
		"",
	)

	if err != nil {
		return err
	}

	header.Name = filepath.ToSlash(base)

	err = writer.WriteHeader(header)

	if err != nil {
		return err
	}

	return copyTarContent(
		writer,
		src,
	)
}

func copyTarContent(
	writer io.Writer,
	src string,
) error {

	file, err := os.Open(src)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(
		writer,
		file,
	)

	return err
}
