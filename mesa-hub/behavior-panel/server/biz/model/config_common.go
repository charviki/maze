package model

import (
	"os"

	"github.com/charviki/maze-cradle/configutil"
)

// atomicWriteFile 封装 configutil.AtomicWriteFile，供 node.go 等内部调用
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	return configutil.AtomicWriteFile(path, data, perm)
}
