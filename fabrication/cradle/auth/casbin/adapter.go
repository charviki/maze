package casbin

import (
	"strings"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

// LoadFunc 从数据库加载所有策略行的回调函数。
// 每行至少返回 ptype，后续字段按当前策略模型数量提供。
type LoadFunc func() ([][]string, error)

// SaveFunc 持久化策略变更的回调函数。
type SaveFunc func(rules [][]string) error

// DBAdapter 基于回调的 Casbin 适配器，
// 允许调用方注入数据库读写逻辑而不直接依赖具体驱动。
type DBAdapter struct {
	loadFn LoadFunc
	saveFn SaveFunc
}

// NewDBAdapter 创建适配器。
func NewDBAdapter(loadFn LoadFunc, saveFn SaveFunc) *DBAdapter {
	return &DBAdapter{loadFn: loadFn, saveFn: saveFn}
}

// LoadPolicy 从数据库加载所有策略到 model。
func (a *DBAdapter) LoadPolicy(model model.Model) error {
	rows, err := a.loadFn()
	if err != nil {
		return err
	}
	for _, row := range rows {
		if len(row) < 1 {
			continue
		}
		if err := persist.LoadPolicyLine(row[0]+", "+joinFields(row[1:]), model); err != nil {
			return err
		}
	}
	return nil
}

// SavePolicy 保存所有策略到数据库（全量覆盖）。
func (a *DBAdapter) SavePolicy(model model.Model) error {
	var rules [][]string
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			row := []string{ptype}
			row = append(row, rule...)
			rules = append(rules, row)
		}
	}
	if a.saveFn != nil {
		return a.saveFn(rules)
	}
	return nil
}

// AddPolicy 单条添加（DBAdapter 未实现增量写，返回 nil）。
func (a *DBAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemovePolicy 单条删除（DBAdapter 未实现增量写，返回 nil）。
func (a *DBAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy 按过滤条件删除（DBAdapter 未实现增量写，返回 nil）。
func (a *DBAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func joinFields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, ", ")
}
