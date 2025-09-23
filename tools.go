package main

import (
	"fmt"
	"os"
)

// FileExists 检查文件是否存在于文件系统中
//
// path: 要检查的文件路径
//
// returns: 如果文件存在则返回true，否则返回false
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SlicesLast 返回切片的最后一个元素，如果切片为空则返回 nil
func SlicesLast[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[len(slice)-1]
}

func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// 检查文件是否已存在
	if FileExists(dst) {
		fmt.Printf("dst already exists, skipping: %s\n", dst)
		return nil
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {

		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("Successfully copied file from [%s] to [%s] \n", src, dst)

	return nil
}
