package routes

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/types"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// 全局上传组件
var (
	fileValidator    *utils.FileValidator
	rateLimiter     *utils.UploadRateLimiter
	quotaManager     *utils.QuotaManager
	uploadMetrics   *UploadMetrics
)

// 上传指标
type UploadMetrics struct {
	TotalUploads     int64 `json:"totalUploads"`
	TotalSize       int64 `json:"totalSize"`
	FailedUploads   int64 `json:"failedUploads"`
	BlockedUploads   int64 `json:"blockedUploads"`
	ActiveUsers     int    `json:"activeUsers"`
	AverageSize     float64 `json:"averageSize"`
	mutex           sync.Mutex
}

// 初始化上传组件
func initUploadComponents(log *zap.Logger) {
	// 初始化文件验证器
	fileValidator = utils.NewFileValidator(
		config.ValueOf.AllowedMimeTypes,
		config.ValueOf.AllowedExtensions,
		config.ValueOf.MaxFileSize,
		log.Named("FileValidator"),
	)

	// 初始化速率限制器
	rateLimiter = utils.NewUploadRateLimiter(
		config.ValueOf.UploadsPerMinute,
		config.ValueOf.UploadsPerHour,
	)

	// 初始化配额管理器
	quotaManager = utils.NewQuotaManager(
		config.ValueOf.UserQuota,
		log.Named("QuotaManager"),
	)

	// 初始化上传指标
	uploadMetrics = &UploadMetrics{}
}

// 注册上传路由
func (e *allRoutes) LoadUpload(r *Route) {
	log := e.log.Named("Upload")
	log.Info("正在加载上传路由...")

	// 如果未启用上传API，不注册路由
	if !config.ValueOf.EnableUploadAPI {
		log.Warn("上传API未启用，跳过路由注册")
		return
	}

	// 初始化上传组件
	initUploadComponents(log)

	// 注册路由
	r.Engine.POST("/upload", handleUpload)
	r.Engine.POST("/upload/batch", handleBatchUpload)
	r.Engine.GET("/upload/status", handleUploadStatus)
	r.Engine.GET("/upload/metrics", handleUploadMetrics)

	log.Info("上传路由加载完成")
}

// 单文件上传处理器
func handleUpload(ctx *gin.Context) {
	log := utils.Logger.Named("Upload")
	startTime := time.Now()

	// 1. 认证检查
	if !authenticateUpload(ctx) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "认证失败",
			"code":  401,
		})
		return
	}

	// 2. 获取用户标识
	userID := getUserIDFromAuth(ctx)
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无法识别用户",
			"code":  400,
		})
		return
	}

	// 3. 速率检查
	canUpload, waitTime := rateLimiter.CheckLimit(userID)
	if !canUpload {
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"error": fmt.Sprintf("请等待 %v 后再试", waitTime),
			"code":  429,
			"waitTime": waitTime.Seconds(),
		})
		return
	}

	// 4. 文件获取
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "文件获取失败: " + err.Error(),
			"code":  400,
		})
		updateMetrics(false, 0, userID)
		return
	}
	defer file.Close()

	// 5. 文件验证
	if err := validateUploadedFile(header); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  400,
		})
		updateMetrics(false, 0, userID)
		return
	}

	// 6. 用户配额检查
	canUseQuota, quotaErr := quotaManager.CheckQuota(parseUserID(userID), header.Size)
	if !canUseQuota {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": quotaErr.Error(),
			"code":  403,
		})
		updateMetrics(false, 0, userID)
		return
	}

	// 7. 执行上传
	result, err := uploadToTelegram(ctx, file, header)
	if err != nil {
		log.Error("上传到Telegram失败",
			zap.Error(err),
			zap.String("filename", header.Filename),
			zap.String("userID", userID))

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "上传失败: " + err.Error(),
			"code":  500,
		})
		updateMetrics(false, 0, userID)
		return
	}

	// 8. 更新配额使用量
	quotaManager.UpdateUsage(parseUserID(userID), header.Size)

	// 9. 返回成功结果
	uploadDuration := time.Since(startTime)
	updateMetrics(true, header.Size, userID)

	log.Info("文件上传成功",
		zap.String("filename", header.Filename),
		zap.Int64("size", header.Size),
		zap.String("userID", userID),
		zap.Duration("duration", uploadDuration))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "上传成功",
		"data": result,
		"uploadTime": uploadDuration.Seconds(),
	})
}

