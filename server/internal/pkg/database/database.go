// Package database 提供 Eino Career Agent 的数据库初始化功能
// 使用 GORM + SQLite 驱动，自动迁移所有领域模型
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
)

// Init 初始化数据库连接并执行自动迁移
// dbPath 为 SQLite 数据库文件路径
// 返回 GORM 数据库实例或错误
func Init(dbPath string) (*gorm.DB, error) {
	// 确保数据库文件所在目录存在
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	// 配置 GORM 日志级别
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	// 连接 SQLite 数据库
	// 使用 _journal_mode=WAL 提升并发读写性能
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接 SQLite 数据库失败: %w", err)
	}

	// 启用 SQLite 的外键约束（默认不启用）
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 SQL DB 失败: %w", err)
	}
	sqlDB.Exec("PRAGMA foreign_keys = ON")

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// 自动迁移所有领域模型
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库自动迁移失败: %w", err)
	}

	return db, nil
}

// autoMigrate 执行所有领域模型的自动迁移
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Conversation{},
		&domain.Message{},
		&domain.Report{},
		&domain.UploadedFile{},
	)
}
