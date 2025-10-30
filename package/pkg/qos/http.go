package qos

import (
	"errors"
	"github.com/SianHH/frp-package/pkg/plugin/server"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

// QosConfig 代表远程下发的限速配置
type QosConfig struct {
	RPS   int // 每秒允许的请求数
	Burst int
}

// LimiterEntry 存储单个隧道的限速器信息
type LimiterEntry struct {
	limiter    *rate.Limiter
	lastUpdate time.Time // 上次更新时间
	lastUsed   time.Time // 上次被访问时间
}

// LimiterManager 管理所有隧道的限速器
type LimiterManager struct {
	mu          sync.RWMutex
	limiters    map[string]*LimiterEntry
	expireAfter time.Duration // 配置有效时长
	idleTimeout time.Duration // 清理间隔时长
	loadConfig  func(key string) (QosConfig, error)
	destroy     bool
}

// NewLimiterManager 创建一个懒加载限速器管理器
func NewLimiterManager(expire, idle time.Duration, loadConfig func(key string) (cfg QosConfig, err error)) *LimiterManager {
	m := &LimiterManager{
		limiters:    make(map[string]*LimiterEntry),
		expireAfter: expire,
		idleTimeout: idle,
		loadConfig:  loadConfig,
	}
	go m.cleanerLoop()
	return m
}

// GetLimiter 懒加载 + 异步刷新逻辑
func (m *LimiterManager) GetLimiter(tunnel string) *rate.Limiter {
	m.mu.RLock()
	entry, ok := m.limiters[tunnel]
	m.mu.RUnlock()

	if ok {
		entry.lastUsed = time.Now()
		// 检查是否过期
		if time.Since(entry.lastUpdate) > m.expireAfter {
			go m.refreshLimiter(tunnel) // 异步刷新配置
		}
		return entry.limiter
	}

	// 异步加载配置
	go m.refreshLimiter(tunnel)

	// 先返回一个“临时默认限速器”（默认不过滤）
	return rate.NewLimiter(rate.Inf, 0)
}

// refreshLimiter 异步加载或更新配置
func (m *LimiterManager) refreshLimiter(key string) {
	cfg, err := m.loadConfig(key)
	oldLimiter := m.GetLimiter(key)
	var limiter *rate.Limiter
	var entry *LimiterEntry
	if err != nil {
		if errors.Is(err, server.ErrorPluginsSendFail) {
			// 如果是网络原因导致的，依然采用旧的Limiter
			entry = &LimiterEntry{
				limiter:    oldLimiter,
				lastUpdate: time.Now(),
				lastUsed:   time.Now(),
			}
		} else {
			// 否则就不允许请求
			limiter = rate.NewLimiter(0, 0)
			entry = &LimiterEntry{
				limiter:    limiter,
				lastUpdate: time.Now(),
				lastUsed:   time.Now(),
			}
		}
	} else {
		if cfg.RPS == 0 && cfg.Burst == 0 {
			// 如果都是0，则不限制
			limiter = rate.NewLimiter(rate.Inf, 0)
		} else {
			// 否则按返回的配置设置限速器
			limiter = rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst)
		}
		entry = &LimiterEntry{
			limiter:    limiter,
			lastUpdate: time.Now(),
			lastUsed:   time.Now(),
		}
	}
	m.mu.Lock()
	m.limiters[key] = entry
	m.mu.Unlock()
}

// Allow 检查是否允许当前请求
func (m *LimiterManager) Allow(key string) bool {
	limiter := m.GetLimiter(key)
	return limiter.Allow()
}

func (m *LimiterManager) Destroy() {
	m.destroy = true
}

// 后台定期清理协程
func (m *LimiterManager) cleanerLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if m.destroy {
			return
		}
		now := time.Now()
		m.mu.Lock()
		for name, entry := range m.limiters {
			if now.Sub(entry.lastUsed) > m.idleTimeout {
				delete(m.limiters, name)
			}
		}
		m.mu.Unlock()
	}
}
