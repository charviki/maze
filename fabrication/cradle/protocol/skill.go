package protocol

import "time"

// Skill 技能定义
type Skill struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Config      map[string]string `json:"config,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
