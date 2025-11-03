package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"mysql-inspector/types"
	"os"
	"path/filepath"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*types.GlobalConfig, error) {
	var config types.GlobalConfig

	fmt.Printf("正在加载配置文件: %s\n", configPath)

	// 如果配置文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("配置文件不存在，使用默认配置\n")
		return getDefaultConfig(), nil
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	fmt.Printf("配置文件加载成功\n")

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	fmt.Printf("配置解析成功 - MySQL主机: %s, 端口: %d\n", config.MySQL.Host, config.MySQL.Port)
	return &config, nil
}

// getDefaultConfig 返回默认配置
func getDefaultConfig() *types.GlobalConfig {
	config := &types.GlobalConfig{}

	// MySQL 默认配置
	config.MySQL.Host = "localhost"
	config.MySQL.Port = 3306
	config.MySQL.Username = "root"
	config.MySQL.Database = "information_schema"
	config.MySQL.Timeout = 10

	// 巡检默认配置
	config.Inspection.SnapshotInterval = 60 // 快照间隔60秒
	config.Inspection.Interval = 300        // 巡检间隔300秒
	config.Inspection.MaxSnapshots = 10
	config.Inspection.LogLevel = "INFO"

	// 默认插件配置
	config.Plugins = make(map[string]types.PluginConfig)
	config.Plugins["performance"] = types.PluginConfig{
		Enabled:  true,
		Interval: 60,
		Params: map[string]interface{}{
			"max_connections": 500,
			"qps_threshold":   1000.0,
			"tps_threshold":   500.0,
		},
	}

	config.Plugins["slow_query"] = types.PluginConfig{
		Enabled:  true,
		Interval: 300,
		Params: map[string]interface{}{
			"slow_query_threshold": 2.0,
		},
	}

	// 默认阈值配置
	config.Thresholds = map[string]interface{}{
		"cpu_usage":       80.0,
		"memory_usage":    85.0,
		"disk_usage":      90.0,
		"slow_query_time": 2.0,
		"max_connections": 500,
		"qps_threshold":   1000.0,
		"tps_threshold":   500.0,
	}

	return config
}

// SaveConfig 保存配置到文件
func SaveConfig(config *types.GlobalConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
