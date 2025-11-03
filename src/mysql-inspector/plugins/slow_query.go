package plugins

import (
	"database/sql"
	"fmt"
	"mysql-inspector/types"
	"strconv"
)

// SlowQueryPlugin 慢查询插件
type SlowQueryPlugin struct {
	*BasePlugin
}

// NewSlowQueryPlugin 创建慢查询插件
func NewSlowQueryPlugin(config types.PluginConfig) *SlowQueryPlugin {
	base := NewBasePlugin("slow_query", config)
	return &SlowQueryPlugin{BasePlugin: base}
}

// Execute 执行慢查询检查
func (p *SlowQueryPlugin) Execute(db *sql.DB, snapshot *types.SnapshotData, lastSnapshot *types.SnapshotData) (*types.InspectionResult, error) {
	metrics := make(map[string]interface{})

	// 检查慢查询配置
	slowQueryEnabled := p.checkSlowQueryEnabled(db)
	metrics["slow_query_enabled"] = slowQueryEnabled

	// 获取慢查询阈值
	slowQueryTime := p.getSlowQueryTime(db)
	metrics["slow_query_time"] = slowQueryTime

	// 获取慢查询数量
	slowQueries := p.getSlowQueriesCount(snapshot)
	metrics["slow_queries"] = slowQueries

	// 生成结果
	level := "INFO"
	message := fmt.Sprintf("慢查询配置正常 - 阈值: %.2f秒, 数量: %d", slowQueryTime, slowQueries)

	// 检查阈值
	threshold := p.GetParamFloat("slow_query_threshold", 2.0)
	if slowQueryTime > threshold {
		level = "WARNING"
		message = fmt.Sprintf("慢查询阈值设置过高: %.2f秒", slowQueryTime)
	}

	if slowQueries > 0 {
		level = "WARNING"
		message = fmt.Sprintf("发现慢查询: %d个", slowQueries)
	}

	if !slowQueryEnabled {
		level = "WARNING"
		message = "慢查询日志未开启"
	}

	return p.CreateResult(level, message, metrics), nil
}

// checkSlowQueryEnabled 检查慢查询是否启用
func (p *SlowQueryPlugin) checkSlowQueryEnabled(db *sql.DB) bool {
	var slowQueryLog string
	var logOutput string

	err := db.QueryRow("SHOW VARIABLES LIKE 'slow_query_log'").Scan(&slowQueryLog, &logOutput)
	if err != nil {
		return false
	}

	return slowQueryLog == "ON"
}

// getSlowQueryTime 获取慢查询时间阈值
func (p *SlowQueryPlugin) getSlowQueryTime(db *sql.DB) float64 {
	var variableName string
	var value string

	err := db.QueryRow("SHOW VARIABLES LIKE 'long_query_time'").Scan(&variableName, &value)
	if err != nil {
		return 10.0 // 默认值
	}

	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}

	return 10.0
}

// getSlowQueriesCount 获取慢查询数量
func (p *SlowQueryPlugin) getSlowQueriesCount(snapshot *types.SnapshotData) int64 {
	return p.getStatusValue(snapshot, "Slow_queries")
}

// getStatusValue 获取状态值
func (p *SlowQueryPlugin) getStatusValue(snapshot *types.SnapshotData, key string) int64 {
	if value, exists := snapshot.GlobalStats[key]; exists {
		if strValue, ok := value.(string); ok {
			if intValue, err := strconv.ParseInt(strValue, 10, 64); err == nil {
				return intValue
			}
		}
	}
	return 0
}
