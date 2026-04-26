package model

import (
	"os"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/maskutil"
)

// 以下类型从 cradle 共享库重导出，保持 model 包的 API 兼容性

type ConfigLayer = configutil.ConfigLayer
type ConfigFile = configutil.ConfigFile
type SessionSchema = configutil.SessionSchema
type EnvDef = configutil.EnvDef
type FileDef = configutil.FileDef

// MergeConfigLayers 将多个配置层合并，后者覆盖前者
var MergeConfigLayers = configutil.MergeConfigLayers

// MaskedValue 对敏感值做脱敏处理
var MaskedValue = maskutil.MaskedValue

// AtomicWriteFile 原子写入文件
var AtomicWriteFile = configutil.AtomicWriteFile

// atomicWriteFile 保持内部调用的兼容性
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	return configutil.AtomicWriteFile(path, data, perm)
}
