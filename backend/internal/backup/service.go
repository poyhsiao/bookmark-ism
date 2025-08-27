package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BackupService 备份服务
type BackupService struct {
	logger       *zap.Logger
	db           *gorm.DB
	config       *BackupConfig
	scheduler    *BackupScheduler
	encryption   *EncryptionService
	validator    *BackupValidator
	monitoring   *BackupMonitoring
}

// BackupConfig 备份配置
type BackupConfig struct {
	// 数据库备份配置
	DatabaseHost     string `yaml:"database_host"`
	DatabasePort     int    `yaml:"database_port"`
	DatabaseUser     string `yaml:"database_user"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseName     string `yaml:"database_name"`

	// 存储配置
	BackupDir         string `yaml:"backup_dir"`
	RemoteBackupDir   string `yaml:"remote_backup_dir"`
	CrossRegionBackup bool   `yaml:"cross_region_backup"`

	// 备份策略
	FullBackupSchedule        string        `yaml:"full_backup_schedule"`        // cron表达式
	IncrementalBackupSchedule string        `yaml:"incremental_backup_schedule"` // cron表达式
	RetentionDays             int           `yaml:"retention_days"`
	CompressionEnabled        bool          `yaml:"compression_enabled"`
	EncryptionEnabled         bool          `yaml:"encryption_enabled"`

	// 性能配置
	MaxConcurrentBackups int           `yaml:"max_concurrent_backups"`
	BackupTimeout        time.Duration `yaml:"backup_timeout"`
}

// BackupType 备份类型
type BackupType string

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
	BackupTypeStorage     BackupType = "storage"
)

// BackupJob 备份任务
type BackupJob struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Type        BackupType `json:"type"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	ErrorMessage string   `json:"error_message,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewBackupService 创建备份服务
func NewBackupService(logger *zap.Logger, db *gorm.DB, config *BackupConfig) *BackupService {
	bs := &BackupService{
		logger:     logger,
		db:         db,
		config:     config,
		encryption: NewEncryptionService(logger),
		validator:  NewBackupValidator(logger),
		monitoring: NewBackupMonitoring(logger),
	}

	bs.scheduler = NewBackupScheduler(logger, bs)
	return bs
}

// Start 启动备份服务
func (bs *BackupService) Start(ctx context.Context) error {
	bs.logger.Info("Starting backup service")

	// 确保备份目录存在
	if err := os.MkdirAll(bs.config.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 启动调度器
	if err := bs.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start backup scheduler: %w", err)
	}

	// 启动监控
	bs.monitoring.Start(ctx)

	return nil
}

// CreateFullBackup 创建完整备份
func (bs *BackupService) CreateFullBackup(ctx context.Context) (*BackupJob, error) {
	job := &BackupJob{
		Type:      BackupTypeFull,
		Status:    "running",
		StartTime: time.Now(),
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("bookmark_full_backup_%s.sql", timestamp)
	filePath := filepath.Join(bs.config.BackupDir, fileName)
	job.FilePath = filePath

	bs.logger.Info("Starting full database backup", zap.String("file", filePath))

	// 执行pg_dump
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", bs.config.DatabaseHost,
		"-p", fmt.Sprintf("%d", bs.config.DatabasePort),
		"-U", bs.config.DatabaseUser,
		"-d", bs.config.DatabaseName,
		"-f", filePath,
		"--verbose",
		"--clean",
		"--if-exists",
	)

	// 设置环境变量
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", bs.config.DatabasePassword))

	// 执行命令
	if err := cmd.Run(); err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		endTime := time.Now()
		job.EndTime = &endTime

		bs.logger.Error("Full backup failed", zap.Error(err))
		return job, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		return job, err
	}
	job.FileSize = fileInfo.Size()

	// 压缩备份文件
	if bs.config.CompressionEnabled {
		compressedPath, err := bs.compressFile(filePath)
		if err != nil {
			bs.logger.Error("Failed to compress backup", zap.Error(err))
		} else {
			// 删除原文件，使用压缩文件
			os.Remove(filePath)
			job.FilePath = compressedPath

			// 更新文件大小
			if compFileInfo, err := os.Stat(compressedPath); err == nil {
				job.FileSize = compFileInfo.Size()
			}
		}
	}

	// 加密备份文件
	if bs.config.EncryptionEnabled {
		encryptedPath, err := bs.encryption.EncryptFile(job.FilePath)
		if err != nil {
			bs.logger.Error("Failed to encrypt backup", zap.Error(err))
		} else {
			// 删除原文件，使用加密文件
			os.Remove(job.FilePath)
			job.FilePath = encryptedPath

			// 更新文件大小
			if encFileInfo, err := os.Stat(encryptedPath); err == nil {
				job.FileSize = encFileInfo.Size()
			}
		}
	}

	// 计算校验和
	checksum, err := bs.calculateChecksum(job.FilePath)
	if err != nil {
		bs.logger.Error("Failed to calculate checksum", zap.Error(err))
	} else {
		job.Checksum = checksum
	}

	// 验证备份
	if err := bs.validator.ValidateBackup(job.FilePath); err != nil {
		job.Status = "validation_failed"
		job.ErrorMessage = err.Error()
		bs.logger.Error("Backup validation failed", zap.Error(err))
	} else {
		job.Status = "completed"
	}

	endTime := time.Now()
	job.EndTime = &endTime

	// 保存任务记录
	if err := bs.db.Create(job).Error; err != nil {
		bs.logger.Error("Failed to save backup job", zap.Error(err))
	}

	bs.logger.Info("Full backup completed",
		zap.String("file", job.FilePath),
		zap.Int64("size", job.FileSize),
		zap.String("status", job.Status))

	return job, nil
}

// CreateIncrementalBackup 创建增量备份
func (bs *BackupService) CreateIncrementalBackup(ctx context.Context) (*BackupJob, error) {
	// 查找最近的完整备份
	var lastFullBackup BackupJob
	if err := bs.db.Where("type = ? AND status = ?", BackupTypeFull, "completed").
		Order("created_at DESC").First(&lastFullBackup).Error; err != nil {
		return nil, fmt.Errorf("no full backup found for incremental backup: %w", err)
	}

	job := &BackupJob{
		Type:      BackupTypeIncremental,
		Status:    "running",
		StartTime: time.Now(),
	}

	// 生成增量备份文件名
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("bookmark_incremental_backup_%s.sql", timestamp)
	filePath := filepath.Join(bs.config.BackupDir, fileName)
	job.FilePath = filePath

	bs.logger.Info("Starting incremental database backup",
		zap.String("file", filePath),
		zap.Time("since", lastFullBackup.CreatedAt))

	// 执行增量备份（简化实现，实际应使用WAL或其他增量机制）
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", bs.config.DatabaseHost,
		"-p", fmt.Sprintf("%d", bs.config.DatabasePort),
		"-U", bs.config.DatabaseUser,
		"-d", bs.config.DatabaseName,
		"-f", filePath,
		"--verbose",
		"--data-only",
		"--inserts",
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", bs.config.DatabasePassword))

	if err := cmd.Run(); err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		endTime := time.Now()
		job.EndTime = &endTime

		bs.logger.Error("Incremental backup failed", zap.Error(err))
		return job, err
	}

	// 处理压缩和加密（与完整备份相同的逻辑）
	if err := bs.postProcessBackup(job); err != nil {
		bs.logger.Error("Failed to post-process incremental backup", zap.Error(err))
	}

	endTime := time.Now()
	job.EndTime = &endTime
	job.Status = "completed"

	// 保存任务记录
	if err := bs.db.Create(job).Error; err != nil {
		bs.logger.Error("Failed to save backup job", zap.Error(err))
	}

	bs.logger.Info("Incremental backup completed",
		zap.String("file", job.FilePath),
		zap.Int64("size", job.FileSize))

	return job, nil
}

// CreateStorageBackup 创建存储备份
func (bs *BackupService) CreateStorageBackup(ctx context.Context) (*BackupJob, error) {
	job := &BackupJob{
		Type:      BackupTypeStorage,
		Status:    "running",
		StartTime: time.Now(),
	}

	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("bookmark_storage_backup_%s.tar.gz", timestamp)
	filePath := filepath.Join(bs.config.BackupDir, fileName)
	job.FilePath = filePath

	bs.logger.Info("Starting storage backup", zap.String("file", filePath))

	// 这里应该实现MinIO存储的备份逻辑
	// 简化实现：假设存储文件在本地目录
	storageDir := "/tmp/minio-data" // 实际应从配置获取

	cmd := exec.CommandContext(ctx, "tar", "-czf", filePath, "-C", storageDir, ".")
	if err := cmd.Run(); err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		endTime := time.Now()
		job.EndTime = &endTime

		bs.logger.Error("Storage backup failed", zap.Error(err))
		return job, err
	}

	if err := bs.postProcessBackup(job); err != nil {
		bs.logger.Error("Failed to post-process storage backup", zap.Error(err))
	}

	endTime := time.Now()
	job.EndTime = &endTime
	job.Status = "completed"

	// 保存任务记录
	if err := bs.db.Create(job).Error; err != nil {
		bs.logger.Error("Failed to save backup job", zap.Error(err))
	}

	bs.logger.Info("Storage backup completed",
		zap.String("file", job.FilePath),
		zap.Int64("size", job.FileSize))

	return job, nil
}

// postProcessBackup 后处理备份文件
func (bs *BackupService) postProcessBackup(job *BackupJob) error {
	// 获取文件信息
	fileInfo, err := os.Stat(job.FilePath)
	if err != nil {
		return err
	}
	job.FileSize = fileInfo.Size()

	// 加密
	if bs.config.EncryptionEnabled {
		encryptedPath, err := bs.encryption.EncryptFile(job.FilePath)
		if err != nil {
			return err
		}
		os.Remove(job.FilePath)
		job.FilePath = encryptedPath

		if encFileInfo, err := os.Stat(encryptedPath); err == nil {
			job.FileSize = encFileInfo.Size()
		}
	}

	// 计算校验和
	checksum, err := bs.calculateChecksum(job.FilePath)
	if err != nil {
		return err
	}
	job.Checksum = checksum

	return nil
}

// compressFile 压缩文件
func (bs *BackupService) compressFile(filePath string) (string, error) {
	compressedPath := filePath + ".gz"
	cmd := exec.Command("gzip", filePath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return compressedPath, nil
}

// calculateChecksum 计算文件校验和
func (bs *BackupService) calculateChecksum(filePath string) (string, error) {
	cmd := exec.Command("sha256sum", filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// 返回校验和的第一部分
	return string(output[:64]), nil
}

// CleanupOldBackups 清理旧备份
func (bs *BackupService) CleanupOldBackups(ctx context.Context) error {
	cutoffTime := time.Now().AddDate(0, 0, -bs.config.RetentionDays)

	var oldBackups []BackupJob
	if err := bs.db.Where("created_at < ?", cutoffTime).Find(&oldBackups).Error; err != nil {
		return err
	}

	for _, backup := range oldBackups {
		// 删除备份文件
		if err := os.Remove(backup.FilePath); err != nil {
			bs.logger.Error("Failed to delete backup file",
				zap.String("file", backup.FilePath),
				zap.Error(err))
		}

		// 删除数据库记录
		if err := bs.db.Delete(&backup).Error; err != nil {
			bs.logger.Error("Failed to delete backup record",
				zap.Uint("id", backup.ID),
				zap.Error(err))
		}
	}

	bs.logger.Info("Cleaned up old backups",
		zap.Int("count", len(oldBackups)),
		zap.Time("cutoff", cutoffTime))

	return nil
}

// ListBackups 列出备份
func (bs *BackupService) ListBackups(limit int) ([]BackupJob, error) {
	var backups []BackupJob
	query := bs.db.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&backups).Error; err != nil {
		return nil, err
	}

	return backups, nil
}