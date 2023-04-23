package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
)

func getLogLevel() logrus.Level {
	// 设置日志级别，可以根据需要修改
	return logrus.DebugLevel
}

func getLogFormatter() logrus.Formatter {
	// 定义日志输出格式
	return &logrus.TextFormatter{
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05", //时间格式,
		CallerPrettyfier: callerPrettyfier,
	}
}

func callerPrettyfier(f *runtime.Frame) (string, string) {
	// 格式化输出文件名和行号
	filename := filepath.Base(f.File)
	return fmt.Sprintf("%s:%d", filename, f.Line), ""
}

func NewLogger() *logrus.Logger {
	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(getLogLevel())
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(getLogFormatter())

	return logger
}
