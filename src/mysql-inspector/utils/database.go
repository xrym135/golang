package utils

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"mysql-inspector/types"
	"time"
)

// SafeDB 安全的数据库连接封装
type SafeDB struct {
	db     *sql.DB
	logger *Logger
	config types.MySQLConfig
}

// NewSafeDB 创建安全的数据库连接
func NewSafeDB(config types.MySQLConfig, logger *Logger) (*SafeDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&readTimeout=30s&writeTimeout=30s&parseTime=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Timeout,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 设置连接池参数，避免对线上造成压力
	db.SetMaxOpenConns(5)                   // 最大打开连接数
	db.SetMaxIdleConns(2)                   // 最大空闲连接数
	db.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期

	safeDB := &SafeDB{
		db:     db,
		logger: logger,
		config: config,
	}

	// 测试连接
	if err := safeDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	logger.Info("数据库连接成功: %s:%d", config.Host, config.Port)
	return safeDB, nil
}

// Ping 测试数据库连接
func (s *SafeDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Timeout)*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

// QueryWithTimeout 带超时的查询执行
func (s *SafeDB) QueryWithTimeout(query string, timeout time.Duration, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 验证SQL安全性（白名单机制）
	if !isSafeQuery(query) {
		return nil, fmt.Errorf("不安全的SQL查询: %s", query)
	}

	s.logger.Debug("执行SQL查询: %s", query)
	return s.db.QueryContext(ctx, query, args...)
}

// QueryRowWithTimeout 带超时的单行查询
func (s *SafeDB) QueryRowWithTimeout(query string, timeout time.Duration, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if !isSafeQuery(query) {
		s.logger.Error("尝试执行不安全SQL: %s", query)
		return nil
	}

	s.logger.Debug("执行SQL查询(单行): %s", query)
	return s.db.QueryRowContext(ctx, query, args...)
}

// ExecWithTimeout 带超时的执行操作
func (s *SafeDB) ExecWithTimeout(query string, timeout time.Duration, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if !isSafeQuery(query) {
		return nil, fmt.Errorf("不安全的SQL执行: %s", query)
	}

	s.logger.Debug("执行SQL: %s", query)
	return s.db.ExecContext(ctx, query, args...)
}

// Close 关闭数据库连接
func (s *SafeDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// isSafeQuery SQL白名单验证
func isSafeQuery(query string) bool {
	// 允许的只读查询列表
	safeQueries := []string{
		"SHOW GLOBAL STATUS",
		"SHOW GLOBAL VARIABLES",
		"SHOW PROCESSLIST",
		"SELECT * FROM information_schema.tables",
		"SELECT * FROM information_schema.innodb_metrics",
		"SELECT * FROM mysql.slow_log",
		"SELECT * FROM information_schema.PROCESSLIST",
		"SELECT @@version",
		"SELECT NOW()",
	}

	// 清理查询字符串
	cleanQuery := strings.TrimSpace(query)

	for _, safeQuery := range safeQueries {
		if strings.HasPrefix(strings.ToUpper(cleanQuery), strings.ToUpper(safeQuery)) {
			return true
		}
	}

	return false
}

// GetDB 获取原始数据库连接（谨慎使用）
func (s *SafeDB) GetDB() *sql.DB {
	return s.db
}
