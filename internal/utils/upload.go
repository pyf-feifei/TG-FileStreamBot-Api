package utils

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"EverythingSuckz/fsb/config"
	"github.com/celestix/gotgproto"
	"go.uber.org/zap"
)

// 速率限制器
type UploadRateLimiter struct {
	userUploads map[string][]time.Time
	mutex       sync.RWMutex
	maxPerMinute int
	maxPerHour   int
}

// 创建新的速率限制器
func NewUploadRateLimiter(maxPerMinute, maxPerHour int) *UploadRateLimiter {
	return &UploadRateLimiter{
		userUploads: make(map[string][]time.Time),
		maxPerMinute: maxPerMinute,
		maxPerHour:   maxPerHour,
	}
}

// 检查用户是否超出速率限制
func (url *UploadRateLimiter) CheckLimit(userID string) (bool, time.Duration) {
	url.mutex.Lock()
	defer url.mutex.Unlock()

	now := time.Now()
	url.cleanupOldUploads(now)

	uploads := url.userUploads[userID]

	// 检查每分钟限制
	recentCount := 0
	for _, uploadTime := range uploads {
		if now.Sub(uploadTime) < time.Minute {
			recentCount++
		}
	}

	if recentCount >= url.maxPerMinute {
		oldestRecent := time.Now()
		for _, uploadTime := range uploads {
			if now.Sub(uploadTime) < time.Minute && uploadTime.Before(oldestRecent) {
				oldestRecent = uploadTime
			}
		}
		waitTime := time.Minute - now.Sub(oldestRecent)
		return false, waitTime
	}

	// 检查每小时限制
	hourAgo := now.Add(-time.Hour)
	hourCount := 0
	for _, uploadTime := range uploads {
		if uploadTime.After(hourAgo) {
			hourCount++
		}
	}

	if hourCount >= url.maxPerHour {
		oldestHour := time.Now()
		for _, uploadTime := range uploads {
			if uploadTime.After(hourAgo) && uploadTime.Before(oldestHour) {
				oldestHour = uploadTime
			}
		}
		waitTime := time.Hour - now.Sub(oldestHour)
		return false, waitTime
	}

	// 记录此次上传
	url.userUploads[userID] = append(uploads, now)
	return true, 0
}

// 清理过期的上传记录
func (url *UploadRateLimiter) cleanupOldUploads(now time.Time) {
	for userID, uploads := range url.userUploads {
		var validUploads []time.Time
		for _, uploadTime := range uploads {
			if now.Sub(uploadTime) < time.Hour {
				validUploads = append(validUploads, uploadTime)
			}
		}
		if len(validUploads) == 0 {
			delete(url.userUploads, userID)
		} else {
			url.userUploads[userID] = validUploads
		}
	}
}

// 文件安全验证器
type FileValidator struct {
	allowedMimeTypes  []string
	allowedExtensions []string
	maxFileSize      int64
	logger           *zap.Logger
}

// 创建文件验证器
func NewFileValidator(mimeTypes, extensions string, maxSize int64, logger *zap.Logger) *FileValidator {
	return &FileValidator{
		allowedMimeTypes:  strings.Split(mimeTypes, ","),
		allowedExtensions: strings.Split(extensions, ","),
		maxFileSize:      maxSize,
		logger:           logger,
	}
}

// 验证文件
func (fv *FileValidator) ValidateFile(filename string, fileSize int64, contentType string, fileData []byte) error {
	// 1. 文件大小检查
	if fileSize > fv.maxFileSize {
		return fmt.Errorf("文件大小 %d 超过最大限制 %d", fileSize, fv.maxFileSize)
	}

	// 2. 文件扩展名检查
	ext := strings.ToLower(filepath.Ext(filename))
	if !fv.isAllowedExtension(ext) {
		return fmt.Errorf("不允许的文件扩展名: %s", ext)
	}

	// 3. MIME类型检查
	if !fv.isAllowedMimeType(contentType) {
		return fmt.Errorf("不允许的MIME类型: %s", contentType)
	}

	// 4. 文件头验证（深度扫描）
	if len(fileData) >= 512 {
		detectedType := http.DetectContentType(fileData)
		if detectedType != contentType && detectedType != "application/octet-stream" {
			fv.logger.Warn("文件类型不匹配",
				zap.String("declared", contentType),
				zap.String("detected", detectedType),
				zap.String("filename", filename))

			// 如果启用了深度扫描，拒绝不匹配的文件
			if config.ValueOf.EnableDeepScan {
				return fmt.Errorf("文件类型与内容不符，可能存在安全风险")
			}
		}
	}

	return nil
}

// 检查是否为允许的扩展名
func (fv *FileValidator) isAllowedExtension(ext string) bool {
	for _, allowed := range fv.allowedExtensions {
		if strings.TrimSpace(allowed) == ext {
			return true
		}
	}
	return false
}

// 检查是否为允许的MIME类型
func (fv *FileValidator) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range fv.allowedMimeTypes {
		if strings.TrimSpace(allowed) == mimeType {
			return true
		}
	}
	return false
}

// 文件名清理
func SanitizeFilename(filename string) string {
	// 移除路径分隔符
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// 移除特殊字符
	dangerous := []string{"<", ">", ":", "\"", "|", "?", "*", "&", "%", "#"}
	for _, char := range dangerous {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// 限制长度
	if len(filename) > 255 {
		name := strings.TrimSuffix(filename, filepath.Ext(filename))
		ext := filepath.Ext(filename)
		maxNameLen := 255 - len(ext)
		if maxNameLen > 0 {
			filename = name[:maxNameLen] + ext
		} else {
			filename = ext
		}
	}

	return filename
}

// 计算用户当前存储使用量
func GetUserStorageUsage(userID int64, logChannelID int64, client *gotgproto.Client) (int64, error) {
	// 这里需要通过Telegram API获取用户在日志频道中的消息
	// 简化实现，返回当前用户的存储配额使用情况
	// 在实际实现中，需要调用Telegram API查询历史消息并统计

	// 暂时返回0，表示未实现
	// TODO: 实现真实的存储使用量查询
	return 0, nil
}

// 生成文件MD5哈希用于去重检查
func CalculateFileMD5(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// 用户配额检查器
type QuotaManager struct {
	quotas map[int64]int64 // userID -> used quota
	mutex   sync.RWMutex
	logger  *zap.Logger
	maxQuota int64
}

// 创建配额管理器
func NewQuotaManager(maxQuota int64, logger *zap.Logger) *QuotaManager {
	return &QuotaManager{
		quotas:   make(map[int64]int64),
		logger:    logger,
		maxQuota: maxQuota,
	}
}

// 检查用户配额
func (qm *QuotaManager) CheckQuota(userID int64, fileSize int64) (bool, error) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	used := qm.quotas[userID]
	if used+fileSize > qm.maxQuota {
		return false, fmt.Errorf("超出存储配额: 已使用 %d MB，限制 %d MB",
			used/(1024*1024), qm.maxQuota/(1024*1024))
	}

	return true, nil
}

// 更新用户配额使用量
func (qm *QuotaManager) UpdateUsage(userID int64, fileSize int64) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	qm.quotas[userID] += fileSize
	qm.logger.Info("更新用户配额使用量",
		zap.Int64("userID", userID),
		zap.Int64("used", qm.quotas[userID]),
		zap.Int64("maxQuota", qm.maxQuota))
}