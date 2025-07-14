package cli

import (
	"flag"
	"fmt"
	"github.com/lgdzz/vingo-utils-v3/db"
	"github.com/lgdzz/vingo-utils-v3/vingo"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	Enable      bool
	DatabaseApi *db.Api
	Register    func()
}

func InitCli(options Options) {
	if !options.Enable {
		return
	}
	model := flag.String("m", "", "生成数据库模型，支持多个表生成，格式：table1,table2")

	buildDev := flag.String("build-dev", "", "打包开发版，参数：l=linux;w=windows;m=mac;l_arm=linux arm")
	buildProd := flag.String("build-prod", "", "打包正式版，参数：l=linux;w=windows;m=mac;l_arm=linux arm")

	updateVingo := flag.String("v3", "", "更新vingo-v3版本")

	if options.Register != nil {
		options.Register()
	}

	help := flag.Bool("h", false, "Show help")

	// 解析命令行参数
	flag.Parse()

	if *help {
		// 如果使用 -h 或 --help 标志，则显示帮助信息
		flag.Usage()
		os.Exit(0)
	}

	// 创建数据表模型文件
	if *model != "" {
		_, _ = options.DatabaseApi.ModelFiles(strings.Split(*model, ",")...)
		os.Exit(0)
	}

	if *buildDev != "" {
		BuildProject(*buildDev, "dev")
	}
	if *buildProd != "" {
		BuildProject(*buildProd, "prod")
	}

	if *updateVingo != "" {
		cmd := exec.Command("go", "get", "-u", "github.com/lgdzz/vingo-utils-v3@"+*updateVingo)
		_ = cmd.Run()
		os.Exit(0)
	}

}

func BuildProject(value string, version string) {
	var goos string
	var osName string
	var gOARCH = "amd64"
	switch value {
	case "l":
		goos = "linux"
		osName = "linux"
	case "w":
		goos = "windows"
		osName = "windows"
	case "m":
		goos = "darwin"
		osName = "mac"
	case "l_arm":
		goos = "linux"
		osName = "linux"
		gOARCH = "arm64"
	}
	var err error
	if err = os.Setenv("CGO_ENABLED", "0"); err != nil {
		log.Println("设置CGO_ENABLED错误：", err.Error())
		os.Exit(0)
	}
	if err = os.Setenv("GOOS", goos); err != nil {
		log.Println("设置GOOS错误：", err.Error())
		os.Exit(0)
	}
	if err = os.Setenv("GOARCH", gOARCH); err != nil {
		log.Println("设置GOARCH错误：", err.Error())
		os.Exit(0)
	}

	fmt.Println("开始打包:", osName, gOARCH)

	var moduleName = vingo.GetModuleName()
	var outputName = fmt.Sprintf("%v.%v-%v_%v", moduleName, version, osName, gOARCH)
	if osName == "windows" {
		outputName += ".exe"
	}

	_ = os.MkdirAll("output", 0777)

	outputName = filepath.Join("output", outputName)

	log.Println(strings.Join([]string{"go", "build", "-ldflags=-X " + moduleName + "/extend/config.version=" + version, "-o", outputName}, " "))

	// 执行打包命令
	cmd := exec.Command("go", "build", "-ldflags=-X "+moduleName+"/extend/config.version="+version, "-o", outputName)
	err = cmd.Run()
	if err != nil {
		log.Println("执行打包命令错误：", err.Error())
		os.Exit(0)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(outputName)
	if err != nil {
		log.Println("获取打包文件信息错误：", err.Error())
	}
	fileSize := fileInfo.Size()
	log.Println("✅ 文件名称：", outputName)
	log.Println("✅ 文件大小：", vingo.FormatBytes(fileSize, 2))
	log.Println("✅ 打包完成")
	os.Exit(0)
}
