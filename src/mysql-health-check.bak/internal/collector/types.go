package collector

import (
	"time"

	"laowang/mysql-health-check/pkg/core"
)

// 类型别名，简化在 collector 包中使用 core 包类型时的写法。
type (
	Snapshot        = core.Snapshot
	CollectorResult = core.CollectorResult
	HealthLevel     = core.HealthLevel
	InstanceInfo    = core.InstanceInfo
)

// 常量别名，直接复用 core 包的健康等级。
const (
	HealthOK       = core.HealthOK
	HealthWarn     = core.HealthWarn
	HealthCritical = core.HealthCritical
)

// AggregatedResult 表示一次完整健康检查的聚合结果。
// 包含：
// - Timestamp：采集时间
// - Details：每个采集器的结果
// - Summary：按健康等级统计各采集器数量
// - Overall：总体健康结论

type AggregatedResult struct {
	Timestamp time.Time
	Instance  InstanceInfo
	Details   map[string]CollectorResult
	Summary   map[HealthLevel]int
	Overall   HealthLevel
	Duration  time.Duration // 执行耗时
}

// AggregateOverall 根据 summary 中的分布，得出整体健康状态。
// 规则：
// - 只要有 Critical，则整体 Critical
// - 否则如果有 Warn，则整体 Warn
// - 否则整体 OK
func AggregateOverall(summary map[HealthLevel]int) HealthLevel {
	if summary[HealthCritical] > 0 {
		return HealthCritical
	}
	if summary[HealthWarn] > 0 {
		return HealthWarn
	}
	return HealthOK
}
