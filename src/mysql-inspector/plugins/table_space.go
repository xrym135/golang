package plugins

import (
	"database/sql"
	"fmt"
	"mysql-inspector/types"
)

// TableSpacePlugin 表空间插件
type TableSpacePlugin struct {
	*BasePlugin
}

// NewTableSpacePlugin 创建表空间插件
func NewTableSpacePlugin(config types.PluginConfig) *TableSpacePlugin {
	base := NewBasePlugin("table_space", config)
	return &TableSpacePlugin{BasePlugin: base}
}

// Execute 执行表空间检查
func (p *TableSpacePlugin) Execute(db *sql.DB, snapshot *types.SnapshotData, lastSnapshot *types.SnapshotData) (*types.InspectionResult, error) {
	metrics := make(map[string]interface{})

	// 分析表空间使用情况
	totalSize, tableCount, fragmentedTables, largeTables := p.analyzeTableSpace(snapshot)
	metrics["total_size_gb"] = totalSize
	metrics["table_count"] = tableCount
	metrics["fragmented_tables"] = fragmentedTables
	metrics["large_tables"] = largeTables

	// 生成结果
	level := "INFO"
	message := fmt.Sprintf("表空间正常 - 总大小: %.2fGB, 表数量: %d", totalSize, tableCount)

	// 检查阈值
	maxTableSizeGB := float64(p.GetParamInt("max_table_size", 10*1024*1024*1024)) / 1024 / 1024 / 1024
	maxFragmentation := p.GetParamFloat("max_fragmentation", 30.0)

	if totalSize > maxTableSizeGB {
		level = "WARNING"
		message = fmt.Sprintf("表空间过大: %.2fGB (阈值: %.2fGB)", totalSize, maxTableSizeGB)
	}

	if len(largeTables) > 0 {
		level = "WARNING"
		message = fmt.Sprintf("发现大表: %d个，建议优化", len(largeTables))
		metrics["large_table_details"] = largeTables
	}

	if fragmentedTables > 0 {
		level = "WARNING"
		message = fmt.Sprintf("发现碎片表: %d个，碎片率超过 %.1f%%", fragmentedTables, maxFragmentation)
	}

	return p.CreateResult(level, message, metrics), nil
}

// analyzeTableSpace 分析表空间
func (p *TableSpacePlugin) analyzeTableSpace(snapshot *types.SnapshotData) (float64, int, int, []map[string]interface{}) {
	totalSize := 0.0
	tableCount := 0
	fragmentedTables := 0
	var largeTables []map[string]interface{}

	maxTableSizeBytes := p.GetParamInt("max_table_size", 10*1024*1024*1024) // 默认10GB
	maxFragmentation := p.GetParamFloat("max_fragmentation", 30.0)

	for tableKey, tableInfo := range snapshot.TableStats {
		if info, ok := tableInfo.(map[string]interface{}); ok {
			// 计算表总大小 (数据+索引)
			dataLength := getIntValue(info, "data_length")
			indexLength := getIntValue(info, "index_length")
			dataFree := getIntValue(info, "data_free")

			tableSizeGB := float64(dataLength+indexLength) / 1024 / 1024 / 1024 // GB
			totalSize += tableSizeGB
			tableCount++

			// 检查碎片率 (data_free / data_length)
			if dataLength > 0 {
				fragmentationRate := float64(dataFree) / float64(dataLength) * 100
				if fragmentationRate > maxFragmentation {
					fragmentedTables++
				}
			}

			// 检查大表
			if dataLength+indexLength > int64(maxTableSizeBytes) {
				largeTable := map[string]interface{}{
					"table":      tableKey,
					"size_gb":    tableSizeGB,
					"data_size":  dataLength,
					"index_size": indexLength,
				}
				largeTables = append(largeTables, largeTable)
			}
		}
	}

	return totalSize, tableCount, fragmentedTables, largeTables
}

// getIntValue 从interface{}获取int64值
func getIntValue(data map[string]interface{}, key string) int64 {
	if value, exists := data[key]; exists {
		switch v := value.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	return 0
}
