package collector

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"laowang/mysql-health-check/internal/logger"
	"laowang/mysql-health-check/pkg/core"
)

// BufferPoolCollector 检查 InnoDB 缓冲池相关指标。
type BufferPoolCollector struct{}

func NewBufferPoolCollector() *BufferPoolCollector { return &BufferPoolCollector{} }
func (c *BufferPoolCollector) Name() string        { return "buffer_pool" }

func (c *BufferPoolCollector) Collect(ctx context.Context, db *sql.DB) (core.CollectorResult, error) {
	cr := core.CollectorResult{
		Name:    c.Name(),
		Metrics: map[string]float64{},
		Status:  core.HealthOK,
	}

	// 预置常用字段，默认 -1 表示不可用
	expected := []string{
		"innodb_buffer_pool_size", // MB
		"Innodb_buffer_pool_pages_data",
		"Innodb_buffer_pool_pages_dirty",
		"Innodb_buffer_pool_read_requests",
		"Innodb_buffer_pool_reads",
		"Innodb_buffer_pool_pages_free",
		"Innodb_buffer_pool_pages_total",
		"Innodb_buffer_pool_pages_misc",
		"Innodb_buffer_pool_pages_old",
		"Innodb_buffer_pool_write_requests",
		"Innodb_buffer_pool_pages_flushed",
	}
	for _, k := range expected {
		cr.Metrics[k] = -1
	}

	// 1) 读取 Innodb_buffer_pool_* 系列的 GLOBAL STATUS
	// 只查询数值型的状态变量，避免字符串类型的变量
	rows, err := db.QueryContext(ctx, `
		SHOW GLOBAL STATUS WHERE Variable_name LIKE 'Innodb_buffer_pool%' 
		AND Variable_name NOT LIKE '%status%'
		AND Variable_name NOT LIKE '%dump%'
		AND Variable_name NOT LIKE '%load%'
	`)
	if err != nil {
		logger.Warn("读取 Innodb_buffer_pool_* 状态失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var name, val string
			if err := rows.Scan(&name, &val); err != nil {
				logger.Debug("扫描 buffer_pool 状态失败: %v", err)
				continue
			}

			// 跳过空值或非数字值
			if val == "" || strings.Contains(strings.ToLower(val), "not started") ||
				strings.Contains(strings.ToLower(val), "completed") {
				logger.Debug("跳过非数值状态变量: %s = %s", name, val)
				continue
			}

			if v, e := strconv.ParseFloat(val, 64); e == nil {
				cr.Metrics[name] = v
			} else {
				logger.Debug("解析 buffer_pool 指标 %s 值失败: %v (raw=%s)", name, e, val)
			}
		}
	}

	// 2) 单独读取 innodb_buffer_pool_size（bytes -> MB）
	var poolSizeInt sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT @@innodb_buffer_pool_size").Scan(&poolSizeInt); err != nil {
		logger.Warn("读取 innodb_buffer_pool_size 失败: %v", err)
	} else if poolSizeInt.Valid {
		if v, e := strconv.ParseFloat(poolSizeInt.String, 64); e == nil {
			// 转换为 MB
			cr.Metrics["innodb_buffer_pool_size"] = v / 1024.0 / 1024.0
		} else {
			logger.Debug("解析 innodb_buffer_pool_size 值失败: %v (raw=%s)", e, poolSizeInt.String)
		}
	}

	// 计算命中率（如果有读请求与读取次数）
	if req, ok1 := cr.Metrics["Innodb_buffer_pool_read_requests"]; ok1 && req >= 0 {
		if reads, ok2 := cr.Metrics["Innodb_buffer_pool_reads"]; ok2 && reads >= 0 && req > 0 {
			hit := 1.0 - reads/req
			cr.Metrics["hit_rate"] = hit
			if hit < 0 {
				cr.Metrics["hit_rate"] = -1
			}
		}
	}

	return cr, nil
}
