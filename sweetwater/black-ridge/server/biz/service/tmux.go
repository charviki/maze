package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/creack/pty/v2"
)

// ErrSessionNotFound session 不存在时返回的 sentinel error
var ErrSessionNotFound = errors.New("session not found")

// ErrWorkspaceRootProtected 表示命中了基础工作目录根保护，不能删除整个根目录。
var ErrWorkspaceRootProtected = errors.New("workspace root is protected")

// generateUUID 生成符合 RFC 4122 v4 的 UUID 字符串，用于 Claude CLI 的 --session-id 参数
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// TmuxService Tmux 会话管理接口。通过接口抽象便于测试时 mock
type TmuxService interface {
	ListSessions() ([]model.Session, error)
	CreateSession(name string, command string, workingDir string, configs []model.ConfigItem, restoreStrategy string, templateID string, restoreCommand string) (*model.Session, error)
	KillSession(name string) error
	GetSession(name string) (*model.Session, error)
	CapturePane(name string, lines int) (string, error)
	SendKeys(name string, command string) error
	SendSignal(name string, signal string) error
	AttachSession(name string, rows, cols uint16) (*os.File, error)
	ResizeSession(name string, rows, cols uint16) error
	GetSessionEnv(name string) (map[string]string, error)
	ExecutePipeline(sessionName string, pipeline model.Pipeline) error
	BuildPipeline(workingDir string, command string, configs []model.ConfigItem) model.Pipeline
	SavePipelineState(sessionName string, pipeline model.Pipeline, restoreStrategy string, templateID string, cliSessionID string, restoreCommand string) error
	SaveAllPipelineStates() error
	GetSavedSessions() ([]model.SessionState, error)
	GetSessionState(sessionName string) (*model.SessionState, error)
	RestoreSession(sessionName string) error
	DeleteSessionWorkspace(sessionName string, workspaceRoot string) error
	DeleteSessionState(sessionName string) error
}

// TrustBootstrapper 为外部 CLI 工具注入工作目录信任的接口。
// 具体实现由外部注入，TmuxService 不感知特定 CLI 的配置格式。
type TrustBootstrapper interface {
	TrustDir(workingDir string) error
}

// noopTrustBootstrapper 空实现，不做任何信任操作
type noopTrustBootstrapper struct{}

func (n *noopTrustBootstrapper) TrustDir(_ string) error { return nil }

// tmuxServiceImpl TmuxService 的实现类，未导出，外部只能通过接口使用
type tmuxServiceImpl struct {
	socketPath        string
	defaultShell      string
	stateDir          string
	logger            logutil.Logger
	trustBootstrapper TrustBootstrapper
	saveMu            sync.Mutex
}

// NewTmuxService 根据 TmuxConfig 创建 TmuxService 实例。
// stateDir 为 session 状态文件存储目录，由调用方从配置注入。
// bootstrapper 为可选参数，为 nil 时使用空实现（不执行任何信任操作）。
func NewTmuxService(cfg *config.TmuxConfig, stateDir string, logger logutil.Logger, bootstrapper ...TrustBootstrapper) TmuxService {
	var tb TrustBootstrapper = &noopTrustBootstrapper{}
	if len(bootstrapper) > 0 && bootstrapper[0] != nil {
		tb = bootstrapper[0]
	}
	return &tmuxServiceImpl{
		socketPath:        cfg.SocketPath,
		defaultShell:      cfg.DefaultShell,
		stateDir:          stateDir,
		logger:            logger,
		trustBootstrapper: tb,
	}
}

// tmuxArgs 构建 tmux 命令参数，自动添加 -u（UTF-8）和可选的 -L（socket 路径）
func (s *tmuxServiceImpl) tmuxArgs(args ...string) []string {
	base := []string{"-u"}
	if s.socketPath != "" {
		base = append(base, "-L", s.socketPath)
	}
	return append(base, args...)
}

