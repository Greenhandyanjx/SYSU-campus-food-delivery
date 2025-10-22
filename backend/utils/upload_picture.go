package utils

import (
	"os"
	"path/filepath"
)

// EnsureDir 确保目录存在
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, os.ModePerm)
}