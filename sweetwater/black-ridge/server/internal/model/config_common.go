package model

import (
	"os"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/maskutil"
)

// 以下类型从 cradle 共享库重导出，保持 model 包的 API 兼容性

// ConfigLayer 配置层类型别名（从 cradle 重导出）
type ConfigLayer = configutil.ConfigLayer
// ConfigFile 配置文件类型别名（从 cradle 重导出）
type ConfigFile = configutil.ConfigFile
// SessionSchema Session 配置模式类型别名（从 cradle 重导出）
type SessionSchema = configutil.SessionSchema
// EnvDef 环境变量定义类型别名（从 cradle 重导出）
type EnvDef = configutil.EnvDef
// FileDef 文件定义类型别名（从 cradle 重导出）
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
