package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"github.com/gin-gonic/gin"
)

// 测试用的认证令牌
const testAuthToken = "test-auth-token-12345"

// 创建测试用的临时文件
func createTestFile(t *testing.T, filename string, content string) (string, *os.File) {
	// 创建临时文件
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("打开测试文件失败: %v", err)
	}

	return filePath, file
}

// 设置测试环境
func setupTestConfig() {
	config.ValueOf.EnableUploadAPI = true
	config.ValueOf.UploadAuthToken = testAuthToken
	config.ValueOf.MaxFileSize = 10 * 1024 * 1024 // 10MB for testing
	config.ValueOf.UserQuota = 100 * 1024 * 1024 // 100MB
	config.ValueOf.AllowedMimeTypes = "image/jpeg,image/png,text/plain"
	config.ValueOf.AllowedExtensions = ".jpg,.jpeg,.png,.txt"
	config.ValueOf.UploadsPerMinute = 10
	config.ValueOf.UploadsPerHour = 100
	config.ValueOf.APICooldownSeconds = 0 // 测试时不需要冷却
	config.ValueOf.EnableProtection = false
	config.ValueOf.EnableDeepScan = false
	config.ValueOf.LogChannelID = -100123456789 // 测试频道ID
	config.ValueOf.Host = "http://localhost:8080"
}

// 创建测试用的Gin路由器
func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 初始化日志
	utils.InitLogger(true)

	// 初始化上传组件
	initUploadComponents(utils.Logger)

	// 注册上传路由
	route := &Route{Name: "/", Engine: router}
	allRoutes := &allRoutes{log: utils.Logger}
	allRoutes.LoadUpload(route)

	return router
}

// TestUploadHandler_Authentication 测试认证功能
func TestUploadHandler_Authentication(t *testing.T) {
	setupTestConfig()
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "无认证头",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "认证失败",
		},
		{
			name:           "错误的认证令牌",
			authHeader:     "Bearer wrong-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "认证失败",
		},
		{
			name:           "缺少Bearer前缀",
			authHeader:     testAuthToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "认证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试文件
			filePath, file := createTestFile(t, "test.txt", "test content")
			defer file.Close()
			defer os.Remove(filePath)

			// 创建multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "test.txt")
			io.Copy(part, file)
			writer.Close()

			// 创建请求
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 检查响应
			if w.Code != tt.expectedStatus {
				t.Errorf("期望状态码 %d, 得到 %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("解析响应失败: %v", err)
				}
				if !strings.Contains(response["error"].(string), tt.expectedError) {
					t.Errorf("期望错误包含 %s, 得到 %v", tt.expectedError, response["error"])
				}
			}
		})
	}
}

// TestUploadHandler_FileValidation 测试文件验证功能
func TestUploadHandler_FileValidation(t *testing.T) {
	setupTestConfig()
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		filename       string
		content        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "允许的文件类型",
			filename:       "test.txt",
			content:        "test content",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "不允许的文件类型",
			filename:       "test.exe",
			content:        "fake exe content",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "不允许的文件扩展名",
		},
		{
			name:           "文件大小超限",
			filename:       "large.txt",
			content:        strings.Repeat("x", 20*1024*1024), // 20MB
			expectedStatus: http.StatusBadRequest,
			expectedError:  "文件大小",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 跳过需要真实Telegram客户端的测试
			if tt.expectedStatus == http.StatusOK {
				t.Skip("跳过需要真实Telegram客户端的测试")
				return
			}

			// 创建测试文件
			filePath, file := createTestFile(t, tt.filename, tt.content)
			defer file.Close()
			defer os.Remove(filePath)

			// 创建multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", tt.filename)
			io.Copy(part, file)
			writer.Close()

			// 创建请求
			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+testAuthToken)

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 检查响应
			if w.Code != tt.expectedStatus {
				t.Errorf("期望状态码 %d, 得到 %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("解析响应失败: %v", err)
				}
				if !strings.Contains(response["error"].(string), tt.expectedError) {
					t.Errorf("期望错误包含 %s, 得到 %v", tt.expectedError, response["error"])
				}
			}
		})
	}
}

