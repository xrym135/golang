package collector

import (
	"context"
	"database/sql"
	"strconv"

	"laowang/mysql-health-check/pkg/core"
)

type ReplicationCollector struct{}

func NewReplicationCollector() *ReplicationCollector { return &ReplicationCollector{} }
func (c *ReplicationCollector) Name() string         { return "replication" }

func (c *ReplicationCollector) Collect(ctx context.Context, db *sql.DB) (core.CollectorResult, error) {
	cr := core.CollectorResult{
		Name:    c.Name(),
		Metrics: map[string]float64{},
		Status:  core.HealthOK,
		Message: "主库或未配置复制",
	}
	// 尝试 SHOW SLAVE STATUS; 若无行则不是从库
	rows, err := db.QueryContext(ctx, "SHOW SLAVE STATUS")
	if err != nil {
		// 非致命，返回默认
		return cr, nil
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	if !rows.Next() {
		return cr, nil
	}
	// 动态 scan 所有列为字符串
	values := make([]sql.NullString, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return cr, nil
	}
	// 尝试抽取 Seconds_Behind_Master 字段
	for i, n := range cols {
		if n == "Seconds_Behind_Master" {
			if values[i].Valid {
				if v, e := strconv.ParseFloat(values[i].String, 64); e == nil {
					cr.Metrics["seconds_behind_master"] = v
					if v > 300 {
						cr.Status = core.HealthWarn
						cr.Message = "复制延迟较高"
					}
				}
			}
		}
	}
	cr.Message = "检测到从库"
	return cr, nil
}
