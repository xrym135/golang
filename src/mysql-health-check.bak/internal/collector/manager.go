package collector

import (
	"context"
	"database/sql"
	"laowang/mysql-health-check/pkg/core"
)

// Collector interfaces are defined here for easier import within this package.

type Collector interface {
	Name() string
	Collect(ctx context.Context, db *sql.DB) (core.CollectorResult, error)
}

type SnapshotCollector interface {
	Collector
	TakeSnapshot(ctx context.Context, db *sql.DB) (core.Snapshot, error)
	CalculateDelta(prev, curr core.Snapshot) (core.CollectorResult, error)
}

// CollectorManager 管理采集器注册与查询
type CollectorManager struct {
	singleCollectors   []Collector
	periodicCollectors []SnapshotCollector
}

func NewCollectorManager() *CollectorManager {
	return &CollectorManager{
		singleCollectors:   make([]Collector, 0),
		periodicCollectors: make([]SnapshotCollector, 0),
	}
}

func (m *CollectorManager) RegisterSingleCollector(c Collector) {
	m.singleCollectors = append(m.singleCollectors, c)
}

func (m *CollectorManager) RegisterPeriodicCollector(c SnapshotCollector) {
	m.periodicCollectors = append(m.periodicCollectors, c)
}

// GetAllSingleCollectors 返回注册的所有单次采集器
func (m *CollectorManager) GetAllSingleCollectors() []Collector {
	return m.singleCollectors
}

// GetAllPeriodicCollectors 返回注册的所有周期采集器
func (m *CollectorManager) GetAllPeriodicCollectors() []SnapshotCollector {
	return m.periodicCollectors
}
