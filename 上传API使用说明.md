# TG-FileStreamBot 上传API使用说明

## 概述

TG-FileStreamBot 现在支持HTTP API文件上传功能，允许通过RESTful API接口上传文件到Telegram，并自动生成流媒体下载链接。

## API端点

### 1. 单文件上传
```http
POST /upload
Content-Type: multipart/form-data
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

#### 请求参数
- `file`: 上传的文件（必需）
- 其他字段将被忽略

#### 响应格式
```json
{
  "success": true,
  "message": "上传成功",
  "data": {
    "filename": "document.pdf",
    "size": 1048576,
    "mimeType": "application/pdf",
    "messageId": 12345,
    "streamUrl": "http://your-domain.com/stream/12345?hash=abc123",
    "downloadUrl": "http://your-domain.com/stream/12345?hash=abc123&d=true",
    "hash": "abc123",
    "uploadTime": "2024-01-01T12:00:00Z"
  },
  "uploadTime": 1.234
}
```

### 2. 批量文件上传
```http
POST /upload/batch
Content-Type: multipart/form-data
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

#### 请求参数
- `files`: 多个上传文件（最多10个）

#### 响应格式
```json
{
  "success": true,
  "message": "批量上传完成",
  "summary": {
    "totalFiles": 3,
    "successCount": 2,
    "failedCount": 1,
    "totalSize": 2097152
  },
  "results": [
    {
      "filename": "file1.jpg",
      "success": true,
      "data": { /* 同单文件上传的data格式 */ }
    },
    {
      "filename": "file2.exe",
      "success": false,
      "error": "不允许的文件扩展名"
    }
  ]
}
```

### 3. 上传状态查询
```http
GET /upload/status
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

#### 响应格式
```json
{
  "userId": "user123",
  "usedQuota": 104857600,
  "maxQuota": 10737418240,
  "quotaPercent": 0.98,
  "remaining": 18857640
}
```

### 4. 上传指标查询
```http
GET /upload/metrics
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

#### 响应格式
```json
{
  "metrics": {
    "totalUploads": 1500,
    "totalSize": 1073741824,
    "failedUploads": 12,
    "blockedUploads": 8,
    "activeUsers": 25,
    "averageSize": 715827.88
  },
  "timestamp": 1704067200
}
```

## 使用示例

### cURL 示例

#### 单文件上传
```bash
curl -X POST http://your-domain.com/upload \
  -H "Authorization: Bearer your-secret-upload-token-here" \
  -F "file=@/path/to/your/file.pdf"
```

#### 批量上传
```bash
curl -X POST http://your-domain.com/upload/batch \
  -H "Authorization: Bearer your-secret-upload-token-here" \
  -F "files=@/path/to/file1.jpg" \
  -F "files=@/path/to/file2.mp4" \
  -F "files=@/path/to/file3.txt"
```

#### 查询状态
```bash
curl -X GET http://your-domain.com/upload/status \
  -H "Authorization: Bearer your-secret-upload-token-here"
```

### Python 示例

```python
import requests
import json

# 配置
BASE_URL = "http://your-domain.com"
AUTH_TOKEN = "your-secret-upload-token-here"

def upload_file(file_path):
    """上传单个文件"""
    url = f"{BASE_URL}/upload"
    headers = {
        "Authorization": f"Bearer {AUTH_TOKEN}"
    }

    with open(file_path, 'rb') as f:
        files = {'file': f}
        response = requests.post(url, files=files, headers=headers)

    return response.json()

def batch_upload(file_paths):
    """批量上传文件"""
    url = f"{BASE_URL}/upload/batch"
    headers = {
        "Authorization": f"Bearer {AUTH_TOKEN}"
    }

    files = []
    for file_path in file_paths:
        files.append(('files', open(file_path, 'rb')))

    response = requests.post(url, files=files, headers=headers)
    return response.json()

def get_upload_status():
    """获取上传状态"""
    url = f"{BASE_URL}/upload/status"
    headers = {
        "Authorization": f"Bearer {AUTH_TOKEN}"
    }

    response = requests.get(url, headers=headers)
    return response.json()

# 使用示例
if __name__ == "__main__":
    # 单文件上传
    result = upload_file("/path/to/document.pdf")
    print(json.dumps(result, indent=2))

    # 批量上传
    files = ["/path/to/file1.jpg", "/path/to/file2.mp4"]
    result = batch_upload(files)
    print(json.dumps(result, indent=2))

    # 查询状态
    status = get_upload_status()
    print(json.dumps(status, indent=2))
```

### JavaScript 示例