// 批量上传处理器
func handleBatchUpload(ctx *gin.Context) {
	log := utils.Logger.Named("BatchUpload")

	// 认证检查
	if !authenticateUpload(ctx) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "认证失败",
			"code":  401,
		})
		return
	}

	userID := getUserIDFromAuth(ctx)
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无法识别用户",
			"code":  400,
		})
		return
	}

	// 解析多文件
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "解析表单失败: " + err.Error(),
			"code":  400,
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "未找到文件",
			"code":  400,
		})
		return
	}

	// 检查批量限制
	if len(files) > 10 { // 最多同时上传10个文件
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "批量上传最多支持10个文件",
			"code":  400,
		})
		return
	}

	// 处理每个文件
	results := make([]gin.H, 0, len(files))
	successCount := 0
	totalSize := int64(0)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success": false,
				"error":  err.Error(),
			})
			continue
		}

		// 验证文件
		if err := validateUploadedFile(fileHeader); err != nil {
			file.Close()
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success": false,
				"error":  err.Error(),
			})
			continue
		}

		// 检查配额
		canUseQuota, quotaErr := quotaManager.CheckQuota(parseUserID(userID), fileHeader.Size)
		if !canUseQuota {
			file.Close()
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success": false,
				"error":  quotaErr.Error(),
			})
			continue
		}

		// 上传文件
		result, err := uploadToTelegram(ctx, file, fileHeader)
		file.Close()

		if err != nil {
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success": false,
				"error":  err.Error(),
			})
			updateMetrics(false, 0, userID)
		} else {
			results = append(results, gin.H{
				"filename": fileHeader.Filename,
				"success": true,
				"data":     result,
			})
			successCount++
			totalSize += fileHeader.Size
			quotaManager.UpdateUsage(parseUserID(userID), fileHeader.Size)
			updateMetrics(true, fileHeader.Size, userID)
		}
	}

	log.Info("批量上传完成",
		zap.String("userID", userID),
		zap.Int("totalFiles", len(files)),
		zap.Int("successCount", successCount),
		zap.Int64("totalSize", totalSize))

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "批量上传完成",
		"summary": gin.H{
			"totalFiles": len(files),
			"successCount": successCount,
			"failedCount": len(files) - successCount,
			"totalSize": totalSize,
		},
		"results": results,
	})
}

// 上传状态处理器
func handleUploadStatus(ctx *gin.Context) {
	if !authenticateUpload(ctx) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "认证失败"})
		return
	}

	userID := getUserIDFromAuth(ctx)
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无法识别用户"})
		return
	}

	// 获取用户配额使用情况
	usedQuota, err := utils.GetUserStorageUsage(parseUserID(userID), config.ValueOf.LogChannelID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取配额信息失败"})
		return
	}

	quotaPercent := float64(usedQuota) / float64(config.ValueOf.UserQuota) * 100

	ctx.JSON(http.StatusOK, gin.H{
		"userID": userID,
		"usedQuota": usedQuota,
		"maxQuota": config.ValueOf.UserQuota,
		"quotaPercent": quotaPercent,
		"remaining": config.ValueOf.UserQuota - usedQuota,
	})
}

// 上传指标处理器
func handleUploadMetrics(ctx *gin.Context) {
	uploadMetrics.mutex.Lock()
	defer uploadMetrics.mutex.Unlock()

	ctx.JSON(http.StatusOK, gin.H{
		"metrics": uploadMetrics,
		"timestamp": time.Now().Unix(),
	})
}

// 认证检查
func authenticateUpload(ctx *gin.Context) bool {
	// 检查是否配置了上传令牌
	if config.ValueOf.UploadAuthToken == "" {
		return false
	}

	// 检查Authorization头
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// 必须包含Bearer前缀
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return false
	}

	// 提取token
	token := authHeader[7:]

	// 验证token
	return token == config.ValueOf.UploadAuthToken
}

// 从认证信息获取用户ID
func getUserIDFromAuth(ctx *gin.Context) string {
	// 简化实现：使用Authorization头的一部分作为用户标识
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// 必须包含Bearer前缀
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return ""
	}

	// 提取token
	token := authHeader[7:]

	// 在实际实现中，这里应该解析JWT或其他用户标识
	// 暂时使用token作为用户ID
	return token
}

// 解析用户ID为int64
func parseUserID(userID string) int64 {
	// 简化实现，应该从token中解析真实的用户ID
	// 这里使用哈希值模拟
	hash := md5.Sum([]byte(userID))
	return int64(hash[0]) | int64(hash[1])<<8 | int64(hash[2])<<16 | int64(hash[3])<<24
}

// 验证上传的文件
func validateUploadedFile(header *multipart.FileHeader) error {
	// 读取文件前512字节进行验证
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// 使用文件验证器进行验证
	return fileValidator.ValidateFile(
		utils.SanitizeFilename(header.Filename),
		header.Size,
		header.Header.Get("Content-Type"),
		buffer[:n],
	)
}

