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
	maxAge        = 30 * 24 * time.Hour
	maxTokenSize  = 1024 * 1024 * 10
)

var (
	mu         sync.Mutex
	buffer     []string
	file       *os.File
	filename   string
	flushTimer *time.Timer
)

var dstDir = "runtime/logs"
var Enable = true

func InitLogService() {
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err = os.MkdirAll(dstDir, 0755); err != nil {
			panic(err.Error())
		}
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
	writeLog(fmt.Sprintf("[%v][REQUEST][%v] %v", time.Now().Format("2006-01-02 15:04:05"), t, message))
}

func LogInfo(message string) {
	go writeLog(fmt.Sprintf("[%s][INFO] %s", time.Now().Format("2006-01-02 15:04:05"), message))
}

func LogError(message string) {
	go writeLog(fmt.Sprintf("[%s][ERROR] %s", time.Now().Format("2006-01-02 15:04:05"), message))
}

// 获取日志文件列表
func GetLogFiles() []string {
	var files []string

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
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
	return files
}

// 查询日志
func FindLogs(source string, keyword string) []string {
	// Create a regex pattern for matching log records
	regex := fmt.Sprintf(`.*%s`, keyword)
	pattern := regexp.MustCompile(regex)

	var logs = make([]string, 0)

	if len(source) == 0 {
		// Walk through the directory to find log files
		err := filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Open the log file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Read the log file line by line
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, maxTokenSize), maxTokenSize)
			for scanner.Scan() {
				line := scanner.Text()

				// Check if the line matches the regex pattern
				if pattern.MatchString(line) {
					logs = append(logs, line)
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			panic(err)
		}
	} else {
		// Open the log file
		file, err := os.Open(source)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Read the log file line by line
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, maxTokenSize), maxTokenSize)
		for scanner.Scan() {
			line := scanner.Text()

			// Check if the line matches the regex pattern
			if pattern.MatchString(line) {
				logs = append(logs, line)
			}
		}
	}

	// If keyword is not provided, return the latest 3 logs
	if len(keyword) == 0 && len(logs) > 50 {
		logs = logs[len(logs)-50:]
	}

	return logs
}
