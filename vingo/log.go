package vingo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	bufferSize    = 1000
	flushInterval = 5 * time.Second
	maxTokenSize  = 1024 * 1024 * 10
)

var (
	mu         sync.Mutex
	buffer     []string
	file       *os.File
	filename   string
	flushTimer *time.Timer
	maxAge     = 30 * 24 * time.Hour
)

var dstDir = "runtime/logs"
var Enable = true

func InitLogService(maxDay *int) {
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err = os.MkdirAll(dstDir, 0755); err != nil {
			panic(err.Error())
		}
	}

	if maxDay != nil && *maxDay > 0 {
		maxAge = time.Duration(*maxDay) * 24 * time.Hour
	}

	filename = generateFilename()
	err := createLogFile()
	if err != nil {
		panic(err)
	}
	flushTimer = time.AfterFunc(flushInterval, flush)
	go writeLoop()
	go deleteOldLogs()
}

func generateFilename() string {
	now := time.Now().Local()
	return fmt.Sprintf("%v/log_%04d%02d%02d.log", dstDir, now.Year(), now.Month(), now.Day())
}

func createLogFile() error {
	var err error
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if os.IsNotExist(err) {
		// 如果文件不存在则创建新文件
		file, err = os.Create(filename)
	}
	return err
}

func isFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func flush() {
	mu.Lock()
	defer mu.Unlock()

	if len(buffer) == 0 {
		flushTimer.Reset(flushInterval)
		return
	}

	// 获取日志文件名称
	filename = generateFilename()

	// 判断文件是否存在
	if file == nil {
		var err error
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening log file:", err)
			return
		}
	} else if !isSameFile(file.Name(), filename) {
		// 如果文件不同则关闭旧文件，创建新文件
		err := file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error closing log file:", err)
		}
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error opening log file:", err)
			return
		}
	}

	for _, message := range buffer {
		fmt.Fprintln(file, message)
	}
	buffer = buffer[:0]
	err := file.Sync()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error syncing log file:", err)
	}
	flushTimer.Reset(flushInterval)
}

func isSameFile(file1, file2 string) bool {
	info1, err := os.Stat(file1)
	if err != nil {
		return false
	}
	info2, err := os.Stat(file2)
	if err != nil {
		return false
	}
	return os.SameFile(info1, info2)
}

func writeLoop() {
	for {
		select {
		case <-flushTimer.C:
			flush()
		}
	}
}

func writeLog(message string) {
	if !Enable {
		return
	}
	mu.Lock()
	buffer = append(buffer, message)
	if len(buffer) >= bufferSize {
		mu.Unlock()
		flush()
		return
	}
	mu.Unlock()
}

func deleteOldLogs() {
	for {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				LogError(fmt.Sprintf("error walking directory: %v", err.Error()))
				return err
			}
			if !info.IsDir() && isLogFile(path) && isOldLog(info) {
				err := os.Remove(path)
				if err != nil {
					LogError(fmt.Sprintf("error removing old log file: %v", err.Error()))
				} else {
					LogInfo(fmt.Sprintf("Removed old log file: %v", path))
				}
			}
			return nil
		})
		time.Sleep(24 * time.Hour)
	}
}

func isLogFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".log"
}

func isOldLog(info os.FileInfo) bool {
	return time.Since(info.ModTime()) > maxAge
}

func Log(message string) {
	go writeLog(fmt.Sprintf("[%s] %s", time.Now().Local().Format("2006-01-02 15:04:05"), message))
}

func LogRequest(t string, message string) {
	writeLog(fmt.Sprintf("[%s][REQUEST][%v] %s", time.Now().Format("2006-01-02 15:04:05"), t, strings.ReplaceAll(message, "\n", "@n@n@n")))
}

func LogInfo(message string) {
	go writeLog(fmt.Sprintf("[%s][INFO][-] %s", time.Now().Format("2006-01-02 15:04:05"), strings.ReplaceAll(message, "\n", "@n@n@n")))
}

func LogError(message string) {
	go writeLog(fmt.Sprintf("[%s][ERROR][-] %s", time.Now().Format("2006-01-02 15:04:05"), strings.ReplaceAll(message, "\n", "@n@n@n")))
}

type LogFileItem struct {
	Source string `json:"source"`
	Size   int64  `json:"size"`
}

// GetLogFiles 获取日志文件列表
func GetLogFiles() []LogFileItem {
	var files []LogFileItem

	// Walk through the directory to find log files
	err := filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if the file has a .log extension
		if strings.HasSuffix(info.Name(), ".log") {
			files = append(files, LogFileItem{
				Source: path,
				Size:   info.Size(),
			})
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}

var (
	timeLayout = "2006-01-02 15:04:05"
	timeRegex  = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\]`)
)

// FindLogs 查询日志
func FindLogs(source string, keyword string, startTime string, endTime string) []string {
	// keyword 正则
	regex := fmt.Sprintf(`.*%s`, keyword)
	pattern := regexp.MustCompile(regex)

	// 解析时间范围
	var (
		start, end time.Time
		hasStart   = startTime != ""
		hasEnd     = endTime != ""
	)

	if hasStart {
		start, _ = time.Parse(timeLayout, startTime)
	}
	if hasEnd {
		end, _ = time.Parse(timeLayout, endTime)
	}

	var logs = make([]string, 0)

	readFile := func(path string) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, maxTokenSize), maxTokenSize)

		for scanner.Scan() {
			line := scanner.Text()

			// 1️⃣ 解析时间
			match := timeRegex.FindStringSubmatch(line)
			if len(match) == 2 {
				logTime, err := time.Parse(timeLayout, match[1])
				if err == nil {
					// 时间范围过滤
					if hasStart && logTime.Before(start) {
						continue
					}
					if hasEnd && logTime.After(end) {
						continue
					}
				}
			}

			// 2️⃣ 关键字匹配
			if keyword == "" || pattern.MatchString(line) {
				logs = append(logs, line)
			}
		}

		return scanner.Err()
	}

	if source == "" {
		err := filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			return readFile(path)
		})
		if err != nil {
			panic(err)
		}
	} else {
		if err := readFile(source); err != nil {
			panic(err)
		}
	}

	// keyword 为空时，只返回最近 1000 条
	if keyword == "" && len(logs) > 1000 {
		logs = logs[len(logs)-1000:]
	} else if len(logs) > 10000 {
		// 最多返回10000条
		logs = logs[len(logs)-10000:]
	}

	return logs
}
