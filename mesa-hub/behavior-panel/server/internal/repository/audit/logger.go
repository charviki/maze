package audit

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
	defaultPageSize = 50
	defaultPage     = 1
	maxEntries      = 10000
)

// Logger 负责持久化和查询审计日志。
// 审计日志实现收敛到 repository，是为了避免 service/transport 夹带文件持久化细节。
type Logger struct {
	logs       []protocol.AuditLogEntry
	file       *os.File
	mu         sync.RWMutex
	logger     logutil.Logger
	maxEntries int
}

// NewLogger 创建审计日志记录器。
// filePath 为空时降级为纯内存模式，保证文件故障不会阻塞 Manager 启动。
func NewLogger(filePath string, logger ...logutil.Logger) *Logger {
	a := &Logger{
		logs:       make([]protocol.AuditLogEntry, 0),
		maxEntries: maxEntries,
	}
	if len(logger) > 0 {
		a.logger = logger[0]
	}

	if filePath == "" {
		return a
	}

	// 启动时先恢复历史日志，确保重启后仍能追溯此前的代理操作。
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

	f, err := os.OpenFile(filepath.Clean(filePath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		// 审计文件不可写时退回内存模式，避免启动阶段被单点 I/O 故障拖死。
		return a
	}
	a.file = f
	return a
}

// Close 关闭文件句柄并尽量刷盘，避免优雅退出时丢失最后几条审计记录。
func (a *Logger) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file != nil {
		_ = a.file.Sync()
		_ = a.file.Close()
		a.file = nil
	}
}

// generateAuditID 使用 crypto/rand 生成审计日志唯一 ID，降低高并发下碰撞概率。
func generateAuditID() string {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if err != nil || n != len(b) {
		// 真随机失败时退回时间戳兜底，保证日志仍可持续写入。
		binary.BigEndian.PutUint64(b, uint64(time.Now().UnixNano()))
	}
	return "audit-" + hex.EncodeToString(b)
}

// Log 记录一条审计日志，并在可用时追加写入 JSON Lines 文件。
func (a *Logger) Log(entry protocol.AuditLogEntry) {
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

	if a.file != nil {
		line, err := json.Marshal(entry)
		if err == nil {
			if _, writeErr := a.file.Write(append(line, '\n')); writeErr != nil && a.logger != nil {
				a.logger.Warnf("[audit] persist log failed: %v", writeErr)
			}
		}
	}
}

// List 返回所有审计日志，按时间倒序排列。
func (a *Logger) List() []protocol.AuditLogEntry {
	a.mu.RLock()
	result := make([]protocol.AuditLogEntry, len(a.logs))
	copy(result, a.logs)
	a.mu.RUnlock()
	reverseEntries(result)
	return result
}

// ListPage 返回分页审计日志，page 从 1 开始。
func (a *Logger) ListPage(page, pageSize int) (logs []protocol.AuditLogEntry, total int) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if page <= 0 {
		page = defaultPage
	}

	a.mu.RLock()
	total = len(a.logs)

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		a.mu.RUnlock()
		return []protocol.AuditLogEntry{}, total
	}
	if end > total {
		end = total
	}

	realStart := total - end
	realEnd := total - start
	result := make([]protocol.AuditLogEntry, realEnd-realStart)
	copy(result, a.logs[realStart:realEnd])
	a.mu.RUnlock()

	reverseEntries(result)
	return result, total
}

// Query 按 target_node 或 action 过滤审计日志。
func (a *Logger) Query(node, action string) []protocol.AuditLogEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

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

func reverseEntries(entries []protocol.AuditLogEntry) {
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
}
