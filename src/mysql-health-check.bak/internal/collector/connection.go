package collector

import (
	"context"
	"database/sql"
	"strconv"

	"laowang/mysql-health-check/internal/logger"
	"laowang/mysql-health-check/pkg/core"
)

// ConnectionCollector 采集连接相关的状态信息。
type ConnectionCollector struct{}

func NewConnectionCollector() *ConnectionCollector { return &ConnectionCollector{} }
func (c *ConnectionCollector) Name() string        { return "connection" }

func (c *ConnectionCollector) Collect(ctx context.Context, db *sql.DB) (core.CollectorResult, error) {
	cr := core.CollectorResult{
		Name:    c.Name(),
		Metrics: map[string]float64{},
		Status:  core.HealthOK,
	}

	// 预设需要的字段，默认值为 -1（表示不可用/查询失败）
	expected := []string{
		"Threads_connected",
		"Threads_running",
		"Max_used_connections",
		"Aborted_connects",
		"Connection_errors_internal",
	}
	for _, k := range expected {
		cr.Metrics[k] = -1
	}

	rows, err := db.QueryContext(ctx, `
		SHOW GLOBAL STATUS WHERE Variable_name IN (
			'Threads_connected', 'Threads_running', 'Max_used_connections',
			'Aborted_connects', 'Connection_errors_internal'
		)
	`)
	if err != nil {
		logger.Warn("读取连接状态失败: %v", err)
		return cr, nil
	}
	defer rows.Close()

	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err != nil {
			logger.Warn("扫描连接状态失败: %v", err)
			continue
		}
		if v, e := strconv.ParseFloat(val, 64); e == nil {
			cr.Metrics[name] = v
		} else {
			logger.Debug("解析连接指标 %s 值失败: %v (raw=%s)", name, e, val)
		}
	}
	return cr, nil
}