```javascript
class TGFileStreamBot {
    constructor(baseUrl, authToken) {
        this.baseUrl = baseUrl;
        this.authToken = authToken;
    }

    async uploadFile(file) {
        const formData = new FormData();
        formData.append('file', file);

        const response = await fetch(`${this.baseUrl}/upload`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.authToken}`
            },
            body: formData
        });

        return await response.json();
    }

    async batchUpload(files) {
        const formData = new FormData();
        for (let i = 0; i < files.length; i++) {
            formData.append('files', files[i]);
        }

        const response = await fetch(`${this.baseUrl}/upload/batch`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.authToken}`
            },
            body: formData
        });

        return await response.json();
    }

    async getUploadStatus() {
        const response = await fetch(`${this.baseUrl}/upload/status`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${this.authToken}`
            }
        });

        return await response.json();
    }
}

// 使用示例
const bot = new TGFileStreamBot('http://your-domain.com', 'your-secret-upload-token-here');

// 单文件上传
const fileInput = document.getElementById('fileInput');
fileInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (file) {
        try {
            const result = await bot.uploadFile(file);
            console.log('上传成功:', result);
        } catch (error) {
            console.error('上传失败:', error);
        }
    }
});
```

## 配置说明

### 环境变量配置

在 `fsb.env` 文件中配置以下参数：

```bash
# 启用上传API
ENABLE_UPLOAD_API=true

# 设置认证令牌
UPLOAD_AUTH_TOKEN=your-secret-upload-token-here

# 文件限制
MAX_FILE_SIZE=2147483648                    # 2GB
USER_QUOTA=10737418240                      # 10GB

# 允许的文件类型
ALLOWED_MIME_TYPES=image/jpeg,image/png,video/mp4,application/pdf,text/plain
ALLOWED_EXTENSIONS=.jpg,.png,.mp4,.pdf,.txt

# 速率限制
UPLOADS_PER_MINUTE=5
UPLOADS_PER_HOUR=50
CONCURRENT_UPLOADS_PER_USER=3
API_COOLDOWN_SECONDS=1

# 安全设置
ENABLE_PROTECTION_MODE=true
ENABLE_DEEP_SCAN=false
```

### 命令行参数

也可以通过命令行参数配置：

```bash
# 启用上传API
./fsb run --enable-upload-api

# 设置认证令牌
./fsb run --upload-auth-token your-secret-token

# 设置文件大小限制（字节）
./fsb run --max-file-size 2147483648

# 设置用户配额（字节）
./fsb run --user-quota 10737418240

# 设置允许的MIME类型
./fsb run --allowed-mime-types "image/jpeg,image/png,video/mp4"

# 设置允许的扩展名
./fsb run --allowed-extensions ".jpg,.png,.mp4"

# 设置速率限制
./fsb run --uploads-per-minute 5 --uploads-per-hour 50

# 设置并发限制
./fsb run --concurrent-uploads 3

# 设置API冷却时间
./fsb run --api-cooldown-seconds 1

# 启用保护模式
./fsb run --enable-protection

# 启用深度扫描
./fsb run --enable-deep-scan
```

## 安全机制

### 1. 认证
所有上传请求必须包含有效的 `Authorization: Bearer YOUR_TOKEN` 头。

### 2. 文件验证
- **类型检查**: 验证MIME类型和文件扩展名
- **大小检查**: 限制单个文件最大2GB
- **内容验证**: 检查文件头真实性（可选深度扫描）
- **配额检查**: 限制用户总存储使用量

### 3. 速率控制
- **每分钟限制**: 默认5个文件/分钟/用户
- **每小时限制**: 默认50个文件/小时/用户
- **并发限制**: 默认3个同时上传/用户
- **API冷却**: 默认1秒冷却时间

### 4. 自动保护
检测到异常行为时自动启用保护模式：
- 降低文件大小限制
- 提高验证严格度
- 限制上传频率
- 发送管理员警报

## 错误处理

### 常见错误码

| 错误码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 认证失败 | 未提供或认证令牌无效 |
| 400 | 文件获取失败 | multipart表单解析错误 |
| 400 | 不允许的文件类型 | 文件类型或扩展名被禁止 |
| 400 | 文件大小超过最大限制 | 文件超过配置的最大大小 |
| 400 | 文件类型与内容不符 | 文件内容与声明的类型不匹配 |
| 403 | 超出存储配额 | 用户已用完配额 |
| 429 | 请等待 X 秒后再试 | 超出速率限制，需要等待 |
| 500 | 上传失败 | 服务器内部错误或Telegram API错误 |

## 最佳实践

### 1. 错误重试
```python
import time
import requests

