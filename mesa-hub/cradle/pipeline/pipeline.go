package pipeline

// PipelinePhase 管线步骤层级
type PipelinePhase string

const (
	PhaseSystem   PipelinePhase = "system"   // 系统级，不可修改
	PhaseTemplate PipelinePhase = "template" // 模板级，可查看可调整
	PhaseUser     PipelinePhase = "user"     // 用户级，完全自定义
)

// PipelineStepType 步骤类型
type PipelineStepType string

const (
	StepCD      PipelineStepType = "cd"      // 切换目录
	StepEnv     PipelineStepType = "env"     // 设置环境变量
	StepFile    PipelineStepType = "file"    // 写入文件
	StepCommand PipelineStepType = "command" // 执行命令
)

// PipelineStep 管线步骤
type PipelineStep struct {
	ID    string           `json:"id"`
	Type  PipelineStepType `json:"type"`
	Phase PipelinePhase    `json:"phase"`
	Order int              `json:"order"`
	Key   string           `json:"key,omitempty"` // cd: 目录路径; env: 变量名; file: 文件路径; command: 空
	Value string           `json:"value"`         // cd: 空; env: 变量值; file: 文件内容; command: shell 命令
}

// Pipeline 管线（有序步骤列表）
type Pipeline []PipelineStep

// SystemSteps 过滤 system 层步骤
func (p Pipeline) SystemSteps() Pipeline {
	return p.filterByPhase(PhaseSystem)
}

// TemplateSteps 过滤 template 层步骤
func (p Pipeline) TemplateSteps() Pipeline {
	return p.filterByPhase(PhaseTemplate)
}

// UserSteps 过滤 user 层步骤
func (p Pipeline) UserSteps() Pipeline {
	return p.filterByPhase(PhaseUser)
}

// filterByPhase 按层级过滤步骤，避免三个公开方法重复遍历逻辑
func (p Pipeline) filterByPhase(phase PipelinePhase) Pipeline {
	var result Pipeline
	for _, step := range p {
		if step.Phase == phase {
			result = append(result, step)
		}
	}
	return result
}

// Sorted 按 Order 字段排序返回新管线，使用稳定排序以保持相同 Order 步骤的原始顺序
func (p Pipeline) Sorted() Pipeline {
	sorted := make(Pipeline, len(p))
	copy(sorted, p)
	// 插入排序：管线步骤数量通常很少，插入排序在小区间上性能优于快排
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].Order < sorted[j-1].Order; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	return sorted
}
