package core

import (
	"context"
	"database/sql"
	"fmt"
	"mysql-inspector/types"
	"mysql-inspector/utils"
	"time"
)

// SnapshotManager 快照管理器
type SnapshotManager struct {
	snapshots    []*types.SnapshotData
	maxSnapshots int
	db           *sql.DB
	logger       *utils.Logger
}

// NewSnapshotManager 创建快照管理器
func NewSnapshotManager(db *sql.DB, maxSnapshots int, logger *utils.Logger) *SnapshotManager {
	return &SnapshotManager{
		snapshots:    make([]*types.SnapshotData, 0),
		maxSnapshots: maxSnapshots,
		db:           db,
		logger:       logger,
	}
}

// TakeSnapshot 获取当前快照（带超时控制）
func (sm *SnapshotManager) TakeSnapshot() (*types.SnapshotData, error) {
	sm.logger.Debug("开始采集快照数据...")

	snapshot := &types.SnapshotData{
		Timestamp:   time.Now(),
		GlobalStats: make(map[string]interface{}),
		TableStats:  make(map[string]interface{}),
		ProcessList: make([]map[string]interface{}, 0),
		CustomData:  make(map[string]interface{}),
	}

	// 设置全局超时上下文（最多30秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取全局状态（带超时）
	if err := sm.collectGlobalStatus(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("采集全局状态失败: %v", err)
	}

	// 获取表统计信息（带超时）
	if err := sm.collectTableStats(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("采集表统计信息失败: %v", err)
	}

	// 获取进程列表（带超时）
	if err := sm.collectProcessList(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("采集进程列表失败: %v", err)
	}

	// 添加到快照列表
	sm.snapshots = append(sm.snapshots, snapshot)

	// 维护快照数量
	if len(sm.snapshots) > sm.maxSnapshots {
		sm.snapshots = sm.snapshots[1:]
	}

	sm.logger.Debug("快照采集完成，当前快照数量: %d", len(sm.snapshots))
	return snapshot, nil
}

// GetLatestSnapshot 获取最新快照
func (sm *SnapshotManager) GetLatestSnapshot() *types.SnapshotData {
	if len(sm.snapshots) == 0 {
		return nil
	}
	return sm.snapshots[len(sm.snapshots)-1]
}

// GetPreviousSnapshot 获取前一个快照
func (sm *SnapshotManager) GetPreviousSnapshot() *types.SnapshotData {
	if len(sm.snapshots) < 2 {
		return nil
	}
	return sm.snapshots[len(sm.snapshots)-2]
}

// GetSnapshots 获取所有快照
func (sm *SnapshotManager) GetSnapshots() []*types.SnapshotData {
	return sm.snapshots
}

// collectGlobalStatus 收集全局状态（带超时）
func (sm *SnapshotManager) collectGlobalStatus(ctx context.Context, snapshot *types.SnapshotData) error {
	// 设置查询超时（15秒）
	queryCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	rows, err := sm.db.QueryContext(queryCtx, "SHOW GLOBAL STATUS")
	if err != nil {
		return fmt.Errorf("查询全局状态超时或失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var variableName, value string
		if err := rows.Scan(&variableName, &value); err != nil {
			sm.logger.Warning("解析全局状态行失败: %v", err)
			continue
		}
		snapshot.GlobalStats[variableName] = value
	}

	return rows.Err()
}

// collectTableStats 收集表统计信息（带超时）
func (sm *SnapshotManager) collectTableStats(ctx context.Context, snapshot *types.SnapshotData) error {
	// 设置查询超时（20秒）
	queryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	query := `
        SELECT 
            TABLE_SCHEMA,
            TABLE_NAME,
            ENGINE,
            TABLE_ROWS,
            DATA_LENGTH,
            INDEX_LENGTH,
            DATA_FREE,
            CREATE_TIME,
            UPDATE_TIME
        FROM information_schema.tables 
        WHERE TABLE_SCHEMA NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
    `

	rows, err := sm.db.QueryContext(queryCtx, query)
	if err != nil {
		return fmt.Errorf("查询表统计信息超时或失败: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tableSchema, tableName, engine    string
			tableRows                         sql.NullInt64
			dataLength, indexLength, dataFree sql.NullInt64
			createTime, updateTime            sql.NullTime
		)

		if err := rows.Scan(
			&tableSchema, &tableName, &engine,
			&tableRows, &dataLength, &indexLength, &dataFree,
			&createTime, &updateTime,
		); err != nil {
			sm.logger.Warning("解析表统计信息行失败: %v", err)
			continue
		}

		key := fmt.Sprintf("%s.%s", tableSchema, tableName)
		snapshot.TableStats[key] = map[string]interface{}{
			"engine":       engine,
			"table_rows":   tableRows.Int64,
			"data_length":  dataLength.Int64,
			"index_length": indexLength.Int64,
			"data_free":    dataFree.Int64,
			"create_time":  createTime.Time,
			"update_time":  updateTime.Time,
		}
	}

	return rows.Err()
}

// collectProcessList 收集进程列表（带超时）
func (sm *SnapshotManager) collectProcessList(ctx context.Context, snapshot *types.SnapshotData) error {
	// 设置查询超时（10秒）
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := sm.db.QueryContext(queryCtx, "SHOW PROCESSLIST")
	if err != nil {
		return fmt.Errorf("查询进程列表超时或失败: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			sm.logger.Warning("解析进程列表行失败: %v", err)
			continue
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			rowData[col] = values[i]
		}
		snapshot.ProcessList = append(snapshot.ProcessList, rowData)
	}

	return rows.Err()
}
