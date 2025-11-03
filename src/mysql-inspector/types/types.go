package types

import (
	"time"
)

// InspectionResult 巡检结果
type InspectionResult struct {
	PluginName string                 `json:"plugin_name"`
	Level      string                 `json:"level"`     // INFO, WARNING, ERROR
	Message    string                 `json:"message"`   // 描述信息
	Metrics    map[string]interface{} `json:"metrics"`   // 指标数据
	Timestamp  time.Time              `json:"timestamp"` // 检查时间
}

// SnapshotData 快照数据
type SnapshotData struct {
	Timestamp   time.Time                `json:"timestamp"`
	GlobalStats map[string]interface{}   `json:"global_stats"` // SHOW GLOBAL STATUS 结果
	TableStats  map[string]interface{}   `json:"table_stats"`  // 表统计信息
	ProcessList []map[string]interface{} `json:"process_list"` // 进程列表
	CustomData  map[string]interface{}   `json:"custom_data"`  // 插件自定义数据
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled  bool                   `yaml:"enabled"`
	Interval int                    `yaml:"interval"` // 执行间隔(秒)
	Params   map[string]interface{} `yaml:"params"`   // 插件参数
}

// MySQLConfig MySQL连接配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Timeout  int    `yaml:"timeout"` // 连接超时(秒)
}

// InspectionConfig 巡检配置
type InspectionConfig struct {
	SnapshotInterval int    `yaml:"snapshot_interval"` // 快照采集间隔
	Interval         int    `yaml:"interval"`          // 巡检执行间隔
	MaxSnapshots     int    `yaml:"max_snapshots"`     // 最大快照数量
	LogLevel         string `yaml:"log_level"`         // 日志级别
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	MySQL      MySQLConfig             `yaml:"mysql"`
	Inspection InspectionConfig        `yaml:"inspection"`
	Plugins    map[string]PluginConfig `yaml:"plugins"`
	Thresholds map[string]interface{}  `yaml:"thresholds"` // 阈值配置
}
