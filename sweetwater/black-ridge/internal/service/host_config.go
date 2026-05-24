package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

// HostConfigService 从 director-core 拉取 Host 配置并写入文件。
type HostConfigService struct {
	agentClient pb.AgentServiceClient
	hostName    string
	authToken   string
	logger      logutil.Logger
}

// NewHostConfigService 创建 HostConfigService。
func NewHostConfigService(agentClient pb.AgentServiceClient, hostName string, authToken string, logger logutil.Logger) *HostConfigService {
	return &HostConfigService{
		agentClient: agentClient,
		hostName:    hostName,
		authToken:   authToken,
		logger:      logger,
	}
}

func (s *HostConfigService) withAuth(ctx context.Context) context.Context {
	return withBearerAuth(ctx, s.authToken)
}

// FetchAndApply 拉取配置并写入文件。
func (s *HostConfigService) FetchAndApply(ctx context.Context) error {
	resp, err := s.agentClient.GetHostConfig(s.withAuth(ctx), &pb.GetHostConfigRequest{Name: s.hostName})
	if err != nil {
		return fmt.Errorf("fetch host config: %w", err)
	}

	config := &protocol.HostConfig{}
	for _, skill := range resp.GetSkills() {
		config.Skills = append(config.Skills, protocol.SkillConfig{
			Name:        skill.GetName(),
			Description: skill.GetDescription(),
			Config:      skill.GetConfig(),
		})
	}
	for _, rule := range resp.GetRules() {
		config.Rules = append(config.Rules, protocol.RuleConfig{
			Name:    rule.GetName(),
			Content: rule.GetContent(),
		})
	}
	for _, key := range resp.GetGitKeys() {
		config.GitKeys = append(config.GitKeys, protocol.GitKeyItem{
			Name:           key.GetName(),
			TokenType:      key.GetTokenType(),
			Host:           key.GetHost(),
			DecryptedToken: key.GetDecryptedToken(),
		})
	}

	home := resolveHostConfigHome()
	s.logger.Infof("[host-config] fetched config: %d skills, %d rules, %d git_keys",
		len(config.Skills), len(config.Rules), len(config.GitKeys))

	var writeErrors int
	writeErrors += s.writeSkills(home, config.Skills)
	writeErrors += s.writeRules(home, config.Rules)
	writeErrors += s.writeGitKeys(home, config.GitKeys)

	if writeErrors > 0 {
		s.logger.Errorf("[host-config] config injection completed with %d write errors — host may be missing skills/rules/keys", writeErrors)
	} else {
		s.logger.Infof("[host-config] config injection completed successfully")
	}

	return nil
}

func resolveHostConfigHome() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	return "/root"
}

func (s *HostConfigService) writeSkills(home string, skills []protocol.SkillConfig) int {
	if len(skills) == 0 {
		return 0
	}

	var errs int
	claudeSkillsDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(claudeSkillsDir, 0750); err != nil {
		s.logger.Warnf("[host-config] mkdir %s: %v", claudeSkillsDir, err)
		return len(skills) * 2
	}

	codexSkillsDir := filepath.Join(home, ".codex", "skills")
	if err := os.MkdirAll(codexSkillsDir, 0750); err != nil {
		s.logger.Warnf("[host-config] mkdir %s: %v", codexSkillsDir, err)
		return len(skills) * 2
	}

	for _, skill := range skills {
		content := buildSkillContent(skill)

		claudePath := filepath.Join(claudeSkillsDir, skill.Name+".md")
		if err := os.WriteFile(claudePath, []byte(content), 0600); err != nil {
			s.logger.Warnf("[host-config] write skill %s for claude: %v", skill.Name, err)
			errs++
		}

		codexPath := filepath.Join(codexSkillsDir, skill.Name+".md")
		if err := os.WriteFile(codexPath, []byte(content), 0600); err != nil {
			s.logger.Warnf("[host-config] write skill %s for codex: %v", skill.Name, err)
			errs++
		}
	}
	return errs
}

