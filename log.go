package main

import (
	"log"

	"github.com/sirupsen/logrus"
)

func init() {
	// 设置日志级别
	logrus.SetLevel(logrus.DebugLevel)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