// runTmux 执行 tmux 命令并返回输出。统一设置 TERM、COLORTERM、LANG 等环境变量
func (s *tmuxServiceImpl) runTmux(args ...string) (string, error) {
	cmd := exec.Command("tmux", s.tmuxArgs(args...)...)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ListSessions 列出所有 tmux 会话。tmux server 未启动或无 session 时返回空列表而非报错
func (s *tmuxServiceImpl) ListSessions() ([]model.Session, error) {
	out, err := s.runTmux("list-sessions", "-F", "#{session_name}::#{session_created}::#{session_windows}")
	if err != nil {
		if strings.Contains(out, "no server running") || strings.Contains(out, "no sessions") || strings.Contains(strings.ToLower(out), "no such file or directory") {
			return []model.Session{}, nil
		}
		return nil, fmt.Errorf("list sessions: %s", strings.TrimSpace(out))
	}

	var sessions []model.Session
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "::")
		if len(parts) < 3 {
			continue
		}
		createdUnix, _ := strconv.ParseInt(parts[1], 10, 64)
		windowCount, _ := strconv.Atoi(parts[2])
		sessions = append(sessions, model.Session{
			ID:          parts[0],
			Name:        parts[0],
			Status:      "running",
			CreatedAt:   time.Unix(createdUnix, 0).Format("2006-01-02 15:04:05"),
			WindowCount: windowCount,
		})
	}
	return sessions, nil
}

// BuildPipeline 根据 session 参数构建三层管线步骤
// system 层: cd + env + file (由系统根据 session 配置自动生成)
// template 层: command (模板定义的启动命令)
// user 层: 无 (用户自定义命令由前端直接传入 pipeline)
func (s *tmuxServiceImpl) BuildPipeline(workingDir string, command string, configs []model.ConfigItem) model.Pipeline {
	var pipeline model.Pipeline
	order := 0

	// system 层: cd 到工作目录
	if workingDir != "" {
		pipeline = append(pipeline, model.PipelineStep{
			ID:    "sys-cd",
			Type:  model.StepCD,
			Phase: model.PhaseSystem,
			Order: order,
			Key:   workingDir,
			Value: "",
		})
		order++
	}

	// system 层: 环境变量
	for _, cfg := range configs {
		if cfg.Type == "env" {
			pipeline = append(pipeline, model.PipelineStep{
				ID:    fmt.Sprintf("sys-env-%s", cfg.Key),
				Type:  model.StepEnv,
				Phase: model.PhaseSystem,
				Order: order,
				Key:   cfg.Key,
				Value: cfg.Value,
			})
			order++
		}
	}

	// system 层: 配置文件
	for _, cfg := range configs {
		if cfg.Type == "file" {
			pipeline = append(pipeline, model.PipelineStep{
				ID:    fmt.Sprintf("sys-file-%s", cfg.Key),
				Type:  model.StepFile,
				Phase: model.PhaseSystem,
				Order: order,
				Key:   cfg.Key,
				Value: cfg.Value,
			})
			order++
		}
	}

	// template 层: 启动命令
	if command != "" {
		pipeline = append(pipeline, model.PipelineStep{
			ID:    "tpl-command",
			Type:  model.StepCommand,
			Phase: model.PhaseTemplate,
			Order: order,
			Key:   "",
			Value: command,
		})
		order++
	}

	return pipeline
}

