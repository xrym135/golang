package core

import (
	"mysql-inspector/plugins"
	"mysql-inspector/utils"
)

// PluginManager 插件管理器
type PluginManager struct {
	plugins []plugins.Plugin
	logger  *utils.Logger
}

// NewPluginManager 创建插件管理器
func NewPluginManager(logger *utils.Logger) *PluginManager {
	return &PluginManager{
		plugins: make([]plugins.Plugin, 0),
		logger:  logger,
	}
}

// RegisterPlugin 注册插件
func (pm *PluginManager) RegisterPlugin(plugin plugins.Plugin) {
	if plugin.Enabled() {
		pm.plugins = append(pm.plugins, plugin)
		pm.logger.Info("注册插件: %s", plugin.Name())
	}
}

// GetPlugins 获取所有插件
func (pm *PluginManager) GetPlugins() []plugins.Plugin {
	return pm.plugins
}

// ShouldExecute 检查插件是否应该执行
func (pm *PluginManager) ShouldExecute(plugin plugins.Plugin) bool {
	// 这里可以添加执行条件判断
	// 例如基于时间间隔等
	return true
}
