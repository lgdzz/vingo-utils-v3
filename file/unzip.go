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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func unzipZip(src, dst string) error {

	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {

		path := filepath.Join(dst, file.Name)

		// 防止 zip slip
		if !strings.HasPrefix(
			filepath.Clean(path),
			filepath.Clean(dst),
		) {
			continue
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		path = uniquePath(path)

		os.MkdirAll(filepath.Dir(path), 0755)

		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		dstFile, err := os.Create(path)
		if err != nil {
			srcFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)

		dstFile.Close()
		srcFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func unzipTarGz(src, dst string) error {

	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	return extractTar(gz, dst)
}

func unzipTar(src, dst string) error {

	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	return extractTar(file, dst)
}

func extractTar(reader io.Reader, dst string) error {

	tr := tar.NewReader(reader)

	for {

		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:

			os.MkdirAll(target, 0755)

		case tar.TypeReg:

			target = uniquePath(target)

			os.MkdirAll(filepath.Dir(target), 0755)

			file, err := os.Create(target)
			if err != nil {
				return err
			}

			_, err = io.Copy(file, tr)

			file.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func unzipGzip(src, dst string) error {

	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	name := strings.TrimSuffix(
		filepath.Base(src),
		".gz",
	)

	target := uniquePath(
		filepath.Join(dst, name),
	)

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, gz)

	return err
}

func commandUnzip(args []string) error {

	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"解压失败: %v %s",
			err,
			string(output),
		)
	}

	return nil
}
