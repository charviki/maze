package casbin

import (
	_ "embed"

	gocasbin "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

//go:embed model.conf
var modelConf string

// ResourceAction 将 gRPC method 映射到 Casbin 资源 + 操作。
type ResourceAction struct {
	Resource string
	Action   string
}

// Enforcer 封装 Casbin enforcer，首批仅使用三元策略模型。
type Enforcer struct {
	enforcer *gocasbin.Enforcer
}

// NewEnforcer 创建 Casbin enforcer。
func NewEnforcer(adapter persist.Adapter) (*Enforcer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, err
	}

	var e *gocasbin.Enforcer
	if adapter != nil {
		e, err = gocasbin.NewEnforcer(m, adapter)
	} else {
		e, err = gocasbin.NewEnforcer(m)
	}
	if err != nil {
		return nil, err
	}

	return &Enforcer{enforcer: e}, nil
}

// Enforce 检查指定主体是否拥有资源动作权限。
func (e *Enforcer) Enforce(subjectKey, resource, action string) (bool, error) {
	return e.enforcer.Enforce(subjectKey, resource, action)
}

// AddPolicy 添加策略。
func (e *Enforcer) AddPolicy(params ...interface{}) (bool, error) {
	return e.enforcer.AddPolicy(params...)
}

// RemovePolicy 移除策略。
func (e *Enforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return e.enforcer.RemovePolicy(params...)
}

// LoadPolicy 从适配器重新加载策略到内存。
func (e *Enforcer) LoadPolicy() error {
	return e.enforcer.LoadPolicy()
}

// GetPermissionsForUser 返回主体已加载到内存中的策略。
func (e *Enforcer) GetPermissionsForUser(subjectKey string) [][]string {
	perms, _ := e.enforcer.GetPermissionsForUser(subjectKey)
	return perms
}

// Raw 底层 Casbin enforcer，供高级用法使用。
func (e *Enforcer) Raw() *gocasbin.Enforcer {
	return e.enforcer
}
