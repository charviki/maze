package service

// HealthService 提供健康检查。
type HealthService struct{}

// NewHealthService 创建 HealthService。
func NewHealthService() *HealthService {
	return &HealthService{}
}

// Status 返回服务健康状态。
func (s *HealthService) Status() string {
	return "ok"
}
