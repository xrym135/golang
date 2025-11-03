package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"laowang/mysql-health-check/internal/collector"
	"laowang/mysql-health-check/internal/config"
	"laowang/mysql-health-check/internal/exporter"
	"laowang/mysql-health-check/internal/logger"
	"laowang/mysql-health-check/pkg/dbconn"
)

func main() {
	cfgPath := flag.String("config", "", "配置文件路径（可选）")
	flag.Parse()

	cfg := config.NewConfig()
	if *cfgPath != "" {
		if err := cfg.LoadFromFile(*cfgPath); err != nil {
			fmt.Printf("配置文件加载失败: %v\n", err)
			os.Exit(1)
		}
	}

	logger.InitLogger(cfg.GetLogLevel())
	logger.Info("启动 MySQL 健康检查")

	// 建立数据库连接
	db, err := dbconn.NewConnection(dbconn.Config{
		Host:           cfg.GetHost(),
		Port:           cfg.GetPort(),
		User:           cfg.GetUser(),
		Password:       cfg.GetPassword(),
		Database:       cfg.GetDatabase(),
		ConnectTimeout: cfg.GetQueryTimeout(),
		QueryTimeout:   cfg.GetQueryTimeout(),
	})
	if err != nil {
		logger.Fatal("数据库连接失败: %v", err)
	}
	defer db.Close()

	// CollectorManager 与注册
	manager := collector.NewCollectorManager()
	manager.RegisterSingleCollector(collector.NewBaseCollector())
	manager.RegisterSingleCollector(collector.NewConnectionCollector())
	manager.RegisterSingleCollector(collector.NewBufferPoolCollector())
	manager.RegisterSingleCollector(collector.NewReplicationCollector())
	manager.RegisterPeriodicCollector(collector.NewPerformanceCollector())

	// 执行健康检查
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GetTotalTimeout())
	defer cancel()

	startTime := time.Now()
	result, err := runHealthCheck(ctx, db, manager, cfg)
	duration := time.Since(startTime)

	if err != nil {
		logger.Fatal("健康检查失败: %v", err)
	}

	// 设置执行耗时
	result.Duration = duration

	// 输出
	var exp exporter.Exporter = exporter.NewTextExporter()
	if cfg.OutputFormat() == "json" {
		exp = exporter.NewJSONExporter()
	}
	if err := exp.Export(result); err != nil {
		logger.Error("导出结果失败: %v", err)
	}

	logger.Info("健康检查完成")
}

func runHealthCheck(ctx context.Context, db *sql.DB, manager *collector.CollectorManager, cfg *config.Config) (*collector.AggregatedResult, error) {
	res := &collector.AggregatedResult{
		Timestamp: time.Now(),
		Instance: collector.InstanceInfo{
			Host: cfg.GetHost(),
			Port: cfg.GetPort(),
		},
		Details: make(map[string]collector.CollectorResult),
		Summary: map[collector.HealthLevel]int{},
	}

	// 获取版本信息
	baseCollector := collector.NewBaseCollector()
	if version, err := baseCollector.GetVersion(ctx, db); err == nil {
		res.Instance.Version = version
	}

	// 单次采集器
	for _, c := range manager.GetAllSingleCollectors() {
		name := c.Name()
		if !cfg.IsCollectorEnabled("single", name) {
			logger.Debug("跳过单次采集器: %s", name)
			continue
		}
		logger.Debug("运行单次采集器: %s", name)
		cr, err := c.Collect(ctx, db)
		if err != nil {
			logger.Warn("采集器 %s 运行失败: %v", name, err)
			continue
		}
		res.Details[name] = cr
		res.Summary[cr.Status]++
	}

	// 周期采集器
	period := cfg.GetPeriod()
	if cfg.ShouldRunPeriodic() {
		for _, pc := range manager.GetAllPeriodicCollectors() {
			name := pc.Name()
			if !cfg.IsCollectorEnabled("periodic", name) {
				logger.Debug("跳过周期采集器: %s", name)
				continue
			}
			logger.Debug("周期采集器 %s 开始第一次快照", name)
			s1, err := pc.TakeSnapshot(ctx, db)
			if err != nil {
				logger.Warn("第一次快照失败: %v", err)
				continue
			}
			select {
			case <-time.After(period):
			case <-ctx.Done():
				return res, ctx.Err()
			}
			logger.Debug("周期采集器 %s 开始第二次快照", name)
			s2, err := pc.TakeSnapshot(ctx, db)
			if err != nil {
				logger.Warn("第二次快照失败: %v", err)
				continue
			}
			cr, err := pc.CalculateDelta(s1, s2)
			if err != nil {
				logger.Warn("增量计算失败: %v", err)
				continue
			}
			res.Details[name] = cr
			res.Summary[cr.Status]++
		}
	}

	// 计算总体状态
	res.Overall = collector.AggregateOverall(res.Summary)
	return res, nil
}
