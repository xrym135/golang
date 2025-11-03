package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"mysql-inspector/config"
	"mysql-inspector/core"
	"mysql-inspector/plugins"
	"mysql-inspector/types"
	"mysql-inspector/utils"
	"os"
	"os/signal"
	"syscall"
	// "time"
)

var (
	configPath = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 检查配置文件是否存在
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Printf("配置文件不存在: %s\n", *configPath)
		// 尝试当前目录
		*configPath = "config.yaml"
		if _, err := os.Stat(*configPath); os.IsNotExist(err) {
			fmt.Printf("默认配置文件也不存在: %s\n", *configPath)
			os.Exit(1)
		}
		fmt.Printf("使用默认配置文件: %s\n", *configPath)
	} else {
		fmt.Printf("找到配置文件: %s\n", *configPath)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Inspection.LogLevel)
	logger.Info("MySQL巡检工具启动...")

	// 详细打印配置信息
	logger.Debug("=== 配置信息验证 ===")
	logger.Debug("MySQL配置 - 主机: %s, 端口: %d", cfg.MySQL.Host, cfg.MySQL.Port)
	logger.Debug("MySQL配置 - 用户名: %s, 数据库: %s", cfg.MySQL.Username, cfg.MySQL.Database)
	logger.Debug("巡检配置 - 间隔: %d秒, 最大快照: %d", cfg.Inspection.Interval, cfg.Inspection.MaxSnapshots)
	logger.Debug("日志级别: %s", cfg.Inspection.LogLevel)

	// 连接数据库
	logger.Info("连接数据库: %s:%d", cfg.MySQL.Host, cfg.MySQL.Port)
	db, err := connectDatabase(cfg, logger)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// 初始化快照管理器
	snapshotManager := core.NewSnapshotManager(db, cfg.Inspection.MaxSnapshots, logger)

	// 初始化插件管理器
	pluginManager := core.NewPluginManager(logger)

	// 注册插件
	registerPlugins(pluginManager, cfg.Plugins)

	// 初始化巡检器
	inspector := core.NewInspector(db, snapshotManager, pluginManager, logger)

	// 启动巡检
	go inspector.Start(cfg.Inspection.Interval)

	// 等待退出信号
	waitForShutdown(inspector, logger)
}

// connectDatabase 连接数据库
func connectDatabase(cfg *config.GlobalConfig, logger *utils.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&parseTime=true",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
		cfg.MySQL.Timeout,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	logger.Info("成功连接到MySQL数据库: %s:%d", cfg.MySQL.Host, cfg.MySQL.Port)
	return db, nil
}

// registerPlugins 注册插件
func registerPlugins(manager *core.PluginManager, pluginConfigs map[string]types.PluginConfig) {
	// 性能指标插件
	if config, exists := pluginConfigs["performance"]; exists {
		manager.RegisterPlugin(plugins.NewPerformancePlugin(config))
	}

	// 慢查询插件
	if config, exists := pluginConfigs["slow_query"]; exists {
		manager.RegisterPlugin(plugins.NewSlowQueryPlugin(config))
	}

	// 这里可以添加更多插件...
}

// waitForShutdown 等待关闭信号
func waitForShutdown(inspector *core.Inspector, logger *utils.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("接收到关闭信号，正在停止巡检...")

	inspector.Stop()
	logger.Info("MySQL巡检工具已停止")
}
