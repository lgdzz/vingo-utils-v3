package cli

import (
	"bufio"
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

	buildNew := flag.String("build", "", "应用打包")

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

	if *buildNew != "" {
		BuildProjectNew()
	}

	if *updateVingo != "" {
		cmd := exec.Command("go", "get", "-u", "github.com/lgdzz/vingo-utils-v3@"+*updateVingo)
		_ = cmd.Run()
		os.Exit(0)
	}

}

// Deprecated: 使用 BuildProjectNew 替代。
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

// BuildProjectNew 新版打包程序
func BuildProjectNew() {
	reader := bufio.NewReader(os.Stdin)

	// 选择编译平台
	fmt.Println("请选择编译平台:")
	fmt.Println("  1) Linux AMD64")
	fmt.Println("  2) Linux ARM64")
	fmt.Println("  3) Windows AMD64")
	fmt.Println("  4) Mac AMD64")
	fmt.Println("  5) Mac ARM64")
	fmt.Print("请输入选项 [1]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		log.Println("❌ 读取编译平台失败:", err)
		return
	}

	value := strings.TrimSpace(input)
	if value == "" {
		value = "1"
	}

	var goos string
	var osName string
	var goarch = "amd64"

	switch value {
	case "1":
		goos = "linux"
		osName = "linux"

	case "2":
		goos = "linux"
		osName = "linux"
		goarch = "arm64"

	case "3":
		goos = "windows"
		osName = "windows"

	case "4":
		goos = "darwin"
		osName = "mac"

	case "5":
		goos = "darwin"
		osName = "mac"
		goarch = "arm64"

	default:
		log.Printf("❌ 不支持的选项: %s", value)
		return
	}

	// 选择版本
	fmt.Println()
	fmt.Println("请选择发行版本:")
	fmt.Println("  1) 正式版")
	fmt.Println("  2) 开发版")
	fmt.Print("请输入选项 [1]: ")

	input, err = reader.ReadString('\n')
	if err != nil {
		log.Println("❌ 读取版本失败:", err)
		return
	}

	versionOption := strings.TrimSpace(input)
	if versionOption == "" {
		versionOption = "1"
	}

	var version string

	switch versionOption {
	case "1":
		version = "prod"

	case "2":
		version = "dev"

	default:
		log.Printf("❌ 不支持的版本选项: %s", versionOption)
		return
	}

	fmt.Println()
	fmt.Println("================================")
	fmt.Println("编译平台:", osName, goarch)
	fmt.Println("编译版本:", version)
	fmt.Println("================================")
	fmt.Println()

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
