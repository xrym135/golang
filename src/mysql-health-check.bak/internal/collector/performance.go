package collector

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"laowang/mysql-health-check/internal/logger"
	"laowang/mysql-health-check/pkg/core"
)

// PerformanceCollector 用于采集 MySQL 性能类指标（QPS、TPS、慢查询率等）。
type PerformanceCollector struct{}

func NewPerformanceCollector() *PerformanceCollector { return &PerformanceCollector{} }

func (c *PerformanceCollector) Name() string { return "performance" }

func (c *PerformanceCollector) Collect(ctx context.Context, db *sql.DB) (core.CollectorResult, error) {
	return core.CollectorResult{
		Name:    c.Name(),
		Metrics: map[string]float64{},
		Status:  core.HealthOK,
		Message: "use TakeSnapshot/CalculateDelta instead of Collect",
	}, nil
}

func (c *PerformanceCollector) TakeSnapshot(ctx context.Context, db *sql.DB) (core.Snapshot, error) {
	s := core.Snapshot{
		Name:   c.Name(),
		Values: map[string]float64{},
		Time:   time.Now(),
	}

	rows, err := db.QueryContext(ctx, `
		SHOW GLOBAL STATUS WHERE Variable_name IN (
			'Questions','Com_commit','Com_rollback','Slow_queries',
			'Bytes_sent','Bytes_received'
		)
	`)
	if err != nil {
		logger.Warn("性能采集器快照查询失败: %v", err)
		return s, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err != nil {
			logger.Debug("扫描性能指标失败: %v", err)
			continue
		}
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			s.Values[name] = v
		} else {
			logger.Debug("解析性能指标 %s 值失败: %v (raw=%s)", name, e, val)
		}
	}

	return s, nil
}

func (c *PerformanceCollector) CalculateDelta(prev, curr core.Snapshot) (core.CollectorResult, error) {
	cr := core.CollectorResult{
		Name:    c.Name(),
		Metrics: map[string]float64{},
		Status:  core.HealthOK,
	}

	// 计算两次快照间的时间差
	td := curr.Time.Sub(prev.Time).Seconds()
	if td <= 0 {
		logger.Warn("性能采集器时间差无效: %v", td)
		return cr, nil
	}

	// QPS = (Questions 差值) / 秒数
	if q1, ok1 := prev.Values["Questions"]; ok1 {
		if q2, ok2 := curr.Values["Questions"]; ok2 {
			cr.Metrics["qps"] = (q2 - q1) / td
		}
	}

	// TPS = (Com_commit+Com_rollback 差值) / 秒数
	if c1, ok1 := prev.Values["Com_commit"]; ok1 {
		if c2, ok2 := curr.Values["Com_commit"]; ok2 {
			var r1, r2 float64
			if rb1, ok := prev.Values["Com_rollback"]; ok {
				r1 = rb1
			}
			if rb2, ok := curr.Values["Com_rollback"]; ok {
				r2 = rb2
			}
			cr.Metrics["tps"] = (c2 + r2 - c1 - r1) / td
		}
	}

	// 慢查询率 (Slow_queries 差值 / 秒数)
	if s1, ok1 := prev.Values["Slow_queries"]; ok1 {
		if s2, ok2 := curr.Values["Slow_queries"]; ok2 {
			cr.Metrics["slow_qps"] = (s2 - s1) / td
			// 如果慢查询大于 1 QPS，则标记为警告
			if cr.Metrics["slow_qps"] > 1.0 {
				cr.Status = core.HealthWarn
				cr.Message = "慢查询较多"
			}
		}
	}

	return cr, nil
}
