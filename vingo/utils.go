// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/27
// 描述：
// *****************************************************************************

package vingo

import (
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/formatter"
	"github.com/google/uuid"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Print 格式化打印输出
func Print(content any) {
	result, _ := formatter.Pretty(content)
	fmt.Println(result)
}

// Of 返回传入参数的指针
func Of[T any](v T) *T {
	return &v
}

// GetModuleName 获取当前项目模块名称(mod-name)
func GetModuleName() (name string) {
	// 获取当前项目的根目录路径
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录路径：", err)
		return
	}

	// 执行go mod命令获取模块名称
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = rootDir
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("无法获取模块名称：", err)
		return
	}

	// 解析输出结果，获取模块名称
	name = strings.TrimSpace(string(output))

	return
}

// SY 三元运算
func SY[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	} else {
		return falseValue
	}
}

// GetCurrentFunctionName 获取当前函数名
func GetCurrentFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	currentFunction := runtime.FuncForPC(pc).Name()
	return currentFunction
}

// GetUUID 生成UUID
func GetUUID() string {
	return uuid.NewString()
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// RandomNumber 生成随机数
func RandomNumber(length int) string {
	digits := []rune("0123456789")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = digits[r.Intn(len(digits))]
	}
	return string(b)
}

// OrderNo 生成按时间+随机数的单号
func OrderNo(length int, check func(string) bool) string {
	if length <= 14 {
		panic("编号长度不少于15位")
	}
	orderNo := fmt.Sprintf("%v%v", time.Now().Format("20060102150405"), RandomNumber(length-14))
	if check != nil && check(orderNo) {
		// 已存在，重新生成
		return OrderNo(length, check)
	}
	return strings.ToUpper(orderNo)
}

// OrderNoPrefix 生成按时间+随机数的单号
func OrderNoPrefix(prefix string, length int, check func(string) bool) string {
	if length <= 14 {
		panic("编号长度不少于15位")
	}
	orderNo := fmt.Sprintf("%v%v%v", prefix, time.Now().Format("20060102150405"), RandomNumber(length-14))
	if check != nil && check(orderNo) {
		// 已存在，重新生成
		return OrderNo(length, check)
	}
	return strings.ToUpper(orderNo)
}

// ComputeGrowRate 增长率计算
// now 现在值
// prev 过去值
func ComputeGrowRate(now float64, prev float64) string {
	if now == prev {
		return "0.00"
	} else if prev == 0 {
		return "-"
	} else {
		return fmt.Sprintf("%.2f", ((now - prev) / prev * 100))
	}
}

// Convert 结构类型转换
func Convert[T any](input any) T {
	var out T
	data, err := json.Marshal(input)
	if err != nil {
		panic(fmt.Sprintf("marshal error: %v", err))
	}
	if err = json.Unmarshal(data, &out); err != nil {
		panic(fmt.Sprintf("unmarshal error: %v", err))
	}
	return out
}

// JsonToString 结构体转字符串
func JsonToString(data any) string {
	output, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}
	return string(output)
}

// StringToJson 字符串转结构体
func StringToJson(data string, output any) {
	err := json.Unmarshal([]byte(data), &output)
	if err != nil {
		panic(err.Error())
	}
}

func ToInt(value any) int {
	return int(ToInt64(value))
}

func ToInt64(value any) int64 {
	v, _ := convertor.ToInt(value)
	return v
}

func ToUint(value any) uint {
	return uint(ToInt64(value))
}

func ToFloat(value any) float64 {
	v, _ := convertor.ToFloat(value)
	return v
}

func ToBool(value string) bool {
	v, _ := convertor.ToBool(value)
	return v
}

func ToBase64(value any) string {
	return convertor.ToStdBase64(value)
}

func ToUrlBase64(value any) string {
	return convertor.ToUrlBase64(value)
}

func ToString(value any) string {
	return convertor.ToString(value)
}

func ToMap[T any, K comparable, V any](array []T, iteratee func(T) (K, V)) map[K]V {
	return convertor.ToMap(array, iteratee)
}

// FormatBytes 将字节转换为可读文本
func FormatBytes(size int64, precision int) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	i := 0
	fsize := big.NewFloat(float64(size))
	for fsize.Cmp(big.NewFloat(1024)) >= 0 && i < 6 {
		fsize.Quo(fsize, big.NewFloat(1024))
		i++
	}
	format := fmt.Sprintf("%%.%df %%s", precision)
	return fmt.Sprintf(format, fsize, units[i])
}