func buildSkillContent(skill protocol.SkillConfig) string {
	var sb strings.Builder
	if skill.Description != "" {
		sb.WriteString("# ")
		sb.WriteString(skill.Name)
		sb.WriteString("\n\n")
		sb.WriteString(skill.Description)
		sb.WriteString("\n\n")
	}
	for k, v := range skill.Config {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (s *HostConfigService) writeRules(home string, rules []protocol.RuleConfig) int {
	if len(rules) == 0 {
		return 0
	}

	var errs int
	var sb strings.Builder
	for _, rule := range rules {
		sb.WriteString("## Rule: ")
		sb.WriteString(rule.Name)
		sb.WriteString("\n\n")
		sb.WriteString(rule.Content)
		sb.WriteString("\n\n")
	}
	rulesContent := sb.String()

	claudeMDPath := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := overwriteManagedSection(claudeMDPath, rulesContent); err != nil {
		s.logger.Warnf("[host-config] write rules to %s: %v", claudeMDPath, err)
		errs++
	}

	agentsMDPath := filepath.Join(home, ".codex", "AGENTS.md")
	if err := overwriteManagedSection(agentsMDPath, rulesContent); err != nil {
		s.logger.Warnf("[host-config] write rules to %s: %v", agentsMDPath, err)
		errs++
	}
	return errs
}

const managedMarker = "<!-- maze-managed:start -->"

func overwriteManagedSection(path, content string) error {
	cleanPath := filepath.Clean(path)
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var newContent string
	if err != nil && os.IsNotExist(err) {
		newContent = managedMarker + "\n" + content
	} else {
		existing := string(data)
		idx := strings.Index(existing, managedMarker)
		if idx >= 0 {
			newContent = existing[:idx] + managedMarker + "\n" + content
		} else {
			newContent = existing + "\n" + managedMarker + "\n" + content
		}
	}

	tmp, err := os.CreateTemp(dir, ".maze-managed-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(newContent); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0600); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, cleanPath)
}

func (s *HostConfigService) writeGitKeys(home string, keys []protocol.GitKeyItem) int {
	if len(keys) == 0 {
		return 0
	}

	var errs int
	sshDir := filepath.Join(home, ".ssh")
	sshKeyIndex := 0

	for _, key := range keys {
		switch key.TokenType {
		case protocol.GitKeyTypeSSHKey:
			if err := s.writeSSHKey(sshDir, key, sshKeyIndex); err != nil {
				s.logger.Warnf("[host-config] write ssh key %s: %v", key.Name, err)
				errs++
			}
			sshKeyIndex++

		case protocol.GitKeyTypePAT:
			if err := s.writePAT(home, key); err != nil {
				s.logger.Warnf("[host-config] write pat %s: %v", key.Name, err)
				errs++
			}
		}
	}
	return errs
}

func (s *HostConfigService) writeSSHKey(sshDir string, key protocol.GitKeyItem, index int) error {
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}

	keyFile := "id_ed25519"
	if index > 0 {
		keyFile = fmt.Sprintf("id_ed25519_%d", index)
	}
	keyPath := filepath.Join(sshDir, keyFile)
	if err := os.WriteFile(keyPath, []byte(key.DecryptedToken), 0600); err != nil {
		return err
	}

	if key.Host != "" {
		configEntry := fmt.Sprintf("Host %s\n  IdentityFile ~/.ssh/%s\n  StrictHostKeyChecking no\n", key.Host, keyFile)
		configPath := filepath.Join(sshDir, "config")
		if err := overwriteManagedSection(configPath, configEntry); err != nil {
			return err
		}
	}
	return nil
}

func (s *HostConfigService) writePAT(home string, key protocol.GitKeyItem) error {
	credLine := "https://x-access-token:" + key.DecryptedToken
	if key.Host != "" {
		credLine += "@" + key.Host
	}
	credLine += "\n"

	credPath := filepath.Join(home, ".git-credentials")
	if err := overwriteManagedSection(credPath, credLine); err != nil {
		return err
	}

	gitconfig := filepath.Join(home, ".gitconfig")
	content := "[credential]\n\thelper = store\n"
	if err := os.WriteFile(filepath.Clean(gitconfig), []byte(content), 0600); err != nil {
		return err
	}
	return nil
}
