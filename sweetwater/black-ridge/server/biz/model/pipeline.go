package model

import (
	"github.com/charviki/maze-cradle/pipeline"
)

// 从 cradle 共享库重导出 Pipeline 类型，保持 model 包的 API 兼容性
type PipelinePhase = pipeline.PipelinePhase
type PipelineStepType = pipeline.PipelineStepType
type PipelineStep = pipeline.PipelineStep
type Pipeline = pipeline.Pipeline

const (
	PhaseSystem   = pipeline.PhaseSystem
	PhaseTemplate = pipeline.PhaseTemplate
	PhaseUser     = pipeline.PhaseUser
)

const (
	StepCD      = pipeline.StepCD
	StepEnv     = pipeline.StepEnv
	StepFile    = pipeline.StepFile
	StepCommand = pipeline.StepCommand
)
