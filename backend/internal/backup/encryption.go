package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// EncryptionService 加密服务
type EncryptionService struct {
	logger *zap.Logger
	key    []byte
}

// NewEncryptionService 创建加密服务
func NewEncryptionService(logger *zap.Logger) *EncryptionService {
	// 在实际应用中，密钥应该从安全的配置源获取
	password := "bookmark-backup-encryption-key"
	key := sha256.Sum256([]byte(password))

	return &EncryptionService{
		logger: logger,
		key:    key[:],
	}
}

// EncryptFile 加密文件
func (es *EncryptionService) EncryptFile(filePath string) (string, error) {
	es.logger.Debug("Encrypting backup file", zap.String("file", filePath))

	// 读取原文件
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 生成加密文件名
	encryptedPath := filePath + ".enc"

	// 写入加密文件
	if err := os.WriteFile(encryptedPath, ciphertext, 0644); err != nil {
		return "", fmt.Errorf("failed to write encrypted file: %w", err)
	}

	es.logger.Debug("File encrypted successfully",
		zap.String("original", filePath),
		zap.String("encrypted", encryptedPath))

	return encryptedPath, nil
}

// DecryptFile 解密文件
func (es *EncryptionService) DecryptFile(encryptedPath string) (string, error) {
	es.logger.Debug("Decrypting backup file", zap.String("file", encryptedPath))

	// 读取加密文件
	ciphertext, err := os.ReadFile(encryptedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(es.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 创建GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 提取nonce和密文
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	// 生成解密文件名
	decryptedPath := filepath.Dir(encryptedPath) + "/decrypted_" + filepath.Base(encryptedPath)
	if filepath.Ext(encryptedPath) == ".enc" {
		decryptedPath = encryptedPath[:len(encryptedPath)-4]
	}

	// 写入解密文件
	if err := os.WriteFile(decryptedPath, plaintext, 0644); err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}

	es.logger.Debug("File decrypted successfully",
		zap.String("encrypted", encryptedPath),
		zap.String("decrypted", decryptedPath))

	return decryptedPath, nil
}