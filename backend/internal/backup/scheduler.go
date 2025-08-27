package backup

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// BackupScheduler 备份调度器
type BackupScheduler struct {
	logger        *zap.Logger
	backupService *BackupService
	ticker        *time.Ticker
	stopChan      chan struct{}
}

// NewBackupScheduler 创建新的备份调度器
func NewBackupScheduler(logger *zap.Logger, backupService *BackupService) *BackupScheduler {
	return &BackupScheduler{
		logger:        logger,
		backupService: backupService,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动调度器
func (bs *BackupScheduler) Start(ctx context.Context) error {
	bs.logger.Info("Starting backup scheduler")

	// 每24小时执行一次完整备份
	bs.ticker = time.NewTicker(24 * time.Hour)

	go func() {
		for {
			select {
			case <-bs.ticker.C:
				bs.logger.Info("Executing scheduled full backup")
				if _, err := bs.backupService.CreateFullBackup(ctx); err != nil {
					bs.logger.Error("Scheduled backup failed", zap.Error(err))
				}
			case <-bs.stopChan:
				bs.logger.Info("Backup scheduler stopped")
				return
			case <-ctx.Done():
				bs.logger.Info("Backup scheduler stopped due to context cancellation")
				return
			}
		}
	}()

	return nil
}

// Stop 停止调度器
func (bs *BackupScheduler) Stop() {
	if bs.ticker != nil {
		bs.ticker.Stop()
	}
	close(bs.stopChan)
}