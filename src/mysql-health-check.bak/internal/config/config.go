package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type CollectorToggle struct {
	Enabled bool `json:"enabled"`
}

type CollectorConfig struct {
	Single   map[string]CollectorToggle `json:"single_collectors"`
	Periodic map[string]CollectorToggle `json:"periodic_collectors"`
}

type Config struct {
	// db 简化字段，调试模式下使用 debugDSN
	DebugMode bool `json:"debug_mode"`

	Host         string        `json:"host"`
	Port         int           `json:"port"`
	User         string        `json:"user"`
	Password     string        `json:"password"`
	Database     string        `json:"database"`
	LogLevel     string        `json:"log_level"`
	RunPeriodic  bool          `json:"run_periodic"`
	Period       time.Duration `json:"period"`
	QueryTimeout time.Duration `json:"query_timeout"`
	TotalTimeout time.Duration `json:"total_timeout"`
	Output       string        `json:"output_format"`

	Collectors CollectorConfig `json:"collectors"`

	// debug defaults
	debugDSN          string
	debugLogLevel     string
	debugRunPeriodic  bool
	debugPeriod       time.Duration
	debugQueryTimeout time.Duration
}

func NewConfig() *Config {
	c := &Config{
		DebugMode:    true,
		Host:         "172.17.139.15",
		Port:         16315,
		User:         "admin",
		Password:     "!QAZ2wsx",
		Database:     "",
		LogLevel:     "debug",
		RunPeriodic:  true,
		Period:       5 * time.Second,
		QueryTimeout: 5 * time.Second,
		TotalTimeout: 20 * time.Second,
		Output:       "text",
	}
	c.debugDSN = "root:root1234@tcp(127.0.0.1:3306)/?parseTime=true"
	c.debugLogLevel = "debug"
	c.debugRunPeriodic = true
	c.debugPeriod = 5 * time.Second
	c.debugQueryTimeout = 5 * time.Second

	c.setDefaultCollectors()
	return c
}

func (c *Config) setDefaultCollectors() {
	c.Collectors = CollectorConfig{
		Single: map[string]CollectorToggle{
			"base":        {Enabled: true},
			"connection":  {Enabled: true},
			"buffer_pool": {Enabled: true},
			"replication": {Enabled: true},
		},
		Periodic: map[string]CollectorToggle{
			"performance": {Enabled: true},
		},
	}
}

func (c *Config) LoadFromFile(path string) error {
	if c.DebugMode {
		fmt.Printf("调试模式：尝试加载配置文件 %s\n", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if c.DebugMode {
			fmt.Printf("读取配置失败，使用默认: %v\n", err)
		}
		return nil
	}
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("配置解析失败: %w", err)
	}
	return nil
}

func (c *Config) GetDSN() string {
	if c.DebugMode {
		return c.debugDSN
	}
	return ""
}

func (c *Config) GetHost() string     { return c.Host }
func (c *Config) GetPort() int        { return c.Port }
func (c *Config) GetUser() string     { return c.User }
func (c *Config) GetPassword() string { return c.Password }
func (c *Config) GetDatabase() string { return c.Database }
func (c *Config) GetLogLevel() string {
	if c.DebugMode {
		return c.debugLogLevel
	}
	return c.LogLevel
}
func (c *Config) ShouldRunPeriodic() bool {
	if c.DebugMode {
		return c.debugRunPeriodic
	}
	return c.RunPeriodic
}
func (c *Config) GetPeriod() time.Duration {
	if c.DebugMode {
		return c.debugPeriod
	}
	return c.Period
}
func (c *Config) GetQueryTimeout() time.Duration {
	if c.DebugMode {
		return c.debugQueryTimeout
	}
	return c.QueryTimeout
}
func (c *Config) GetTotalTimeout() time.Duration { return c.TotalTimeout }
func (c *Config) OutputFormat() string           { return c.Output }
func (c *Config) IsCollectorEnabled(typ, name string) bool {
	if typ == "single" {
		if v, ok := c.Collectors.Single[name]; ok {
			return v.Enabled
		}
		return true
	}
	if typ == "periodic" {
		if v, ok := c.Collectors.Periodic[name]; ok {
			return v.Enabled
		}
		return true
	}
	return false
}
