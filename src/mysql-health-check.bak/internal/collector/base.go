package collector

import (
	"context"
	"database/sql"
	"strconv"

	"laowang/mysql-health-check/pkg/core"
)

type BaseCollector struct{}

func NewBaseCollector() *BaseCollector { return &BaseCollector{} }

func (c *BaseCollector) Name() string { return "base" }

func (c *BaseCollector) Collect(ctx context.Context, db *sql.DB) (core.CollectorResult, error) {
	cr := core.CollectorResult{
		Name:    c.Name(),
		Metrics: map[string]float64{},
		Status:  core.HealthOK,
		Message: "基础信息", // 简化消息，版本信息已在头部显示
	}

	// 读取 Uptime
	var name, val string
	rows, err := db.QueryContext(ctx, "SHOW GLOBAL STATUS LIKE 'Uptime'")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			_ = rows.Scan(&name, &val)
			if v, e := strconv.ParseFloat(val, 64); e == nil {
				cr.Metrics["uptime_s"] = v
				cr.Metrics["uptime_hours"] = v / 3600
				cr.Metrics["uptime_days"] = v / 86400

				// 若 uptime 小于 300s 标记警告
				if v < 300 {
					cr.Status = core.HealthWarn
					cr.Message = "新启动实例"
				}
			}
		}
	}

	return cr, nil
}

// GetVersion 辅助函数，用于获取版本信息
func (c *BaseCollector) GetVersion(ctx context.Context, db *sql.DB) (string, error) {
	var version string
	err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", err
	}
	return version, nil
}
