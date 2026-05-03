package kit

import "fmt"

// LeasePool 为固定资源集合提供阻塞式独占租约。
// 每个 profile 都有一个预填充的 channel，Acquire/Release 天然满足“一次只借出一个”。
type LeasePool struct {
	channels map[string]chan string
	all      map[string][]string
}

// NewLeasePool 基于 profile -> host 名称列表创建固定大小的租约池。
// 这里预先把所有 host 放入缓冲 channel，是为了用最简单的阻塞语义表达“资源已占满请等待”。
func NewLeasePool(profiles map[string][]string) *LeasePool {
	channels := make(map[string]chan string, len(profiles))
	all := make(map[string][]string, len(profiles))
	for profile, names := range profiles {
		queue := make(chan string, len(names))
		copied := append([]string(nil), names...)
		for _, name := range copied {
			queue <- name
		}
		channels[profile] = queue
		all[profile] = copied
	}
	return &LeasePool{channels: channels, all: all}
}

// Acquire 获取某个 profile 的一个独占 host 名称。
// 若该 profile 下所有 host 都被借出，调用方会阻塞，直到有测试归还租约。
func (p *LeasePool) Acquire(profile string) (string, error) {
	queue, ok := p.channels[profile]
	if !ok {
		return "", fmt.Errorf("unknown lease profile %q", profile)
	}
	return <-queue, nil
}

// Release 归还先前借出的 host 名称。
// 若 channel 已满，说明调用方重复归还或池状态被破坏，需要显式报错而不是静默吞掉。
func (p *LeasePool) Release(profile, name string) error {
	queue, ok := p.channels[profile]
	if !ok {
		return fmt.Errorf("unknown lease profile %q", profile)
	}
	select {
	case queue <- name:
		return nil
	default:
		return fmt.Errorf("lease profile %q is already full", profile)
	}
}

// Profiles 返回 profile 到 host 列表的副本，避免调用方修改内部状态。
func (p *LeasePool) Profiles() map[string][]string {
	result := make(map[string][]string, len(p.all))
	for profile, names := range p.all {
		result[profile] = append([]string(nil), names...)
	}
	return result
}
