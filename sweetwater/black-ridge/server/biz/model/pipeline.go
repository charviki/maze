package model

import (
	"github.com/charviki/maze-cradle/pipeline"
)

// PipelinePhase 管线步骤层级（从 cradle 重导出）
type PipelinePhase = pipeline.PipelinePhase
// PipelineStepType 管线步骤类型（从 cradle 重导出）
type PipelineStepType = pipeline.PipelineStepType
// PipelineStep 管线步骤（从 cradle 重导出）
type PipelineStep = pipeline.PipelineStep
// Pipeline 管线列表（从 cradle 重导出）
type Pipeline = pipeline.Pipeline

const (
	// PhaseSystem 系统级管线阶段
	PhaseSystem   = pipeline.PhaseSystem
	// PhaseTemplate 模板级管线阶段
	PhaseTemplate = pipeline.PhaseTemplate
	// PhaseUser 用户级管线阶段
	PhaseUser     = pipeline.PhaseUser
)

const (
	// StepCD 切换目录步骤
	StepCD      = pipeline.StepCD
	// StepEnv 设置环境变量步骤
	StepEnv     = pipeline.StepEnv
	// StepFile 写入文件步骤
	StepFile    = pipeline.StepFile
	// StepCommand 执行命令步骤
	StepCommand = pipeline.StepCommand
)
