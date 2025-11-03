package plugins

import (
	"database/sql"
	"mysql-inspector/types"
	"time"
)

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Enabled 检查插件是否启用
	Enabled() bool

	// Execute 执行巡检
	Execute(db *sql.DB, snapshot *types.SnapshotData, lastSnapshot *types.SnapshotData) (*types.InspectionResult, error)

	// GetInterval 获取执行间隔
	GetInterval() int

	// GetLastRunTime 获取上次执行时间
	GetLastRunTime() time.Time

	// SetLastRunTime 设置上次执行时间
	SetLastRunTime(t time.Time)
}

// BasePlugin 插件基类
type BasePlugin struct {
	config     types.PluginConfig
	lastRun    time.Time
	pluginName string
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name string, config types.PluginConfig) *BasePlugin {
	return &BasePlugin{
		config:     config,
		pluginName: name,
		lastRun:    time.Time{},
	}
}

// Name 返回插件名称
func (p *BasePlugin) Name() string {
	return p.pluginName
}

// Enabled 检查插件是否启用
func (p *BasePlugin) Enabled() bool {
	return p.config.Enabled
}

// GetInterval 获取执行间隔
func (p *BasePlugin) GetInterval() int {
	return p.config.Interval
}

// GetLastRunTime 获取上次执行时间
func (p *BasePlugin) GetLastRunTime() time.Time {
	return p.lastRun
}

// SetLastRunTime 设置上次执行时间
func (p *BasePlugin) SetLastRunTime(t time.Time) {
	p.lastRun = t
}

// GetParam 获取插件参数
func (p *BasePlugin) GetParam(key string, defaultValue interface{}) interface{} {
	if value, exists := p.config.Params[key]; exists {
		return value
	}
	return defaultValue
}

// GetParamFloat 获取浮点数参数
func (p *BasePlugin) GetParamFloat(key string, defaultValue float64) float64 {
	value := p.GetParam(key, defaultValue)
	if floatValue, ok := value.(float64); ok {
		return floatValue
	}
	return defaultValue
}

// GetParamInt 获取整数参数
func (p *BasePlugin) GetParamInt(key string, defaultValue int) int {
	value := p.GetParam(key, defaultValue)
	if intValue, ok := value.(int); ok {
		return intValue
	}
	return defaultValue
}

// GetParamString 获取字符串参数
func (p *BasePlugin) GetParamString(key string, defaultValue string) string {
	value := p.GetParam(key, defaultValue)
	if strValue, ok := value.(string); ok {
		return strValue
	}
	return defaultValue
}

// CreateResult 创建巡检结果
func (p *BasePlugin) CreateResult(level, message string, metrics map[string]interface{}) *types.InspectionResult {
	return &types.InspectionResult{
		PluginName: p.Name(),
		Level:      level,
		Message:    message,
		Metrics:    metrics,
		Timestamp:  time.Now(),
	}
}