// ExecutePipeline 按 order 顺序执行管线步骤，每个步骤通过 tmux send-keys 注入。
// env/file 步骤涉及敏感值，执行前关闭 shell 回显防止泄露到终端。
func (s *tmuxServiceImpl) ExecutePipeline(sessionName string, pipeline model.Pipeline) error {
	sorted := pipeline.Sorted()

	// 检测是否包含敏感步骤（env/file），需要临时关闭回显
	hasSensitiveSteps := false
	for _, step := range sorted {
		if step.Type == model.StepEnv || step.Type == model.StepFile {
			hasSensitiveSteps = true
			break
		}
	}

	echoDisabled := false

	// 确保函数退出时恢复回显状态
	defer func() {
		if echoDisabled {
			_ = s.SendKeys(sessionName, "stty echo")
			_ = s.waitForPrompt(sessionName)
		}
	}()

	// 提取工作目录，用于 claude --resume 时匹配正确的 session
	workingDir := ""
	for _, step := range sorted {
		if step.Type == model.StepCD {
			workingDir = step.Key
			if workingDir == "" {
				workingDir = step.Value
			}
			break
		}
	}

	for _, step := range sorted {
		// 在第一个敏感步骤前关闭回显，防止 token 等敏感值泄露到终端
		if !echoDisabled && hasSensitiveSteps && (step.Type == model.StepEnv || step.Type == model.StepFile) {
			_ = s.SendKeys(sessionName, "stty -echo")
			_ = s.waitForPrompt(sessionName)
			echoDisabled = true
		}

		// command 步骤前恢复回显，用户需要看到命令输出
		if echoDisabled && step.Type == model.StepCommand {
			_ = s.SendKeys(sessionName, "stty echo")
			_ = s.waitForPrompt(sessionName)
			echoDisabled = false
		}

		switch step.Type {
		case model.StepCD:
			dir := step.Key
			if dir == "" {
				dir = step.Value
			}
			if err := s.SendKeys(sessionName, fmt.Sprintf("mkdir -p %s && cd %s", dir, dir)); err != nil {
				return fmt.Errorf("pipeline cd step: %w", err)
			}
			if err := s.waitForPrompt(sessionName); err != nil {
				return fmt.Errorf("pipeline cd wait: %w", err)
			}

		case model.StepEnv:
			if _, err := s.runTmux("set-environment", "-t", sessionName, step.Key, step.Value); err != nil {
				return fmt.Errorf("pipeline env tmux set: %w", err)
			}
			escaped := strings.ReplaceAll(step.Value, "'", "'\\''")
			if err := s.SendKeys(sessionName, fmt.Sprintf("export %s='%s'", step.Key, escaped)); err != nil {
				return fmt.Errorf("pipeline env export: %w", err)
			}
			if err := s.waitForPrompt(sessionName); err != nil {
				return fmt.Errorf("pipeline env wait: %w", err)
			}

		case model.StepFile:
			dir := filepath.Dir(step.Key)
			if dir != "." && dir != "" {
				expanded := s.expandPath(dir)
				if err := s.SendKeys(sessionName, fmt.Sprintf("mkdir -p %s", expanded)); err != nil {
					return fmt.Errorf("pipeline file mkdir: %w", err)
				}
				if err := s.waitForPrompt(sessionName); err != nil {
					return fmt.Errorf("pipeline file mkdir wait: %w", err)
				}
			}
			expanded := s.expandPath(step.Key)
			heredoc := fmt.Sprintf("cat > %s << 'SESSIONCONFIGEOF'\n%s\nSESSIONCONFIGEOF", expanded, step.Value)
			if err := s.SendKeys(sessionName, heredoc); err != nil {
				return fmt.Errorf("pipeline file write: %w", err)
			}
			if err := s.waitForPrompt(sessionName); err != nil {
				return fmt.Errorf("pipeline file write wait: %w", err)
			}

		case model.StepCommand:
			if err := s.SendKeys(sessionName, step.Value); err != nil {
				return fmt.Errorf("pipeline command send: %w", err)
			}
		}
	}

	return nil
}

// CreateSession 创建 tmux 会话的完整流程，基于管线执行:
// 1. new-session -d 创建后台会话
// 2. waitForPrompt 等待 shell 就绪
// 3. BuildPipeline 构建三层管线
// 4. 为 command 步骤中的 {session_id} 占位符填充预生成的 UUID
// 5. ExecutePipeline 按顺序执行所有步骤
// 6. SavePipelineState 保存管线状态到本地文件
func (s *tmuxServiceImpl) CreateSession(name string, command string, workingDir string, configs []model.ConfigItem, restoreStrategy string, templateID string, restoreCommand string) (*model.Session, error) {
	args := []string{"new-session", "-d", "-s", name}
	if _, err := s.runTmux(args...); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// 管线执行失败时需要回滚已创建的 session，避免僵尸 session 残留
	if err := s.waitForPrompt(name); err != nil {
		_ = s.KillSession(name)
		return nil, fmt.Errorf("wait for shell init: %w", err)
	}

	pipeline := s.BuildPipeline(workingDir, command, configs)

	// 为 command 步骤中的 {session_id} 占位符生成并填充 UUID，
	// 使 Claude CLI 启动时通过 --session-id 携带已知 ID，恢复时可用 --resume 引用同一 ID。
	cliSessionID := generateUUID()
	for i := range pipeline {
		if pipeline[i].Type == model.StepCommand && strings.Contains(pipeline[i].Value, "{session_id}") {
			pipeline[i].Value = strings.ReplaceAll(pipeline[i].Value, "{session_id}", cliSessionID)
		}
	}

	if err := s.ExecutePipeline(name, pipeline); err != nil {
		// 管线执行失败，回滚：终止 session 并清理状态文件
		_ = s.KillSession(name)
		_ = s.DeleteSessionState(name)
		return nil, fmt.Errorf("execute pipeline: %w", err)
	}

	// 保存管线状态到本地文件，失败时记录警告但不阻塞创建流程
	if restoreStrategy == "" {
		restoreStrategy = "auto"
	}
	if err := s.SavePipelineState(name, pipeline, restoreStrategy, templateID, cliSessionID, restoreCommand); err != nil {
		s.logger.Errorf("[tmux] save pipeline state for session %s failed: %v", name, err)
	}

	session, err := s.GetSession(name)
	if err != nil {
		return nil, err
	}
	return session, nil
}

var defaultPromptPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[#$>]$`),
	regexp.MustCompile(`[❯➜λ%]$`),
}

const (
	promptPollInterval    = 50 * time.Millisecond
	promptTimeout         = 5 * time.Second
	promptMaxRetries      = int(promptTimeout / promptPollInterval)
	terminalSnapshotLines = 50
)

// waitForPrompt 轮询终端内容，直到最后一行出现 shell 提示符（表示命令已执行完毕）
// 超时后记录警告并返回错误，调用方可决定是否继续
func (s *tmuxServiceImpl) waitForPrompt(name string) error {

	for i := 0; i < promptMaxRetries; i++ {
		time.Sleep(promptPollInterval)
		out, err := s.CapturePane(name, 3)
		if err != nil {
			continue
		}
		lines := strings.Split(out, "\n")
		var lastLine string
		for i := len(lines) - 1; i >= 0; i-- {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed != "" {
				lastLine = trimmed
				break
			}
		}
		if lastLine == "" {
			continue
		}
		for _, re := range defaultPromptPatterns {
			if re.MatchString(lastLine) {
				return nil
			}
		}
	}
	s.logger.Warnf("[tmux] timeout waiting for prompt in session %s", name)
	return fmt.Errorf("timeout waiting for prompt in session %s", name)
}

// expandPath 将 ~/ 开头的路径展开为用户主目录的绝对路径
func (s *tmuxServiceImpl) expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			s.logger.Warnf("[tmux] get user home dir failed: %v, using /root as fallback", err)
			home = "/root"
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

// KillSession 终止指定 tmux 会话
func (s *tmuxServiceImpl) KillSession(name string) error {
	// 先做 has-session 检查，统一将“不存在”映射为 ErrSessionNotFound，
	// 避免不同 tmux 版本/输出格式下 kill-session 的 stderr 文案不稳定。
	if _, err := s.runTmux("has-session", "-t", name); err != nil {
		return fmt.Errorf("kill session %s: %w", name, ErrSessionNotFound)
	}

	out, err := s.runTmux("kill-session", "-t", name)
	if err != nil {
		// tmux 对不存在的 session 返回 "can't find session" 错误，统一 wrap 为 ErrSessionNotFound
		if strings.Contains(out, "can't find session") || strings.Contains(out, "session not found") {
			return fmt.Errorf("kill session %s: %w", name, ErrSessionNotFound)
		}
		return fmt.Errorf("kill session: %w", err)
	}
	return nil
}

