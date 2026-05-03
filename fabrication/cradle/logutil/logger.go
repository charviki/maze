package logutil

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Logger 统一日志接口，覆盖当前项目实际使用的日志方法 + 结构化字段方法。
// 定义在 cradle 共享库中，确保 Manager 和 Agent 使用同一套日志规范。
// WithNode/WithSession/WithAction 返回新的 Logger 实例，不修改原始实例。
type Logger interface {
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	WithNode(name string) Logger
	WithSession(id string) Logger
	WithAction(action string) Logger
	WithFields(fields ...string) Logger
}

// SlogLogger 基于 Go 标准库 log/slog 的 Logger 实现。
// attrs 累积 With 链附加的结构化属性，SetLevel/SetOutput 时通过 rebuildInner 重建 inner，
// 避免丢失上下文字段。
type SlogLogger struct {
	inner *slog.Logger
	mu    sync.Mutex
	w     io.Writer
	level slog.Level
	attrs []any
}

// rebuildInner 根据 w、level、attrs 重建 inner，确保 With 链上下文不丢失。
func (l *SlogLogger) rebuildInner() {
	handler := slog.NewJSONHandler(l.w, &slog.HandlerOptions{Level: l.level})
	l.inner = slog.New(handler)
	if len(l.attrs) > 0 {
		l.inner = l.inner.With(l.attrs...)
	}
}

// New 创建基于 slog 的结构化日志实例，输出 JSON 格式到 stdout。
func New(component string) *SlogLogger {
	w := os.Stdout
	level := slog.LevelInfo
	l := &SlogLogger{
		w:     w,
		level: level,
		attrs: []any{"component", component},
	}
	l.rebuildInner()
	return l
}

// NewNop 创建一个丢弃所有输出的 Logger，用于测试。
// Fatalf 仍会调用 os.Exit(1)。
func NewNop() *SlogLogger {
	w := io.Discard
	level := slog.LevelInfo
	l := &SlogLogger{
		w:     w,
		level: level,
		attrs: []any{},
	}
	l.rebuildInner()
	return l
}

// clone 创建当前实例的浅拷贝，共享 attrs 切片的副本
func (l *SlogLogger) clone() *SlogLogger {
	attrs := make([]any, len(l.attrs))
	copy(attrs, l.attrs)
	return &SlogLogger{
		w:     l.w,
		level: l.level,
		attrs: attrs,
	}
}

// WithNode 附加 node_name 字段，用于节点注册/心跳/代理等关键路径
func (l *SlogLogger) WithNode(name string) Logger {
	c := l.clone()
	c.attrs = append(c.attrs, "node_name", name)
	c.rebuildInner()
	return c
}

// WithSession 附加 session_id 字段，用于 Session 创建/恢复/代理等关键路径
func (l *SlogLogger) WithSession(id string) Logger {
	c := l.clone()
	c.attrs = append(c.attrs, "session_id", id)
	c.rebuildInner()
	return c
}

// WithAction 附加 action 字段，用于代理转发/审计等关键路径
func (l *SlogLogger) WithAction(action string) Logger {
	c := l.clone()
	c.attrs = append(c.attrs, "action", action)
	c.rebuildInner()
	return c
}

// WithFields 一次性附加多个结构化字段
func (l *SlogLogger) WithFields(fields ...string) Logger {
	if len(fields)%2 != 0 {
		return l
	}
	c := l.clone()
	for i := 0; i < len(fields); i += 2 {
		c.attrs = append(c.attrs, fields[i], fields[i+1])
	}
	c.rebuildInner()
	return c
}

// --- logutil.Logger 接口实现 ---

func (l *SlogLogger) Infof(format string, v ...interface{}) {
	l.inner.Info(fmt.Sprintf(format, v...))
}

func (l *SlogLogger) Warnf(format string, v ...interface{}) {
	l.inner.Warn(fmt.Sprintf(format, v...))
}

func (l *SlogLogger) Errorf(format string, v ...interface{}) {
	l.inner.Error(fmt.Sprintf(format, v...))
}

func (l *SlogLogger) Fatalf(format string, v ...interface{}) {
	l.inner.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

// --- 扩展格式化日志方法 ---

func (l *SlogLogger) Tracef(format string, v ...interface{}) {
	l.inner.Debug(fmt.Sprintf(format, v...))
}

func (l *SlogLogger) Debugf(format string, v ...interface{}) {
	l.inner.Debug(fmt.Sprintf(format, v...))
}

func (l *SlogLogger) Noticef(format string, v ...interface{}) {
	l.inner.Info(fmt.Sprintf(format, v...))
}

// --- 扩展非格式化日志方法 ---

func (l *SlogLogger) Trace(v ...interface{})  { l.Tracef("%v", v...) }
func (l *SlogLogger) Debug(v ...interface{})  { l.Debugf("%v", v...) }
func (l *SlogLogger) Info(v ...interface{})   { l.Infof("%v", v...) }
func (l *SlogLogger) Notice(v ...interface{}) { l.Noticef("%v", v...) }
func (l *SlogLogger) Warn(v ...interface{})   { l.Warnf("%v", v...) }
func (l *SlogLogger) Error(v ...interface{})  { l.Errorf("%v", v...) }
func (l *SlogLogger) Fatal(v ...interface{})  { l.Fatalf("%v", v...) }

// --- 扩展带 context 的日志方法 ---
// context 参数暂不使用，直接委托给对应 Xxxf 方法

func (l *SlogLogger) CtxTracef(_ context.Context, format string, v ...interface{}) {
	l.Tracef(format, v...)
}
func (l *SlogLogger) CtxDebugf(_ context.Context, format string, v ...interface{}) {
	l.Debugf(format, v...)
}
func (l *SlogLogger) CtxInfof(_ context.Context, format string, v ...interface{}) {
	l.Infof(format, v...)
}
func (l *SlogLogger) CtxNoticef(_ context.Context, format string, v ...interface{}) {
	l.Noticef(format, v...)
}
func (l *SlogLogger) CtxWarnf(_ context.Context, format string, v ...interface{}) {
	l.Warnf(format, v...)
}
func (l *SlogLogger) CtxErrorf(_ context.Context, format string, v ...interface{}) {
	l.Errorf(format, v...)
}
func (l *SlogLogger) CtxFatalf(_ context.Context, format string, v ...interface{}) {
	l.Fatalf(format, v...)
}

// SetLevel 调整底层 slog 等级，便于测试或局部降噪。
func (l *SlogLogger) SetLevel(level slog.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	l.rebuildInner()
}

// SetOutput 替换底层 handler 的输出目标
func (l *SlogLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.w = w
	l.rebuildInner()
}

// --- 辅助类型 ---
