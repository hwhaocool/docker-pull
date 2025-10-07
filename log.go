package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func init() {
	// 设置日志级别
	logrus.SetLevel(logrus.DebugLevel)

	tf := &logrus.TextFormatter{
		ForceColors:     true, // 强制颜色输出
		FullTimestamp:   true, // 完整时间戳
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	}

	// 设置单行文本格式
	Logger.SetFormatter(tf)
	logrus.SetFormatter(tf)

	Logger.SetOutput(os.Stdout)

}