// GetSession 通过名称查找会话，不存在时返回 ErrSessionNotFound
func (s *tmuxServiceImpl) GetSession(name string) (*model.Session, error) {
	sessions, err := s.ListSessions()
	if err != nil {
		return nil, err
	}
	for _, sess := range sessions {
		if sess.Name == name {
			return &sess, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, name)
}

// CapturePane 捕获指定会话的终端输出内容
func (s *tmuxServiceImpl) CapturePane(name string, lines int) (string, error) {
	out, err := s.runTmux("capture-pane", "-t", name, "-p", "-S", fmt.Sprintf("-%d", lines))
	if err != nil {
		return "", fmt.Errorf("capture pane: %w", err)
	}
	return out, nil
}

// SendKeys 向会话发送按键序列，先发送文本再发送回车
func (s *tmuxServiceImpl) SendKeys(name string, command string) error {
	_, err := s.runTmux("send-keys", "-t", name, "-l", command)
	if err != nil {
		return fmt.Errorf("send keys: %w", err)
	}
	_, err = s.runTmux("send-keys", "-t", name, "Enter")
	if err != nil {
		return fmt.Errorf("send enter: %w", err)
	}
	return nil
}

// SendSignal 发送控制信号。sigint 对应 Ctrl+C，up/down/enter 对应方向键和回车
func (s *tmuxServiceImpl) SendSignal(name string, signal string) error {
	switch strings.ToLower(signal) {
	case "sigint":
		_, err := s.runTmux("send-keys", "-t", name, "C-c")
		return err
	case "up":
		_, err := s.runTmux("send-keys", "-t", name, "Up")
		return err
	case "down":
		_, err := s.runTmux("send-keys", "-t", name, "Down")
		return err
	case "enter":
		_, err := s.runTmux("send-keys", "-t", name, "Enter")
		return err
	default:
		return fmt.Errorf("unsupported signal: %s", signal)
	}
}

// AttachSession 通过 PTY 附加到 tmux 会话，返回 PTY 文件描述符用于双向数据传输
func (s *tmuxServiceImpl) AttachSession(name string, rows, cols uint16) (*os.File, error) {
	// 预检查 session 是否存在，避免 PTY 启动后才发现 session 不存在导致 file already closed 错误
	if _, err := s.runTmux("has-session", "-t", name); err != nil {
		return nil, fmt.Errorf("session %q does not exist: %w", name, ErrSessionNotFound)
	}

	tmuxArgs := s.tmuxArgs("attach-session", "-t", name)
	cmd := exec.Command("tmux", tmuxArgs...)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
	)
	winsize := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}
	ptmx, err := pty.StartWithSize(cmd, winsize)
	if err != nil {
		return nil, fmt.Errorf("attach session via pty: %w", err)
	}
	return ptmx, nil
}

// ResizeSession 将前端实际终端尺寸同步到 tmux window。
// 仅调整外层 PTY 不会改变 tmux 自己维护的网格大小，因此 shell 仍可能按旧的 80x24 排版。
func (s *tmuxServiceImpl) ResizeSession(name string, rows, cols uint16) error {
	if rows == 0 || cols == 0 {
		return nil
	}

	target := fmt.Sprintf("%s:0", name)
	out, err := s.runTmux("resize-window", "-t", target, "-x", fmt.Sprintf("%d", cols), "-y", fmt.Sprintf("%d", rows))
	if err != nil {
		return fmt.Errorf("resize tmux window %s to %dx%d: %s", target, cols, rows, strings.TrimSpace(out))
	}

	return nil
}

// GetSessionEnv 获取会话的环境变量列表，解析 tmux show-environment 输出
func (s *tmuxServiceImpl) GetSessionEnv(name string) (map[string]string, error) {
	out, err := s.runTmux("show-environment", "-t", name)
	if err != nil {
		return nil, fmt.Errorf("show environment: %w", err)
	}
	env := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		if idx := strings.Index(line, "="); idx >= 0 {
			env[line[:idx]] = line[idx+1:]
		}
	}
	return env, nil
}

// SavePipelineState 保存单个 session 的管线状态到本地 JSON 文件
func (s *tmuxServiceImpl) SavePipelineState(sessionName string, pipeline model.Pipeline, restoreStrategy string, templateID string, cliSessionID string, restoreCommand string) error {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	return s.savePipelineStateLocked(sessionName, pipeline, restoreStrategy, templateID, cliSessionID, restoreCommand)
}

