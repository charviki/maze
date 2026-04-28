package protocol

import "time"

// AuditLogEntry 审计日志条目，记录经过 Manager 代理的所有操作
type AuditLogEntry struct {
	// ID 日志条目唯一标识
	ID string `json:"id"`
	// Timestamp 操作发生时间
	Timestamp time.Time `json:"timestamp"`
	// Operator 操作发起者（暂时为 "frontend"，后续可扩展为用户 ID）
	Operator string `json:"operator"`
	// TargetNode 目标 Agent 节点名称
	TargetNode string `json:"target_node"`
	// Action 操作类型（如 "list_sessions", "create_session", "delete_session"）
	Action string `json:"action"`
	// PayloadSummary 请求体摘要（截断到 200 字符，避免日志膨胀）
	PayloadSummary string `json:"payload_summary"`
	// Result 操作结果（"success" 或 "error: <message>"）
	Result string `json:"result"`
	// StatusCode HTTP 响应状态码
	StatusCode int `json:"status_code"`
}
