# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

TG-FileStreamBot 是一个用 Go 语言开发的 Telegram 文件流机器人，用于为 Telegram 文件生成可直接流式传输的链接。该项目基于 gotgproto 库构建，使用 Gin 框架提供 HTTP 服务。

## 常用命令

### 构建和运行

```bash
# 本地构建
go build ./cmd/fsb/

# 运行机器人（需要先配置 fsb.env 文件）
./fsb run

# 带参数运行
./fsb run --port 8080 --dev

# 生成用户会话
./fsb session --api-id <API_ID> --api-hash <API_HASH>

# Docker 运行
docker-compose up -d

# 直接使用 Docker 镜像
docker run --env-file fsb.env ghcr.io/everythingsuckz/fsb:latest
```

### 开发相关

```bash
# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 开发模式运行（启用调试日志和代理）
./fsb run --dev
```

## 核心架构

### 目录结构

```
cmd/fsb/          # 主程序入口和命令行处理
internal/
├── bot/          # Telegram 机器人客户端管理
├── cache/        # 缓存系统
├── commands/     # Telegram 命令处理
├── routes/       # HTTP 路由处理
├── types/        # 数据类型定义
└── utils/        # 工具函数
config/           # 配置管理
pkg/             # 公共包
```

### 关键组件

#### 1. 应用启动流程 (cmd/fsb/run.go)
应用启动的核心流程：
- 初始化日志系统
- 加载配置文件（`fsb.env`）
- 启动主要 Telegram 机器人客户端
- 初始化缓存系统
- 启动工作客户端池
- 启动用户机器人（如果配置了 USER_SESSION）
- 启动 HTTP 服务器

#### 2. 客户端管理 (internal/bot/)
- `client.go`: 主机器人客户端启动和管理
- `workers.go`: 工作客户端池，支持多机器人负载均衡
- `userbot.go`: 用户机器人，用于自动将工作机器人添加到日志频道
- `middleware.go`: 中间件处理

#### 3. 配置系统 (config/config.go)
- 支持环境变量和 `.env` 文件配置
- 必需配置：`API_ID`, `API_HASH`, `BOT_TOKEN`, `LOG_CHANNEL`
- 可选配置：`PORT`, `HOST`, `HASH_LENGTH`, `MULTI_TOKEN*` 等
- 支持多机器人配置以提高并发性能

#### 4. HTTP 路由 (internal/routes/)
- 使用反射机制自动加载路由
- 主要提供文件流式传输服务
- 支持断点续传（通过 range-parser）

#### 5. 命令系统 (internal/commands/)
- 处理 Telegram 机器人的命令
- 主要命令：`/start`, 流媒体相关命令

### 技术栈

- **语言**: Go 1.21+
- **Telegram 客户端**: gotgproto (基于 gotd/td)
- **Web 框架**: Gin
- **日志**: zap
- **配置**: envconfig + godotenv
- **缓存**: freecache + SQLite
- **数据库**: SQLite (用于会话存储)

## 配置要求

### 必需环境变量
- `API_ID`: Telegram API ID (从 my.telegram.org 获取)
- `API_HASH`: Telegram API Hash (从 my.telegram.org 获取)
- `BOT_TOKEN`: Telegram 机器人 Token (从 @BotFather 获取)
- `LOG_CHANNEL`: 日志频道 ID (用于存储文件消息)

### 可选环境变量
- `PORT`: 服务器端口 (默认 8080)
- `HOST`: 服务器主机地址
- `HASH_LENGTH`: 生成链接的哈希长度 (默认 6，范围 5-32)
- `MULTI_TOKEN1`, `MULTI_TOKEN2`...: 多机器人支持
- `USER_SESSION`: 用户会话字符串 (用于自动添加机器人到频道)
- `DEV`: 开发模式开关

## 多机器人架构

项目支持配置多个工作机器人来分担负载：
- 主机器人处理用户交互
- 工作机器人池处理文件下载和流传输
- 通过 `MULTI_TOKEN*` 环境变量配置工作机器人
- 最多支持 50 个工作机器人

## 部署选项

### 本地部署
1. 创建 `fsb.env` 配置文件
2. 构建：`go build ./cmd/fsb/`
3. 运行：`./fsb run`

### Docker 部署
- 使用 `docker-compose.yaml` 进行容器化部署
- 支持环境文件挂载和端口映射
- 官方镜像：`ghcr.io/everythingsuckz/fsb:latest`

### 云平台部署
- 支持 Heroku 一键部署
- 支持 Koyeb 平台部署

## 开发注意事项

1. **会话文件**: 默认使用 SQLite 存储会话信息，文件名为 `fsb.session`
2. **日志系统**: 使用 zap 结构化日志，开发模式下会启用详细日志
3. **代理支持**: 开发模式下自动配置代理设置
4. **错误处理**: 使用 panic 处理关键启动错误，确保服务能正常启动或终止
5. **并发设计**: 支持多个 Telegram 客户端并发工作以提高性能

## 版本信息

当前版本：3.1.0
许可证：GNU Affero General Public License v3.0