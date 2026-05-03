package transport

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

const (
	defaultAuditPageSize = 50
	defaultAuditPage     = 1
	maxAuditEntries      = 10000
)

// AuditLogger 审计日志记录器，append-only JSON Lines 文件持久化。
// 所有经过 Manager 代理的操作都会被记录，提供完整的操作可追溯性。
// Manager 重启后可从文件恢复历史审计日志。
type AuditLogger struct {
	logs       []protocol.AuditLogEntry
	file       *os.File
	mu         sync.Mutex
	logger     logutil.Logger
	maxEntries int
}

// NewAuditLogger 创建审计日志记录器。
// 如果 filePath 非空，从文件加载历史日志，并打开文件用于后续 append 写入。
// 如果 filePath 为空，降级为纯内存模式（不持久化）。
func NewAuditLogger(filePath string, logger ...logutil.Logger) *AuditLogger {
	a := &AuditLogger{
		logs:       make([]protocol.AuditLogEntry, 0),
		maxEntries: maxAuditEntries,
	}
	if len(logger) > 0 {
		a.logger = logger[0]
	}

	if filePath == "" {
		return a
	}

	// 从已有文件加载历史日志
	if data, err := os.ReadFile(filepath.Clean(filePath)); err == nil {
		for _, line := range bytes.Split(data, []byte("\n")) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			var entry protocol.AuditLogEntry
			if err := json.Unmarshal(line, &entry); err == nil {
				a.logs = append(a.logs, entry)
			}
		}
	}

	// 打开文件用于 append 写入（O_CREATE: 不存在时创建，O_APPEND: 追加写入）
	f, err := os.OpenFile(filepath.Clean(filePath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		// 文件打开失败不阻塞启动，降级为纯内存模式
		return a
	}
	a.file = f
	return a
}

// Close 关闭审计日志文件句柄，确保缓冲区数据刷盘。
// 优雅关闭时调用，防止 last write 丢失。
func (a *AuditLogger) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file != nil {
		_ = a.file.Sync()
		_ = a.file.Close()
		a.file = nil
	}
}

// generateAuditID 使用 crypto/rand 生成审计日志唯一 ID，避免高并发下的碰撞风险
func generateAuditID() string {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if err != nil || n != len(b) {
		// rand.Read 失败时 fallback：使用时间戳拼接部分随机字节
		binary.BigEndian.PutUint64(b, uint64(time.Now().UnixNano()))
	}
	return "audit-" + hex.EncodeToString(b)
}

// Log 记录一条审计日志（自动填充 ID 和 Timestamp），同时 append 到文件
func (a *AuditLogger) Log(entry protocol.AuditLogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.ID == "" {
		entry.ID = generateAuditID()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.logs = append(a.logs, entry)

	if len(a.logs) > a.maxEntries {
		a.logs = a.logs[len(a.logs)-a.maxEntries:]
	}

	// 持久化到 append-only 文件（JSON Lines 格式，每行一条记录）
	if a.file != nil {
		line, err := json.Marshal(entry)
		if err == nil {
			if _, writeErr := a.file.Write(append(line, '\n')); writeErr != nil && a.logger != nil {
				a.logger.Warnf("[audit] persist log failed: %v", writeErr)
			}
		}
	}
}

// List 返回所有审计日志（按时间倒序）
func (a *AuditLogger) List() []protocol.AuditLogEntry {
	result := make([]protocol.AuditLogEntry, len(a.logs))
	copy(result, a.logs)

	// 倒序排列，最新的在前
	reverseEntries(result)
	return result
}

// ListPage 返回分页审计日志（按时间倒序，最新的在前）
// page 从 1 开始，pageSize 默认 50
func (a *AuditLogger) ListPage(page, pageSize int) (logs []protocol.AuditLogEntry, total int) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if page <= 0 {
		page = 1
	}

	a.mu.Lock()
	total = len(a.logs)

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		a.mu.Unlock()
		return []protocol.AuditLogEntry{}, total
	}
	if end > total {
		end = total
	}

	realStart := total - end
	realEnd := total - start

	result := make([]protocol.AuditLogEntry, realEnd-realStart)
	copy(result, a.logs[realStart:realEnd])
	a.mu.Unlock()

	reverseEntries(result)
	return result, total
}

// Query 按 target_node 或 action 过滤审计日志
func (a *AuditLogger) Query(node, action string) []protocol.AuditLogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	var result []protocol.AuditLogEntry
	for _, entry := range a.logs {
		if node != "" && !strings.Contains(entry.TargetNode, node) {
			continue
		}
		if action != "" && !strings.Contains(entry.Action, action) {
			continue
		}
		result = append(result, entry)
	}
	return result
}

func reverseEntries(s []protocol.AuditLogEntry) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
