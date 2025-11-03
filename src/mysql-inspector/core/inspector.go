package core

import (
	"database/sql"
	"mysql-inspector/types"
	"mysql-inspector/utils"
	"sync"
	"time"
)

// Inspector 巡检器
type Inspector struct {
	db               *sql.DB
	snapshotManager  *SnapshotManager
	pluginManager    *PluginManager
	logger           *utils.Logger
	stopChan         chan bool
	running          bool
	snapshotTicker   *time.Ticker
	inspectionTicker *time.Ticker
	mutex            sync.Mutex
}

// NewInspector 创建巡检器
func NewInspector(db *sql.DB, snapshotManager *SnapshotManager, pluginManager *PluginManager, logger *utils.Logger) *Inspector {
	return &Inspector{
		db:              db,
		snapshotManager: snapshotManager,
		pluginManager:   pluginManager,
		logger:          logger,
		stopChan:        make(chan bool),
		running:         false,
	}
}

// Start 启动巡检（分离快照任务和巡检任务）
func (i *Inspector) Start(snapshotInterval, inspectionInterval int) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.running {
		i.logger.Warning("巡检器已经在运行")
		return
	}

	i.running = true

	// 创建两个独立的定时器
	i.snapshotTicker = time.NewTicker(time.Duration(snapshotInterval) * time.Second)
	i.inspectionTicker = time.NewTicker(time.Duration(inspectionInterval) * time.Second)

	i.logger.Info("启动巡检器 - 快照间隔: %d秒, 巡检间隔: %d秒", snapshotInterval, inspectionInterval)

	// 立即执行一次快照和巡检
	i.takeSnapshot()
	i.executeInspection()

	// 启动协程处理定时任务
	go i.run()
}

// run 运行主循环
func (i *Inspector) run() {
	for {
		select {
		case <-i.snapshotTicker.C:
			// 快照任务：只获取数据，不执行插件
			i.takeSnapshot()

		case <-i.inspectionTicker.C:
			// 巡检任务：执行所有插件，使用已有的快照数据
			i.executeInspection()

		case <-i.stopChan:
			i.logger.Info("停止巡检器")
			return
		}
	}
}

// takeSnapshot 执行快照任务（只获取数据）
func (i *Inspector) takeSnapshot() {
	i.logger.Debug("执行快照任务...")
	_, err := i.snapshotManager.TakeSnapshot()
	if err != nil {
		i.logger.Error("获取快照失败: %v", err)
	}
}

// executeInspection 执行巡检任务（使用已有快照执行插件）
func (i *Inspector) executeInspection() {
	i.logger.Info("开始执行巡检任务...")

	// 获取最新快照
	currentSnapshot := i.snapshotManager.GetLatestSnapshot()
	if currentSnapshot == nil {
		i.logger.Error("没有可用的快照数据")
		return
	}

	// 获取上一个快照（用于计算变化率）
	previousSnapshot := i.snapshotManager.GetPreviousSnapshot()

	// 执行所有插件
	plugins := i.pluginManager.GetPlugins()
	for _, plugin := range plugins {
		if i.shouldExecutePlugin(plugin) {
			i.logger.Debug("执行插件: %s", plugin.Name())
			result, err := plugin.Execute(i.db, currentSnapshot, previousSnapshot)
			if err != nil {
				i.logger.Error("插件 %s 执行失败: %v", plugin.Name(), err)
				continue
			}
			// 记录结果
			i.logResult(result)
		}
	}

	i.logger.Info("巡检任务执行完成")
}

// shouldExecutePlugin 检查插件是否应该执行
func (i *Inspector) shouldExecutePlugin(plugin Plugin) bool {
	now := time.Now()
	lastRun := plugin.GetLastRunTime()

	// 如果是第一次运行，或者超过了插件配置的间隔时间
	if lastRun.IsZero() || now.Sub(lastRun) >= time.Duration(plugin.GetInterval())*time.Second {
		plugin.SetLastRunTime(now)
		return true
	}

	return false
}

// logResult 记录巡检结果
func (i *Inspector) logResult(result *types.InspectionResult) {
	switch result.Level {
	case "ERROR":
		i.logger.Error("[%s] %s", result.PluginName, result.Message)
	case "WARNING":
		i.logger.Warning("[%s] %s", result.PluginName, result.Message)
	default:
		i.logger.Info("[%s] %s", result.PluginName, result.Message)
	}

	// 记录详细指标（DEBUG级别）
	if len(result.Metrics) > 0 {
		i.logger.Debug("[%s] 指标: %+v", result.PluginName, result.Metrics)
	}
}

// Stop 停止巡检
func (i *Inspector) Stop() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.running {
		i.stopChan <- true
		i.running = false

		if i.snapshotTicker != nil {
			i.snapshotTicker.Stop()
		}
		if i.inspectionTicker != nil {
			i.inspectionTicker.Stop()
		}

		i.logger.Info("巡检器已停止")
	}
}

// IsRunning 检查是否在运行
func (i *Inspector) IsRunning() bool {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return i.running
}
