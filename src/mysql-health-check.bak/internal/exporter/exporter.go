package exporter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"laowang/mysql-health-check/internal/collector"
)

type Exporter interface {
	Export(res *collector.AggregatedResult) error
}

type TextExporter struct{}

func NewTextExporter() *TextExporter { return &TextExporter{} }

func (e *TextExporter) Export(res *collector.AggregatedResult) error {
	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Printf("MySQL 健康检查报告\n")
	fmt.Printf("%s\n", strings.Repeat("=", 70))

	// 实例信息
	fmt.Printf("实例地址: %s:%d\n", res.Instance.Host, res.Instance.Port)
	if res.Instance.Version != "" {
		fmt.Printf("MySQL版本: %s\n", res.Instance.Version)
	}
	fmt.Printf("检查时间: %s\n", res.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("执行耗时: %v\n", res.Duration.Round(time.Millisecond))

	// 总体状态
	statusColor := e.getStatusColor(res.Overall)
	fmt.Printf("总体状态: %s\n", statusColor(res.Overall.String()))
	fmt.Printf("%s\n", strings.Repeat("-", 70))

	// 按照固定顺序显示采集器
	collectorOrder := []string{"base", "connection", "buffer_pool", "replication", "performance"}
	for _, name := range collectorOrder {
		if r, exists := res.Details[name]; exists {
			e.displayCollectorResult(name, r)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Printf("统计摘要: OK=%d, WARN=%d, CRITICAL=%d\n",
		res.Summary[collector.HealthOK],
		res.Summary[collector.HealthWarn],
		res.Summary[collector.HealthCritical])
	fmt.Printf("%s\n", strings.Repeat("=", 70))

	return nil
}

func (e *TextExporter) displayCollectorResult(name string, r collector.CollectorResult) {
	statusColor := e.getStatusColor(r.Status)

	switch name {
	case "base":
		fmt.Printf("\n[%s] %s - 基础信息\n", name, statusColor(r.Status.String()))
		// base部分不显示版本信息，只显示运行时间等指标
		e.displayOrderedMetrics(r.Metrics, []string{"uptime_days", "uptime_hours", "uptime_s"})

	case "connection":
		fmt.Printf("\n[%s] %s - 连接状态\n", name, statusColor(r.Status.String()))
		e.displayOrderedMetrics(r.Metrics, []string{
			"Threads_connected",
			"Threads_running",
			"Max_used_connections",
			"Aborted_connects",
			"Connection_errors_internal",
		})

	case "buffer_pool":
		fmt.Printf("\n[%s] %s - InnoDB缓冲池\n", name, statusColor(r.Status.String()))
		e.displayBufferPoolMetrics(r.Metrics)

	case "replication":
		fmt.Printf("\n[%s] %s - %s\n", name, statusColor(r.Status.String()), r.Message)
		if len(r.Metrics) > 0 {
			e.displayOrderedMetrics(r.Metrics, []string{"seconds_behind_master"})
		}

	case "performance":
		fmt.Printf("\n[%s] %s - 性能指标\n", name, statusColor(r.Status.String()))
		e.displayOrderedMetrics(r.Metrics, []string{"qps", "tps", "slow_qps"})

	default:
		fmt.Printf("\n[%s] %s - %s\n", name, statusColor(r.Status.String()), r.Message)
		e.displayGenericMetrics(r.Metrics)
	}
}

func (e *TextExporter) displayBufferPoolMetrics(metrics map[string]float64) {
	// 定义缓冲池指标的分组和顺序
	groups := []struct {
		title   string
		metrics []string
	}{
		{
			"缓冲池大小",
			[]string{"innodb_buffer_pool_size"},
		},
		{
			"页面分布",
			[]string{
				"Innodb_buffer_pool_pages_total",
				"Innodb_buffer_pool_pages_data",
				"Innodb_buffer_pool_pages_free",
				"Innodb_buffer_pool_pages_dirty",
				"Innodb_buffer_pool_pages_misc",
				"Innodb_buffer_pool_pages_old",
			},
		},
		{
			"读写统计",
			[]string{
				"Innodb_buffer_pool_read_requests",
				"Innodb_buffer_pool_reads",
				"Innodb_buffer_pool_write_requests",
				"Innodb_buffer_pool_pages_flushed",
				"hit_rate",
			},
		},
		{
			"预读统计",
			[]string{
				"Innodb_buffer_pool_read_ahead",
				"Innodb_buffer_pool_read_ahead_rnd",
				"Innodb_buffer_pool_read_ahead_evicted",
			},
		},
		{
			"其他指标",
			[]string{
				"Innodb_buffer_pool_bytes_data",
				"Innodb_buffer_pool_bytes_dirty",
				"Innodb_buffer_pool_pages_made_young",
				"Innodb_buffer_pool_pages_made_not_young",
				"Innodb_buffer_pool_pages_LRU_flushed",
				"Innodb_buffer_pool_wait_free",
				"Innodb_buffer_pool_resize_status_code",
				"Innodb_buffer_pool_resize_status_progress",
			},
		},
	}

	for _, group := range groups {
		// 检查该组中是否有实际存在的指标
		hasMetrics := false
		for _, metric := range group.metrics {
			if _, exists := metrics[metric]; exists {
				hasMetrics = true
				break
			}
		}

		if hasMetrics {
			fmt.Printf("  %s:\n", group.title)
			e.displayOrderedMetrics(metrics, group.metrics)
		}
	}
}

func (e *TextExporter) displayOrderedMetrics(metrics map[string]float64, order []string) {
	maxKeyLen := 0
	for _, k := range order {
		if _, exists := metrics[k]; exists && len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	for _, k := range order {
		if v, exists := metrics[k]; exists && v != -1 {
			padding := strings.Repeat(" ", maxKeyLen-len(k))
			formattedValue := e.formatMetricValue(k, v)
			fmt.Printf("  %s%s: %s\n", k, padding, formattedValue)
		}
	}
}

func (e *TextExporter) displayGenericMetrics(metrics map[string]float64) {
	keys := make([]string, 0, len(metrics))
	for k := range metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	e.displayOrderedMetrics(metrics, keys)
}

func (e *TextExporter) formatMetricValue(key string, value float64) string {
	// 特殊处理某些指标的显示格式
	switch {
	case value == -1:
		return "N/A"
	case strings.Contains(key, "rate") || strings.Contains(key, "hit_rate"):
		return fmt.Sprintf("%.2f%%", value*100)
	case key == "innodb_buffer_pool_size":
		return fmt.Sprintf("%.0f MB", value)
	case value == float64(int64(value)):
		// 对大数值使用千位分隔符
		intVal := int64(value)
		if intVal > 1000 {
			return e.formatWithCommas(intVal)
		}
		return fmt.Sprintf("%d", intVal)
	case value < 0.01:
		return fmt.Sprintf("%.6f", value)
	case value < 0.1:
		return fmt.Sprintf("%.5f", value)
	case value < 1:
		return fmt.Sprintf("%.4f", value)
	case value < 10:
		return fmt.Sprintf("%.3f", value)
	case value < 100:
		return fmt.Sprintf("%.2f", value)
	case value < 1000:
		return fmt.Sprintf("%.1f", value)
	default:
		return e.formatWithCommas(int64(value))
	}
}

func (e *TextExporter) formatWithCommas(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	parts := []string{}
	for n > 0 {
		parts = append([]string{fmt.Sprintf("%03d", n%1000)}, parts...)
		n /= 1000
	}

	// 去除前导零
	result := strings.TrimLeft(parts[0], "0")
	if result == "" {
		result = "0"
	}
	for i := 1; i < len(parts); i++ {
		result += "," + parts[i]
	}
	return result
}

func (e *TextExporter) getStatusColor(status collector.HealthLevel) func(string) string {
	colorCodes := map[collector.HealthLevel]string{
		collector.HealthOK:       "32", // 绿色
		collector.HealthWarn:     "33", // 黄色
		collector.HealthCritical: "31", // 红色
	}

	return func(text string) string {
		code := colorCodes[status]
		return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
	}
}

type JSONExporter struct{}

func NewJSONExporter() *JSONExporter { return &JSONExporter{} }

func (e *JSONExporter) Export(res *collector.AggregatedResult) error {
	enhancedResult := map[string]interface{}{
		"timestamp": res.Timestamp,
		"instance": map[string]interface{}{
			"host":    res.Instance.Host,
			"port":    res.Instance.Port,
			"version": res.Instance.Version,
		},
		"duration_seconds": res.Duration.Seconds(),
		"overall_status":   res.Overall.String(),
		"summary":          res.Summary,
		"details":          res.Details,
	}

	b, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
