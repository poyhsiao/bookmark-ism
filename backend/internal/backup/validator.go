package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// BackupValidator 备份验证器
type BackupValidator struct {
	logger *zap.Logger
}

// NewBackupValidator 创建备份验证器
func NewBackupValidator(logger *zap.Logger) *BackupValidator {
	return &BackupValidator{
		logger: logger,
	}
}

// ValidateBackup 验证备份文件
func (bv *BackupValidator) ValidateBackup(filePath string) error {
	bv.logger.Debug("Validating backup file", zap.String("file", filePath))

	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("backup file does not exist: %w", err)
	}

	// 根据文件类型执行验证
	ext := filepath.Ext(filePath)
	switch {
	case strings.Contains(filePath, ".sql"):
		return bv.validateSQLBackup(filePath)
	case ext == ".gz" || ext == ".tar":
		return bv.validateArchiveBackup(filePath)
	case ext == ".enc":
		return bv.validateEncryptedBackup(filePath)
	default:
		bv.logger.Warn("Unknown backup file type, skipping validation",
			zap.String("file", filePath))
		return nil
	}
}

// validateSQLBackup 验证SQL备份文件
func (bv *BackupValidator) validateSQLBackup(filePath string) error {
	// 读取文件开头检查SQL格式
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open SQL backup: %w", err)
	}
	defer file.Close()

	// 读取前1024字节
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("failed to read SQL backup: %w", err)
	}

	content := string(buffer[:n])

	// 检查SQL备份的基本标识
	if !strings.Contains(content, "PostgreSQL") &&
	   !strings.Contains(content, "pg_dump") &&
	   !strings.Contains(content, "CREATE") &&
	   !strings.Contains(content, "INSERT") {
		return fmt.Errorf("invalid SQL backup format")
	}

	bv.logger.Debug("SQL backup validation passed", zap.String("file", filePath))
	return nil
}

// validateArchiveBackup 验证归档备份文件
func (bv *BackupValidator) validateArchiveBackup(filePath string) error {
	var cmd *exec.Cmd

	if strings.HasSuffix(filePath, ".gz") {
		// 测试gzip文件
		cmd = exec.Command("gzip", "-t", filePath)
	} else if strings.HasSuffix(filePath, ".tar") {
		// 测试tar文件
		cmd = exec.Command("tar", "-tf", filePath)
	} else {
		return fmt.Errorf("unsupported archive format")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("archive validation failed: %w", err)
	}

	bv.logger.Debug("Archive backup validation passed", zap.String("file", filePath))
	return nil
}

// validateEncryptedBackup 验证加密备份文件
func (bv *BackupValidator) validateEncryptedBackup(filePath string) error {
	// 检查文件大小（加密文件应该有一定大小）
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get encrypted file info: %w", err)
	}

	if fileInfo.Size() < 100 {
		return fmt.Errorf("encrypted backup file too small")
	}

	// 检查文件开头是否包含随机数据（简单检查）
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open encrypted backup: %w", err)
	}
	defer file.Close()

	// 读取前16字节检查是否为随机数据
	header := make([]byte, 16)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("failed to read encrypted backup header: %w", err)
	}

	// 简单检查：确保不是明文（没有常见的ASCII字符模式）
	asciiCount := 0
	for _, b := range header {
		if b >= 32 && b <= 126 {
			asciiCount++
		}
	}

	if asciiCount > 12 { // 如果超过75%是ASCII字符，可能不是加密数据
		bv.logger.Warn("Encrypted backup may not be properly encrypted",
			zap.String("file", filePath))
	}

	bv.logger.Debug("Encrypted backup validation passed", zap.String("file", filePath))
	return nil
}

// ValidateChecksum 验证文件校验和
func (bv *BackupValidator) ValidateChecksum(filePath, expectedChecksum string) error {
	cmd := exec.Command("sha256sum", filePath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := string(output[:64])
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s",
			expectedChecksum, actualChecksum)
	}

	bv.logger.Debug("Checksum validation passed",
		zap.String("file", filePath),
		zap.String("checksum", actualChecksum))

	return nil
}