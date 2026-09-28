package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lgdzz/vingo-utils-v3/db"
	"github.com/lgdzz/vingo-utils-v3/vingo"
)

var Version = "dev"

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
	var goarch = "amd64"

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
	case "m_arm":
		goos = "darwin"
		osName = "mac"
		goarch = "arm64"
	case "l_arm":
		goos = "linux"
		osName = "linux"
		goarch = "arm64"
	default:
		log.Printf("❌ 不支持的编译目标: %s", value)
		return
	}

	log.Println("开始打包:", osName, goarch)

	moduleName := vingo.GetModuleName()

	outputName := fmt.Sprintf(
		"%s.%s-%s_%s",
		moduleName,
		version,
		osName,
		goarch,
	)

	if osName == "windows" {
		outputName += ".exe"
	}

	if err := os.MkdirAll("output", 0777); err != nil {
		log.Println("❌ 创建输出目录失败:", err)
		return
	}

	outputName = filepath.Join("output", outputName)

	buildTimeVersion := time.Now().Format("V2006.01.02_15.04.05")
	finalVersion := fmt.Sprintf("%s_%s", version, buildTimeVersion)

	ldflags := strings.Join([]string{
		"-s",
		"-w",
		"-X", moduleName + "/extend/config.version=" + version,
		"-X", "github.com/lgdzz/vingo-utils-v3/cli.Version=" + finalVersion,
	}, " ")

	args := []string{
		"build",
		"-trimpath",
		"-buildvcs=false",
		"-ldflags=" + ldflags,
		"-o", outputName,
	}

	log.Println("执行:", "go", strings.Join(args, " "))

	startTime := time.Now()

	cmd := exec.Command("go", args...)

	// 不修改当前程序的环境变量，只对本次编译生效
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH="+goarch,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("❌ 执行打包失败:", err)
		return
	}

	fileInfo, err := os.Stat(outputName)
	if err != nil {
		log.Println("❌ 获取打包文件信息失败:", err)
		return
	}

	log.Println("✅ 文件名称:", outputName)
	log.Println("✅ 文件大小:", vingo.FormatBytes(fileInfo.Size(), 2))
	log.Println("✅ 编译耗时:", time.Since(startTime).Round(time.Millisecond))
	log.Println("✅ 打包完成")
}