def upload_with_retry(file_path, max_retries=3):
    for attempt in range(max_retries):
        try:
            result = upload_file(file_path)
            if result.get('success'):
                return result
            else:
                error = result.get('error', '')
                if '请等待' in error:
                    # 解析等待时间
                    import re
                    match = re.search(r'(\d+)', error)
                    if match:
                        wait_time = int(match.group(1))
                        time.sleep(wait_time)
                        continue
        except requests.exceptions.RequestException as e:
            if attempt < max_retries - 1:
                time.sleep(2 ** attempt)  # 指数退避
                continue
            else:
                return {"success": False, "error": f"网络错误: {str(e)}"}

    return {"success": False, "error": "上传失败：超过最大重试次数"}
```

### 2. 进度监控
```javascript
async function uploadWithProgress(file, onProgress) {
    const formData = new FormData();
    formData.append('file', file);

    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                const percentComplete = (e.loaded / e.total) * 100;
                onProgress(percentComplete, e.loaded, e.total);
            }
        });

        xhr.addEventListener('load', () => {
            resolve(JSON.parse(xhr.responseText));
        });

        xhr.addEventListener('error', () => {
            reject(new Error('上传失败'));
        });

        xhr.open('POST', `${this.baseUrl}/upload`);
        xhr.setRequestHeader('Authorization', `Bearer ${this.authToken}`);
        xhr.send(formData);
    });
}

// 使用示例
uploadWithProgress(file, (percent, loaded, total) => {
    console.log(`上传进度: ${percent.toFixed(2)}% (${loaded}/${total} bytes)`);
});
```

## 注意事项

1. **文件存储**: 所有文件都存储在配置的Telegram日志频道中，不会占用服务器磁盘空间
2. **流量消耗**: 上传和下载都会消耗服务器带宽流量
3. **API限制**: 受Telegram Bot API速率限制影响，建议合理配置多个worker
4. **安全考虑**: 启用深度扫描可以提高安全性但会影响性能
5. **配额管理**: 合理设置用户配额避免恶意占用存储空间

## 集成示例

### Web应用集成
```javascript
// 简单的文件上传组件
class FileUploader extends HTMLElement {
    constructor() {
        this.innerHTML = `
            <div class="upload-area">
                <input type="file" id="fileInput" multiple accept="image/*,video/*,.pdf,.txt" />
                <button onclick="this.uploadFiles()">上传</button>
                <div id="progress"></div>
                <div id="results"></div>
            </div>
        `;
    }

    async uploadFiles() {
        const fileInput = this.querySelector('#fileInput');
        const files = Array.from(fileInput.files);
        const progress = this.querySelector('#progress');
        const results = this.querySelector('#results');

        if (files.length === 0) return;

        progress.innerHTML = '上传中... 0%';
        results.innerHTML = '';

        try {
            const endpoint = files.length === 1 ? '/upload' : '/upload/batch';
            const response = await this.uploadFiles(files, endpoint);

            progress.innerHTML = '上传完成！';

            if (response.success) {
                this.displayResults(response.results || [response.data]);
            } else {
                results.innerHTML = `<div class="error">上传失败: ${response.error}</div>`;
            }
        } catch (error) {
            progress.innerHTML = '';
            results.innerHTML = `<div class="error">上传错误: ${error.message}</div>`;
        }
    }

    async uploadFiles(files, endpoint) {
        const formData = new FormData();
        files.forEach(file => formData.append('files', file));

        const response = await fetch(`/api${endpoint}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('uploadToken')}`
            },
            body: formData
        });

        return await response.json();
    }

    displayResults(results) {
        const results = this.querySelector('#results');
        const html = results.map(result => {
            if (result.success) {
                return `
                    <div class="result success">
                        <h4>${result.data.filename}</h4>
                        <p>大小: ${(result.data.size / 1024 / 1024).toFixed(2)} MB</p>
                        <a href="${result.data.streamUrl}" target="_blank">播放</a>
                        <a href="${result.data.downloadUrl}" target="_blank">下载</a>
                    </div>
                `;
            } else {
                return `
                    <div class="result error">
                        <h4>${result.filename}</h4>
                        <p class="error-msg">错误: ${result.error}</p>
                    </div>
                `;
            }
        }).join('');

        results.innerHTML = html;
    }
}

customElements.define('file-uploader', FileUploader);
```

这个上传API为TG-FileStreamBot提供了完整的HTTP文件上传能力，同时保持了原有的安全性和流媒体特性。