// 上传文件到Telegram
func uploadToTelegram(ctx *gin.Context, file multipart.File, header *multipart.FileHeader) (*types.UploadResult, error) {
	// 获取可用的上传worker
	worker := bot.GetNextUploadWorker()
	if worker == nil {
		return nil, fmt.Errorf("没有可用的worker")
	}

	// 上传文件到Telegram - 使用FromReader方法
	sanitizedFilename := utils.SanitizeFilename(header.Filename)
	u := uploader.NewUploader(worker.Client.API())
	upload, err := u.FromReader(ctx, sanitizedFilename, file)
	if err != nil {
		return nil, fmt.Errorf("文件上传失败: %w", err)
	}

	// 确定媒体类型
	mediaType := determineMediaType(header.Header.Get("Content-Type"))

	// 构建媒体消息
	var media tg.InputMediaClass

	switch mediaType {
	case "photo":
		media = &tg.InputMediaUploadedPhoto{
			File: upload,
		}
	case "video":
		media = &tg.InputMediaUploadedDocument{
			File:     upload,
			MimeType: header.Header.Get("Content-Type"),
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: sanitizedFilename},
				&tg.DocumentAttributeVideo{
					Duration: 0, // 视频时长，可以后续扩展
					W:         0, // 视频宽度，可以后续扩展
					H:         0, // 视频高度，可以后续扩展
				},
			},
		}
	default: // document
		media = &tg.InputMediaUploadedDocument{
			File:     upload,
			MimeType: header.Header.Get("Content-Type"),
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: sanitizedFilename},
			},
		}
	}

	// 获取LOG_CHANNEL的InputPeer
	logChannelPeer, err := utils.GetLogChannelPeer(ctx, worker.Client.API(), worker.Client.PeerStorage)
	if err != nil {
		return nil, fmt.Errorf("获取日志频道失败: %w", err)
	}

	// 发送到LOG_CHANNEL
	req := &tg.MessagesSendMediaRequest{
		Peer:     &tg.InputPeerChannel{ChannelID: logChannelPeer.ChannelID, AccessHash: logChannelPeer.AccessHash},
		Media:    media,
		Message:  fmt.Sprintf("通过API上传: %s", sanitizedFilename),
		RandomID: time.Now().UnixNano(), // 必需的RandomID字段
	}

	update, err := worker.Client.API().MessagesSendMedia(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}

	// 解析结果获取消息ID和文件ID
	var messageID int
	var fileID int64
	switch result := update.(type) {
	case *tg.Updates:
		if len(result.Updates) > 0 {
			for _, upd := range result.Updates {
				if updateMsg, ok := upd.(*tg.UpdateMessageID); ok {
					messageID = updateMsg.ID
				}
				if updateNewMsg, ok := upd.(*tg.UpdateNewChannelMessage); ok {
					if msg, ok := updateNewMsg.Message.(*tg.Message); ok {
						messageID = msg.ID
						// 从消息中提取文件ID
						if msgMedia, ok := msg.Media.(*tg.MessageMediaDocument); ok {
							if doc, ok := msgMedia.Document.(*tg.Document); ok {
								fileID = doc.ID
							}
						} else if msgMedia, ok := msg.Media.(*tg.MessageMediaPhoto); ok {
							if photo, ok := msgMedia.Photo.(*tg.Photo); ok {
								fileID = photo.ID
							}
						}
					}
				}
			}
		}
	}

	if messageID == 0 {
		return nil, fmt.Errorf("无法获取消息ID")
	}

	// 如果没有获取到文件ID，使用消息ID作为备选
	if fileID == 0 {
		fileID = int64(messageID)
	}

	// 生成流媒体链接
	fullHash := utils.PackFile(
		sanitizedFilename,
		header.Size,
		header.Header.Get("Content-Type"),
		fileID,
	)
	hash := utils.GetShortHash(fullHash)

	// 返回结果
	return &types.UploadResult{
		Filename:    sanitizedFilename,
		Size:        header.Size,
		MimeType:    header.Header.Get("Content-Type"),
		MessageID:   messageID,
		StreamURL:    fmt.Sprintf("%s/stream/%d?hash=%s", config.ValueOf.Host, messageID, hash),
		DownloadURL:  fmt.Sprintf("%s/stream/%d?hash=%s&d=true", config.ValueOf.Host, messageID, hash),
		Hash:        hash,
		UploadTime:   time.Now(),
	}, nil
}

// 确定媒体类型
func determineMediaType(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		return "photo"
	} else if strings.HasPrefix(contentType, "video/") {
		return "video"
	} else if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	} else {
		return "document"
	}
}

// 更新上传指标
func updateMetrics(success bool, fileSize int64, userID string) {
	uploadMetrics.mutex.Lock()
	defer uploadMetrics.mutex.Unlock()

	uploadMetrics.TotalUploads++

	if success {
		uploadMetrics.TotalSize += fileSize
	} else {
		uploadMetrics.FailedUploads++
	}

	// 计算平均大小
	uploadMetrics.AverageSize = float64(uploadMetrics.TotalSize) / float64(uploadMetrics.TotalUploads)

	// 记录日志
	if success {
		utils.Logger.Info("上传指标更新",
			zap.String("userID", userID),
			zap.Int64("fileSize", fileSize),
			zap.Int64("totalUploads", uploadMetrics.TotalUploads),
			zap.Float64("averageSize", uploadMetrics.AverageSize))
	}
}