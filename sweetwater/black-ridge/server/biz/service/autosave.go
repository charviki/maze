package service

import (
	"time"

	"github.com/charviki/maze-cradle/logutil"
)

// AutoSaveService 定时保存所有活跃 session 的管线状态
type AutoSaveService struct {
	tmuxService TmuxService
	interval    time.Duration
	logger      logutil.Logger
}

// NewAutoSaveService 创建自动保存服务，interval 单位为秒
func NewAutoSaveService(tmuxService TmuxService, intervalSeconds int, logger logutil.Logger) *AutoSaveService {
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}
	return &AutoSaveService{
		tmuxService: tmuxService,
		interval:    time.Duration(intervalSeconds) * time.Second,
		logger:      logger,
	}
}

// Start 启动定时保存循环，通过 stopCh 实现优雅停止
func (s *AutoSaveService) Start(stopCh <-chan struct{}) {
	s.logger.Infof("[autosave] started, interval=%s", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			s.logger.Infof("[autosave] stopped")
			return
		case <-ticker.C:
			if err := s.tmuxService.SaveAllPipelineStates(); err != nil {
				s.logger.Errorf("[autosave] error: %v", err)
			}
		}
	}
}