// savePipelineStateLocked 在持有 saveMu 时写入单个 session 状态。
// SaveAllPipelineStates 需要批量保存多个 session，若继续调用带锁的 SavePipelineState 会自锁卡死。
func (s *tmuxServiceImpl) savePipelineStateLocked(sessionName string, pipeline model.Pipeline, restoreStrategy string, templateID string, cliSessionID string, restoreCommand string) error {

	if err := os.MkdirAll(s.stateDir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// 捕获环境变量快照（忽略错误，session 可能已退出）
	envSnapshot, _ := s.GetSessionEnv(sessionName)
	if envSnapshot == nil {
		envSnapshot = make(map[string]string)
	}

	// 捕获终端快照（忽略错误）
	terminalSnapshot, _ := s.CapturePane(sessionName, terminalSnapshotLines)

	// 从 pipeline 的 cd 步骤中提取工作目录，恢复时需要正确 cd 到此目录
	workingDir := ""
	for _, step := range pipeline {
		if step.Type == model.StepCD {
			workingDir = step.Key
			break
		}
	}

	state := model.SessionState{
		SessionName:      sessionName,
		Pipeline:         pipeline,
		RestoreStrategy:  restoreStrategy,
		RestoreCommand:   restoreCommand,
		WorkingDir:       workingDir,
		TemplateID:       templateID,
		CLISessionID:     cliSessionID,
		EnvSnapshot:      envSnapshot,
		TerminalSnapshot: terminalSnapshot,
		SavedAt:          time.Now().Format(time.RFC3339),
	}

	data, err := state.ToJSON()
	if err != nil {
		return fmt.Errorf("serialize state: %w", err)
	}

	filePath := filepath.Join(s.stateDir, sessionName+".json")
	// 使用原子写入防止写入过程中崩溃导致状态文件损坏
	if err := configutil.AtomicWriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// SaveAllPipelineStates 遍历所有活跃 session，批量保存管线状态
// 从已保存的状态文件中读取 pipeline 和 restore_strategy，仅更新快照部分
func (s *tmuxServiceImpl) SaveAllPipelineStates() error {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	sessions, err := s.ListSessions()
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	if len(sessions) == 0 {
		return nil
	}

	for _, sess := range sessions {
		stateFile := filepath.Join(s.stateDir, sess.Name+".json")
		var pipeline model.Pipeline
		restoreStrategy := "auto"
		var templateID string
		var cliSessionID string
		var restoreCommand string

		if data, err := os.ReadFile(stateFile); err == nil {
			var existing model.SessionState
			if err := existing.FromJSON(data); err == nil {
				pipeline = existing.Pipeline
				restoreStrategy = existing.RestoreStrategy
				templateID = existing.TemplateID
				cliSessionID = existing.CLISessionID
				restoreCommand = existing.RestoreCommand
			}
		}

		if len(pipeline) == 0 {
			continue
		}

		// 这里已经持有 saveMu，必须调用无锁版本，否则会递归加锁导致删除/保存接口一直阻塞。
		if err := s.savePipelineStateLocked(sess.Name, pipeline, restoreStrategy, templateID, cliSessionID, restoreCommand); err != nil {
			s.logger.Errorf("[tmux] save pipeline state for session %s failed: %v", sess.Name, err)
		}
	}

	return nil
}

// GetSavedSessions 返回 /home/agent/.session-state/ 下所有已保存 session 的列表
func (s *tmuxServiceImpl) GetSavedSessions() ([]model.SessionState, error) {
	entries, err := os.ReadDir(s.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.SessionState{}, nil
		}
		return nil, fmt.Errorf("read state dir: %w", err)
	}

	// 获取当前活跃的 session 列表，用于标记状态
	activeSessions, err := s.ListSessions()
	if err != nil {
		s.logger.Errorf("[tmux] list sessions for saved states check failed: %v", err)
	}
	activeSet := make(map[string]bool)
	for _, sess := range activeSessions {
		activeSet[sess.Name] = true
	}

	var states []model.SessionState
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.stateDir, entry.Name()))
		if err != nil {
			continue
		}
		var state model.SessionState
		if err := state.FromJSON(data); err != nil {
			continue
		}
		// 如果 tmux session 仍然活跃，标记为 running
		if activeSet[state.SessionName] {
			state.RestoreStrategy = "running"
		}
		states = append(states, state)
	}

	return states, nil
}