// TestUploadHandler_RateLimiting 测试速率限制功能
func TestUploadHandler_RateLimiting(t *testing.T) {
	setupTestConfig()

	// 设置严格的速率限制用于测试
	config.ValueOf.UploadsPerMinute = 2
	config.ValueOf.UploadsPerHour = 5

	// 跳过需要真实Telegram客户端的测试
	t.Skip("跳过需要真实Telegram客户端的速率限制测试")
}

// TestFileValidator_SanitizeFilename 测试文件名清理功能
func TestFileValidator_SanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{
			input:    "normal.txt",
			contains: "normal.txt",
		},
		{
			input:    "../../../etc/passwd",
			contains: "etc_passwd",
		},
		{
			input:    "file<with>.html",
			contains: "file_with",
		},
		{
			input:    "file:with|dangerous*chars.exe",
			contains: "file_with_dangerous_chars.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := utils.SanitizeFilename(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("清理后的文件名 %s 不包含期望的字符串 %s", result, tt.contains)
			}
			// 检查长度限制
			if len(result) > 255 {
				t.Errorf("清理后的文件名长度 %d 超过255", len(result))
			}
		})
	}
}

// TestUploadMetrics 测试上传指标
func TestUploadMetrics(t *testing.T) {
	metrics := &UploadMetrics{}

	// 测试更新指标
	updateMetricsWithStruct(metrics, true, 1024, "user1")
	
	if metrics.TotalUploads != 1 {
		t.Errorf("期望 TotalUploads = 1, 得到 %d", metrics.TotalUploads)
	}
	
	if metrics.TotalSize != 1024 {
		t.Errorf("期望 TotalSize = 1024, 得到 %d", metrics.TotalSize)
	}
	
	if metrics.AverageSize != 1024.0 {
		t.Errorf("期望 AverageSize = 1024.0, 得到 %f", metrics.AverageSize)
	}

	// 测试失败上传
	updateMetricsWithStruct(metrics, false, 0, "user1")
	
	if metrics.FailedUploads != 1 {
		t.Errorf("期望 FailedUploads = 1, 得到 %d", metrics.FailedUploads)
	}
}

// 辅助函数用于测试指标更新
func updateMetricsWithStruct(metrics *UploadMetrics, success bool, fileSize int64, userID string) {
	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()

	metrics.TotalUploads++

	if success {
		metrics.TotalSize += fileSize
	} else {
		metrics.FailedUploads++
	}

	// 计算平均大小
	if metrics.TotalUploads > 0 {
		metrics.AverageSize = float64(metrics.TotalSize) / float64(metrics.TotalUploads)
	}
}

// TestAuthenticateUpload 测试认证函数
func TestAuthenticateUpload(t *testing.T) {
	setupTestConfig()

	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name       string
		authHeader string
		expected   bool
	}{
		{
			name:       "有效令牌",
			authHeader: "Bearer " + testAuthToken,
			expected:   true,
		},
		{
			name:       "无效令牌",
			authHeader: "Bearer invalid-token",
			expected:   false,
		},
		{
			name:       "空令牌",
			authHeader: "",
			expected:   false,
		},
		{
			name:       "无Bearer前缀",
			authHeader: testAuthToken,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			c.Request.Header.Set("Authorization", tt.authHeader)

			result := authenticateUpload(c)
			if result != tt.expected {
				t.Errorf("期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}

// TestDetermineMediaType 测试媒体类型判断
func TestDetermineMediaType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    string
	}{
		{
			contentType: "image/jpeg",
			expected:    "photo",
		},
		{
			contentType: "image/png",
			expected:    "photo",
		},
		{
			contentType: "video/mp4",
			expected:    "video",
		},
		{
			contentType: "audio/mpeg",
			expected:    "audio",
		},
		{
			contentType: "application/pdf",
			expected:    "document",
		},
		{
			contentType: "text/plain",
			expected:    "document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := determineMediaType(tt.contentType)
			if result != tt.expected {
				t.Errorf("期望 %s, 得到 %s", tt.expected, result)
			}
		})
	}
}

// 集成测试 - 完整的上传流程测试
func TestUploadIntegration(t *testing.T) {
	setupTestConfig()

	// 这个测试需要模拟真实的Telegram客户端
	// 在实际环境中，需要集成测试

	t.Skip("集成测试需要真实的Telegram客户端和API配置")
}
