package casbin

// AdminSubjectKey 是首批 bootstrap 的超级管理员主体。
const AdminSubjectKey = "user:admin"

// SeedPolicies 仅写入首批最小系统策略。
func SeedPolicies(e *Enforcer) error {
	_, err := e.AddPolicy(AdminSubjectKey, "*", "*")
	return err
}