// RestoreSession 读取指定 session 的状态文件并重放管线。
// 使用 ExecutePipeline 统一执行 env/file/cd 步骤（literal mode + waitForPrompt），
// 单独处理 command 步骤以支持 {session_id} 替换。
func (s *tmuxServiceImpl) RestoreSession(sessionName string) error {
	stateFile := filepath.Join(s.stateDir, sessionName+".json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("read state file: %w", err)
	}

	var state model.SessionState
	if err := state.FromJSON(data); err != nil {
		return fmt.Errorf("parse state file: %w", err)
	}

	s.logger.Infof("[restore] restoring session: %s (strategy=%s, template=%s, cli_sid=%s)",
		sessionName, state.RestoreStrategy, state.TemplateID, state.CLISessionID)

	if _, err := s.GetSession(sessionName); err == nil {
		if err := s.KillSession(sessionName); err != nil {
			return fmt.Errorf("kill existing session before restore: %w", err)
		}
	}

	args := []string{"new-session", "-d", "-s", sessionName}
	if _, err := s.runTmux(args...); err != nil {
		return fmt.Errorf("create session for restore: %w", err)
	}

	if err := s.waitForPrompt(sessionName); err != nil {
		return fmt.Errorf("wait for shell init: %w", err)
	}

	// 优先使用模板定义的 RestoreCommand（含 --dangerously-skip-permissions 等恢复专用标志），
	// 降级到 pipeline 中的 command 步骤（向后兼容旧状态文件）
	var nonCommandSteps model.Pipeline
	var pipelineCmd string
	for _, step := range state.Pipeline {
		if step.Type == model.StepCommand {
			pipelineCmd = step.Value
		} else {
			nonCommandSteps = append(nonCommandSteps, step)
		}
	}

	if len(nonCommandSteps) > 0 {
		if err := s.ExecutePipeline(sessionName, nonCommandSteps); err != nil {
			return fmt.Errorf("restore pipeline steps: %w", err)
		}
	}

	// 确定最终恢复命令：优先用 RestoreCommand，否则降级到 pipeline 中的 command
	restoreCmd := state.RestoreCommand
	if restoreCmd == "" {
		restoreCmd = pipelineCmd
	}

	// 替换 {session_id} 占位符为实际的 CLI session ID
	if state.CLISessionID != "" {
		restoreCmd = strings.ReplaceAll(restoreCmd, "{session_id}", state.CLISessionID)
		if strings.Contains(restoreCmd, "--resume") {
			s.logger.Infof("[restore] using restore command with session_id: %s", state.CLISessionID)
		}
	}

	if restoreCmd != "" {
		if err := s.SendKeys(sessionName, restoreCmd); err != nil {
			return fmt.Errorf("send restore command: %w", err)
		}
	}

	return nil
}

// DeleteSessionWorkspace 删除 session 实际使用的工作目录。
// 目录来源以状态文件快照为准，避免仅靠 session 名猜目录导致误删。
func (s *tmuxServiceImpl) DeleteSessionWorkspace(sessionName string, workspaceRoot string) error {
	state, err := s.loadSessionState(sessionName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load session state for workspace cleanup: %w", err)
	}

	workingDir := filepath.Clean(state.WorkingDir)
	if workingDir == "" || workingDir == "." {
		return nil
	}

	protectedRoot := filepath.Clean(workspaceRoot)
	if protectedRoot != "" && workingDir == protectedRoot {
		return fmt.Errorf("workspace %s: %w", workingDir, ErrWorkspaceRootProtected)
	}

	if err := os.RemoveAll(workingDir); err != nil {
		return fmt.Errorf("delete session workspace %s: %w", workingDir, err)
	}
	return nil
}

// DeleteSessionState 删除指定 session 的状态文件
func (s *tmuxServiceImpl) DeleteSessionState(sessionName string) error {
	stateFile := filepath.Join(s.stateDir, sessionName+".json")
	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete state file: %w", err)
	}
	return nil
}

func (s *tmuxServiceImpl) loadSessionState(sessionName string) (*model.SessionState, error) {
	stateFile := filepath.Join(s.stateDir, sessionName+".json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state model.SessionState
	if err := state.FromJSON(data); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}
	return &state, nil
}

// GetSessionState 返回单个 session 的已保存状态快照，供配置查看接口定位工作目录与模板信息。
func (s *tmuxServiceImpl) GetSessionState(sessionName string) (*model.SessionState, error) {
	state, err := s.loadSessionState(sessionName)
	if err != nil {
		return nil, fmt.Errorf("get session state %s: %w", sessionName, err)
	}
	return state, nil
}
