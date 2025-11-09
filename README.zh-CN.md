<div align="center">

# TG-FileStreamBot-Api

<a herf="https://github.com/EverythingSuckz/TG-FileStreamBot-Api">
    <img src="https://telegra.ph/file/a8bb3f6b334ad1200ddb4.png" height="100" width="100" alt="File Stream Bot Logo">
</a>

### 一个为您的 Telegram 文件**生成直链**的 Telegram 机器人，支持 **HTTP 上传 API**。

[![English](https://img.shields.io/badge/Language-English-blue?style=for-the-badge)](README.md)
[![中文](https://img.shields.io/badge/语言-简体中文-red?style=for-the-badge)](README.zh-CN.md)

</div>

<hr>

> [!NOTE]
> 如果您对 Python 版本感兴趣，请查看 [python 分支](https://github.com/EverythingSuckz/TG-FileStreamBot-Api/tree/python)。

<hr>

## 目录

<details open="open">
  <summary>目录导航</summary>
  <ol>
    <li>
      <a href="#如何搭建">如何搭建</a>
      <ul>
        <li><a href="#部署到-koyeb">部署到 Koyeb</a></li>
        <li><a href="#部署到-heroku">部署到 Heroku</a></li>
      </ul>
      <ul>
        <li><a href="#下载并运行">下载并运行</a></li>
        <li><a href="#使用-docker-compose-运行">使用 Docker Compose 运行</a></li>
        <li><a href="#使用-docker-运行">使用 Docker 运行</a></li>
        <li><a href="#从源码构建">从源码构建</a>
          <ul>
            <li><a href="#ubuntu">Ubuntu</a></li>
            <li><a href="#windows">Windows</a></li>
          </ul>
        </li>
      </ul>
    </li>
    <li>
      <a href="#配置说明">配置说明</a>
      <ul>
        <li><a href="#必需环境变量">必需环境变量</a></li>
        <li><a href="#可选环境变量">可选环境变量</a></li>
        <li><a href="#使用多个-bot-加速">使用多个 Bot 加速</a></li>
        <li><a href="#使用用户会话自动添加-bot">使用用户会话自动添加 Bot</a>
          <ul>
            <li><a href="#这个功能是做什么的">这个功能是做什么的？</a></li>
            <li><a href="#如何生成会话字符串">如何生成会话字符串？</a></li>
          </ul>
        </li>
      </ul>
    </li>
    <li>
      <a href="#http-上传-api">HTTP 上传 API</a>
      <ul>
        <li><a href="#api-端点">API 端点</a></li>
        <li><a href="#api-配置">API 配置</a></li>
        <li><a href="#api-使用示例">使用示例</a></li>
        <li><a href="#api-安全特性">安全特性</a></li>
      </ul>
    </li>
    <li><a href="#贡献">贡献</a></li>
    <li><a href="#联系我">联系我</a></li>
    <li><a href="#致谢">致谢</a></li>
  </ol>
</details>

## 如何搭建

### 部署到 Koyeb

> [!IMPORTANT]
> 您需要展开"环境变量和文件"部分，并在点击部署按钮之前更新环境变量。

> [!NOTE]
> 这将部署**最新的 Docker 发布版本而非最新提交**。由于使用预构建的 Docker 容器，部署速度会明显更快。

[![部署到 Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=docker&name=file-stream-bot&image=ghcr.io/everythingsuckz/fsb:latest&env%5BAPI_HASH%5D=&env%5BAPI_ID%5D=&env%5BAPI_HASH%5D=&env%5BAPI_ID%5D=&env%5BBOT_TOKEN%5D=&env%5BHOST%5D=https%3A%2F%2F%7B%7B+KOYEB_PUBLIC_DOMAIN+%7D%7D&env%5BLOG_CHANNEL%5D=&env%5BPORT%5D=8038&ports=8038%3Bhttp%3B%2F&hc_protocol%5B8038%5D=tcp&hc_grace_period%5B8038%5D=5&hc_interval%5B8038%5D=30&hc_restart_limit%5B8038%5D=3&hc_timeout%5B8038%5D=5&hc_path%5B8038%5D=%2F&hc_method%5B8038%5D=get)

### 部署到 Heroku

> [!NOTE]
> 您需要[分叉](https://github.com/EverythingSuckz/TG-FileStreamBot-Api/fork)此仓库才能部署到 Heroku。

按下下面的按钮快速部署到 Heroku

[![部署到 Heroku](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

[点击这里](https://devcenter.heroku.com/articles/config-vars#using-the-heroku-dashboard)了解如何在 Heroku 中添加/编辑[环境变量](#必需环境变量)。

<hr>

### 下载并运行

- 前往[发布页面](https://github.com/EverythingSuckz/TG-FileStreamBot-Api/releases)，从*预发布*部分下载适合您平台和架构的版本。
- 将 zip 文件解压到一个文件夹。
- 创建一个名为 `fsb.env` 的文件并添加所有变量（参考 `fsb.sample.env` 文件）。
- 使用命令 `chmod +x fsb` 给可执行文件执行权限（Windows 不需要）。
- 使用 `./fsb run` 命令运行机器人（Windows 使用 `./fsb.exe run`）。

<hr>

### 使用 Docker Compose 运行

- 克隆仓库
```sh
git clone https://github.com/EverythingSuckz/TG-FileStreamBot-Api
cd TG-FileStreamBot-Api
```

- 创建一个名为 `fsb.env` 的文件并添加所有变量（参考 `fsb.sample.env` 文件）。

```sh
nano fsb.env
```

- 构建并运行 docker-compose 文件

```sh
docker-compose up -d
```
或

```sh
docker compose up -d
```

<hr>

### 使用 Docker 运行

```sh
docker run --env-file fsb.env ghcr.io/everythingsuckz/fsb:latest
```
其中 `fsb.env` 是包含所有变量的环境文件。

<hr>

### 从源码构建

#### Ubuntu

> [!NOTE]
> 确保安装 Go 1.21 或更高版本。
> 参考 https://stackoverflow.com/a/17566846/15807350

```sh
git clone https://github.com/EverythingSuckz/TG-FileStreamBot-Api
cd TG-FileStreamBot-Api
go build ./cmd/fsb/
chmod +x fsb
mv fsb.sample.env fsb.env
nano fsb.env
# (添加您的环境变量，详见下一节)
./fsb run
```

停止程序请按 <kbd>CTRL</kbd>+<kbd>C</kbd>

#### Windows

> [!NOTE]
> 确保安装 Go 1.21 或更高版本。

```powershell
git clone https://github.com/EverythingSuckz/TG-FileStreamBot-Api
cd TG-FileStreamBot-Api
go build ./cmd/fsb/
Rename-Item -LiteralPath ".\fsb.sample.env" -NewName ".\fsb.env"
notepad fsb.env
# (添加您的环境变量，详见下一节)
.\fsb run
```

停止程序请按 <kbd>CTRL</kbd>+<kbd>C</kbd>

## 配置说明

如果您在本地托管，请在根目录创建一个名为 `fsb.env` 的文件并添加所有变量。
您可以查看 `fsb.sample.env`。
`fsb.env` 文件示例：

```sh
API_ID=452525
API_HASH=esx576f8738x883f3sfzx83
BOT_TOKEN=55838383:yourbottokenhere
LOG_CHANNEL=-10045145224562
PORT=8080
HOST=http://yourserverip
# (如果要设置多个 bot)
MULTI_TOKEN1=55838373:yourworkerbottokenhere
MULTI_TOKEN2=55838355:yourworkerbottokenhere
```

### 必需环境变量

在运行机器人之前，您需要设置以下必需变量：

- `API_ID`：您的 Telegram 账户的 API ID，可从 my.telegram.org 获取。

- `API_HASH`：您的 Telegram 账户的 API Hash，也可从 my.telegram.org 获取。

- `BOT_TOKEN`：Telegram 媒体流机器人的 bot token，可从 [@BotFather](https://telegram.dog/BotFather) 获取。

- `LOG_CHANNEL`：日志频道的频道 ID，机器人将在此转发媒体消息并存储这些文件以使生成的直链正常工作。要获取频道 ID，请创建一个新的 telegram 频道（公开或私有），在频道中发布一些内容，将消息转发给 [@missrose_bot](https://telegram.dog/MissRose_bot) 并**回复转发的消息**使用 /id 命令。复制转发的频道 ID 并粘贴到此字段中。

### 可选环境变量

除了必需变量外，您还可以设置以下可选变量：

- `PORT`：设置您的 Web 应用程序将监听的端口。默认值为 8080。

- `HOST`：完全限定域名（如果有）或使用您的服务器 IP。（例如：`https://example.com` 或 `http://14.1.154.2:8080`）

- `HASH_LENGTH`：生成的 URL 的自定义哈希长度。哈希长度必须大于 5 且小于或等于 32。默认值为 6。

- `USE_SESSION_FILE`：为工作客户端使用会话文件。这会加快工作 bot 的启动速度。（默认：`false`）

- `USER_SESSION`：用户 bot 的 pyrogram 会话字符串。用于自动将 bot 添加到 `LOG_CHANNEL`。（默认：`null`）

- `ALLOWED_USERS`：用逗号（`,`）分隔的用户 ID 列表。如果设置了此项，只有此列表中的用户才能使用机器人。（默认：`null`）

<hr>

### 使用多个 Bot 加速

> [!NOTE]
> **什么是多客户端功能，它有什么作用？** <br>
> 此功能在工作 bot 之间共享 Telegram API 请求，以在许多用户使用服务器时加快下载速度，并避免 Telegram 设置的洪水限制。<br>

> [!NOTE]
> 您可以添加最多 50 个 bot，因为 50 是您可以在 Telegram 频道中设置的最大 bot 管理员数量。

要启用多客户端，生成新的 bot token 并将其添加到您的 `fsb.env` 中，使用以下键名。

`MULTI_TOKEN1`：在此添加您的第一个 bot token。

`MULTI_TOKEN2`：在此添加您的第二个 bot token。

您也可以添加任意数量的 bot。（最大限制为 50）
`MULTI_TOKEN3`、`MULTI_TOKEN4` 等。

> [!WARNING]
> 不要忘记将所有这些工作 bot 添加到 `LOG_CHANNEL` 以确保正常运行

### 使用用户会话自动添加 Bot

> [!WARNING]
> 这有时可能导致您的账户被限制或封禁。
> **只有新创建的账户容易出现这种情况。**

要使用此功能，您需要为用户账户生成一个 pyrogram 会话字符串，并将其添加到 `fsb.env` 文件中的 `USER_SESSION` 变量。

#### 这个功能是做什么的？

此功能用于在工作 bot 启动时自动将它们添加到 `LOG_CHANNEL`。当您有很多工作 bot 而不想手动将它们添加到 `LOG_CHANNEL` 时，这很有用。

#### 如何生成会话字符串？

最简单的生成会话字符串的方法是运行

```sh
./fsb session --api-id <your api id> --api-hash <your api hash>
```

<img src="https://github.com/EverythingSuckz/TG-FileStreamBot-Api/assets/65120517/b5bd2b88-0e1f-4dbc-ad9a-faa6d5a17320" height=300>

<br><br>

这将使用二维码认证为您的用户账户生成会话字符串。目前还不支持通过手机号码认证，将在未来添加。

## HTTP 上传 API

TG-FileStreamBot-Api 现在支持 HTTP 文件上传功能，允许您通过 RESTful API 上传文件并自动生成流媒体下载链接。

### API 端点

#### 1. 单文件上传
```http
POST /upload
Content-Type: multipart/form-data
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

**请求参数：**
- `file`：要上传的文件（必需）

**响应格式：**
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
  }
}
```

#### 2. 批量文件上传
```http
POST /upload/batch
Content-Type: multipart/form-data
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

**请求参数：**
- `files`：要上传的多个文件（最多 10 个）

**响应格式：**
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
      "data": { /* 与单文件上传相同 */ }
    },
    {
      "filename": "file2.exe",
      "success": false,
      "error": "不允许的文件扩展名"
    }
  ]
}
```

#### 3. 上传状态查询
```http
GET /upload/status
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

**响应格式：**
```json
{
  "userId": "user123",
  "usedQuota": 104857600,
  "maxQuota": 10737418240,
  "quotaPercent": 0.98,
  "remaining": 18857640
}
```

#### 4. 上传指标查询
```http
GET /upload/metrics
Authorization: Bearer YOUR_UPLOAD_TOKEN
```

**响应格式：**
```json
{
  "metrics": {
    "totalUploads": 1500,
    "totalSize": 1073741824,
    "failedUploads": 12,
    "blockedUploads": 8,
    "activeUsers": 25,
    "averageSize": 715827.88
  }
}
```

### API 配置

在您的 `fsb.env` 文件中添加以下配置：

```bash
# 启用上传 API
ENABLE_UPLOAD_API=true

# 认证令牌
UPLOAD_AUTH_TOKEN=your-secret-upload-token-here

# 文件限制
MAX_FILE_SIZE=2147483648                    # 2GB
USER_QUOTA=10737418240                      # 每个用户 10GB

# 允许的文件类型
ALLOWED_MIME_TYPES=image/jpeg,image/png,video/mp4,application/pdf,text/plain
ALLOWED_EXTENSIONS=.jpg,.jpeg,.png,.mp4,.pdf,.txt

# 速率限制
UPLOADS_PER_MINUTE=5
UPLOADS_PER_HOUR=50
CONCURRENT_UPLOADS_PER_USER=3
API_COOLDOWN_SECONDS=1

# 安全设置
ENABLE_PROTECTION_MODE=true
ENABLE_DEEP_SCAN=false
```

### API 使用示例

#### cURL 示例

**单文件上传：**
```bash
curl -X POST http://your-domain.com/upload \
  -H "Authorization: Bearer your-secret-upload-token-here" \
  -F "file=@/path/to/your/file.pdf"
```

**批量上传：**
```bash
curl -X POST http://your-domain.com/upload/batch \
  -H "Authorization: Bearer your-secret-upload-token-here" \
  -F "files=@/path/to/file1.jpg" \
  -F "files=@/path/to/file2.mp4" \
  -F "files=@/path/to/file3.txt"
```

**查询状态：**
```bash
curl -X GET http://your-domain.com/upload/status \
  -H "Authorization: Bearer your-secret-upload-token-here"
```

#### Python 示例

```python
import requests

BASE_URL = "http://your-domain.com"
AUTH_TOKEN = "your-secret-upload-token-here"

def upload_file(file_path):
    """上传单个文件"""
    url = f"{BASE_URL}/upload"
    headers = {"Authorization": f"Bearer {AUTH_TOKEN}"}
    
    with open(file_path, 'rb') as f:
        files = {'file': f}
        response = requests.post(url, files=files, headers=headers)
    
    return response.json()

# 上传文件
result = upload_file("/path/to/document.pdf")
print(result)
```

#### JavaScript 示例

```javascript
async function uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await fetch('http://your-domain.com/upload', {
        method: 'POST',
        headers: {
            'Authorization': 'Bearer your-secret-upload-token-here'
        },
        body: formData
    });
    
    return await response.json();
}

// 使用示例
const fileInput = document.getElementById('fileInput');
const file = fileInput.files[0];
const result = await uploadFile(file);
console.log(result);
```

### API 安全特性

#### 认证
所有上传请求必须包含有效的 `Authorization: Bearer YOUR_TOKEN` 头。

#### 文件验证
- **类型检查**：验证 MIME 类型和文件扩展名
- **大小检查**：限制单个文件最大 2GB（可配置）
- **内容验证**：检查文件头真实性（可选深度扫描）
- **配额检查**：限制每个用户的总存储使用量

#### 速率控制
- **每分钟限制**：默认 5 个文件/分钟/用户
- **每小时限制**：默认 50 个文件/小时/用户
- **并发限制**：默认 3 个同时上传/用户
- **API 冷却**：默认上传之间 1 秒冷却时间

#### 错误代码

| 代码 | 消息 | 描述 |
|------|---------|-------------|
| 400 | 认证失败 | 未提供令牌或令牌无效 |
| 400 | 文件获取失败 | multipart 表单解析错误 |
| 400 | 不允许的文件类型 | 文件类型或扩展名被禁止 |
| 400 | 文件大小超过限制 | 文件超过配置的最大值 |
| 403 | 存储配额已超出 | 用户已用完配额 |
| 429 | 请等待 X 秒后再试 | 超出速率限制 |
| 500 | 上传失败 | 服务器错误或 Telegram API 错误 |

> [!TIP]
> 有关更详细的 API 文档，请参阅仓库中的 [上传API使用说明.md](上传API使用说明.md) 文件。

## 贡献

如果您有任何进一步的想法，欢迎为此项目做出贡献

## 联系我

[![Telegram 频道](https://img.shields.io/static/v1?label=Join&message=Telegram%20Channel&color=blueviolet&style=for-the-badge&logo=telegram&logoColor=violet)](https://xn--r1a.click/wrench_labs)
[![Telegram 群组](https://img.shields.io/static/v1?label=Join&message=Telegram%20Group&color=blueviolet&style=for-the-badge&logo=telegram&logoColor=violet)](https://xn--r1a.click/AlteredVoid)

您可以通过我的 [Telegram 群组](https://xn--r1a.click/AlteredVoid)联系我，或者在 [@EverythingSuckz](https://xn--r1a.click/EverythingSuckz) 给我发消息

## 致谢

- [@celestix](https://github.com/celestix) 的 [gotgproto](https://github.com/celestix/gotgproto)
- [@divyam234](https://github.com/divyam234/teldrive) 的 [Teldrive](https://github.com/divyam234/teldrive) 项目
- [@karu](https://github.com/krau) 添加图片支持

## 版权

Copyright (C) 2023 [EverythingSuckz](https://github.com/EverythingSuckz) under [GNU Affero General Public License](https://www.gnu.org/licenses/agpl-3.0.en.html).

TG-FileStreamBot 是自由软件：您可以随意使用、学习、分享和改进它。具体来说，您可以根据自由软件基金会发布的 [GNU Affero General Public License](https://www.gnu.org/licenses/agpl-3.0.en.html) 的条款重新分发和/或修改它，可以是许可证的第 3 版，也可以是（根据您的选择）任何更高版本。同时请记住，此仓库的所有分支必须是开源的，并且必须使用相同的许可证。

