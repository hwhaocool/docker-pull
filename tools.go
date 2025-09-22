package main

import "os"

// FileExists 检查文件是否存在于文件系统中
//
// path: 要检查的文件路径
//
// returns: 如果文件存在则返回true，否则返回false
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
