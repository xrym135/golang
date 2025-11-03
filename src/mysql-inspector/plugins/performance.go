package plugins

import (
	"database/sql"
	"fmt"
	"mysql-inspector/types"
	"strconv"
	// "time"
)

// PerformancePlugin 性能指标插件
type PerformancePlugin struct {
	*BasePlugin
}

// NewPerformancePlugin 创建性能指标插件
func NewPerformancePlugin(config types.PluginConfig) *PerformancePlugin {
	base := NewBasePlugin("performance", config)
	return &PerformancePlugin{BasePlugin: base}
}

// Execute 执行性能指标检查
func (p *PerformancePlugin) Execute(db *sql.DB, snapshot *types.SnapshotData, lastSnapshot *types.SnapshotData) (*types.InspectionResult, error) {
	metrics := make(map[string]interface{})

	// 计算QPS和TPS
	qps, tps := p.calculateQPSAndTPS(snapshot, lastSnapshot)
	metrics["qps"] = qps
	metrics["tps"] = tps

	// 检查连接数
	connections := p.getConnections(snapshot)
	metrics["connections"] = connections

	// 检查缓冲池命中率
	hitRate := p.getBufferPoolHitRate(snapshot)
	metrics["buffer_pool_hit_rate"] = hitRate

	// 生成结果消息
	level := "INFO"
	message := fmt.Sprintf("性能指标正常 - QPS: %.2f, TPS: %.2f, 连接数: %d", qps, tps, connections)

	// 检查阈值
	maxConnections := p.GetParamInt("max_connections", 500)
	qpsThreshold := p.GetParamFloat("qps_threshold", 1000.0)

	if connections > maxConnections {
		level = "WARNING"
		message = fmt.Sprintf("连接数过高: %d (阈值: %d)", connections, maxConnections)
	}

	if qps > qpsThreshold {
		level = "WARNING"
		message = fmt.Sprintf("QPS过高: %.2f (阈值: %.2f)", qps, qpsThreshold)
	}

	return p.CreateResult(level, message, metrics), nil
}

// calculateQPSAndTPS 计算QPS和TPS
func (p *PerformancePlugin) calculateQPSAndTPS(current *types.SnapshotData, previous *types.SnapshotData) (float64, float64) {
	if previous == nil {
		return 0, 0
	}

	// 计算时间间隔(秒)
	timeDiff := current.Timestamp.Sub(previous.Timestamp).Seconds()
	if timeDiff <= 0 {
		return 0, 0
	}

	// 获取查询次数
	currentQueries := p.getStatusValue(current, "Queries")
	previousQueries := p.getStatusValue(previous, "Queries")

	// 获取事务次数
	currentCommits := p.getStatusValue(current, "Com_commit")
	currentRollbacks := p.getStatusValue(current, "Com_rollback")
	previousCommits := p.getStatusValue(previous, "Com_commit")
	previousRollbacks := p.getStatusValue(previous, "Com_rollback")

	// 计算QPS
	qps := float64(currentQueries-previousQueries) / timeDiff

	// 计算TPS
	transactions := (currentCommits + currentRollbacks) - (previousCommits + previousRollbacks)
	tps := float64(transactions) / timeDiff

	return qps, tps
}

// getConnections 获取连接数
func (p *PerformancePlugin) getConnections(snapshot *types.SnapshotData) int {
	if value, exists := snapshot.GlobalStats["Threads_connected"]; exists {
		if strValue, ok := value.(string); ok {
			if intValue, err := strconv.Atoi(strValue); err == nil {
				return intValue
			}
		}
	}
	return 0
}

// getBufferPoolHitRate 获取缓冲池命中率
func (p *PerformancePlugin) getBufferPoolHitRate(snapshot *types.SnapshotData) float64 {
	reads := p.getStatusValue(snapshot, "Innodb_buffer_pool_reads")
	requests := p.getStatusValue(snapshot, "Innodb_buffer_pool_read_requests")

	if requests == 0 {
		return 0
	}

	hitRate := (1 - float64(reads)/float64(requests)) * 100
	return hitRate
}

// getStatusValue 获取状态值
func (p *PerformancePlugin) getStatusValue(snapshot *types.SnapshotData, key string) int64 {
	if value, exists := snapshot.GlobalStats[key]; exists {
		if strValue, ok := value.(string); ok {
			if intValue, err := strconv.ParseInt(strValue, 10, 64); err == nil {
				return intValue
			}
		}
	}
	return 0
}
