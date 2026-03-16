package core

import (
	"fmt"
	"sync"
)

// StrategyRegistry 策略注册表
type StrategyRegistry struct {
	strategies map[MigrationType]MigrationStrategy
	mu         sync.RWMutex
}

// globalRegistry 全局策略注册表
var globalRegistry = NewStrategyRegistry()

// NewStrategyRegistry 创建策略注册表
func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[MigrationType]MigrationStrategy),
	}
}

// Register 注册策略
func (r *StrategyRegistry) Register(strategy MigrationStrategy) error {
	if strategy == nil {
		return fmt.Errorf("strategy cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	migrationType := strategy.Type()
	if _, exists := r.strategies[migrationType]; exists {
		return fmt.Errorf("strategy for type %s already registered", migrationType)
	}

	r.strategies[migrationType] = strategy
	return nil
}

// Unregister 注销策略
func (r *StrategyRegistry) Unregister(migrationType MigrationType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.strategies, migrationType)
}

// Get 获取策略
func (r *StrategyRegistry) Get(migrationType MigrationType) (MigrationStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategy, exists := r.strategies[migrationType]
	if !exists {
		return nil, fmt.Errorf("no strategy registered for type: %s", migrationType)
	}
	return strategy, nil
}

// Has 检查策略是否存在
func (r *StrategyRegistry) Has(migrationType MigrationType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.strategies[migrationType]
	return exists
}

// List 获取所有已注册的策略列表
func (r *StrategyRegistry) List() []MigrationStrategy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategies := make([]MigrationStrategy, 0, len(r.strategies))
	for _, strategy := range r.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// ListTypes 获取所有已注册的类型列表
func (r *StrategyRegistry) ListTypes() []MigrationType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]MigrationType, 0, len(r.strategies))
	for migrationType := range r.strategies {
		types = append(types, migrationType)
	}
	return types
}

// Count 获取已注册策略数量
func (r *StrategyRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.strategies)
}

// Clear 清空所有策略
func (r *StrategyRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategies = make(map[MigrationType]MigrationStrategy)
}

// 全局注册表操作函数

// RegisterStrategy 注册策略到全局注册表
func RegisterStrategy(strategy MigrationStrategy) error {
	return globalRegistry.Register(strategy)
}

// UnregisterStrategy 从全局注册表注销策略
func UnregisterStrategy(migrationType MigrationType) {
	globalRegistry.Unregister(migrationType)
}

// GetStrategy 从全局注册表获取策略
func GetStrategy(migrationType MigrationType) (MigrationStrategy, error) {
	return globalRegistry.Get(migrationType)
}

// HasStrategy 检查全局注册表中是否存在策略
func HasStrategy(migrationType MigrationType) bool {
	return globalRegistry.Has(migrationType)
}

// ListStrategies 获取全局注册表中所有策略
func ListStrategies() []MigrationStrategy {
	return globalRegistry.List()
}

// ListStrategyTypes 获取全局注册表中所有类型
func ListStrategyTypes() []MigrationType {
	return globalRegistry.ListTypes()
}

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *StrategyRegistry {
	return globalRegistry
}
