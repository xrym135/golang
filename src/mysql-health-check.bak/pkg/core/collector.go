package core

import (
	"time"
)

// HealthLevel 健康等级
type HealthLevel int

const (
	HealthOK HealthLevel = iota
	HealthWarn
	HealthCritical
)

func (h HealthLevel) String() string {
	switch h {
	case HealthOK:
		return "OK"
	case HealthWarn:
		return "WARN"
	case HealthCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Snapshot 代表一次快照
type Snapshot struct {
	Name   string
	Values map[string]float64
	Time   time.Time
}

// CollectorResult 单次或增量采集结果
type CollectorResult struct {
	Name    string
	Metrics map[string]float64
	Status  HealthLevel
	Message string
}

// InstanceInfo 实例信息
type InstanceInfo struct {
	Host    string
	Port    int
	Version string
}
