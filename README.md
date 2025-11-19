# 加密货币实时数据监控系统

一个个人使用的加密货币交易对实时数据监控系统，提供可交互的K线图展示和实时数据存储功能。

## 项目概述

本项目用于实时监控加密货币交易对（BTC/USDT, ETH/USDT, BNB/USDT）的K线数据，并通过WebSocket实时推送数据到前端展示。

## 技术栈

- **后端：** Go 1.25.3 + Gin + GORM + PostgreSQL
- **前端：** React 18+ + Vite + TradingView Lightweight Charts
- **数据源：** Binance Public API
- **实时通信：** WebSocket (Gorilla WebSocket)

## 项目结构

```
.
├── cmd/
│   └── server/
│       └── main.go          # 应用入口
├── internal/
│   ├── api/
│   │   ├── handlers/        # API处理器
│   │   └── routes.go        # 路由配置
│   ├── service/             # 业务逻辑层
│   ├── repository/          # 数据访问层
│   └── models/              # 数据模型
├── pkg/
│   └── database/            # 数据库连接
│       └── postgres.go
├── go.mod                   # Go模块定义
└── README.md                # 项目说明
```

## 环境要求

- Go 1.25.3 或更高版本
- PostgreSQL 15 或更高版本
- Docker 和 Docker Compose（用于数据库）

## 快速开始

### 1. 环境配置

**选择环境：**

- **测试网（默认）**：如果没有 `.env` 文件，程序会自动使用测试网配置
- **生产网**：复制生产网配置模板并创建 `.env` 文件

**配置步骤：**

```bash
# 方式一：使用测试网（无需配置，直接运行）
# 程序会自动检测到没有 .env 文件，使用测试网配置

# 方式二：使用生产网
cp .env.prod.example .env
# 编辑 .env 文件，根据需要修改配置

# 方式三：使用测试网（显式配置）
cp .env.test.example .env
# 编辑 .env 文件，根据需要修改配置
```

**环境配置文件说明：**
- `.env.test.example` - 测试网环境配置模板
- `.env.prod.example` - 生产网环境配置模板
- `.env` - 实际使用的环境配置文件（不会被提交到版本控制）

**注意：** 如果没有 `.env` 文件，程序会默认使用 Binance 测试网，适合开发和测试使用。

### 2. 启动数据库

使用 Docker Compose 启动 PostgreSQL：

```bash
# 使用管理脚本
chmod +x scripts/db.sh
./scripts/db.sh start

# 或直接使用 docker-compose
docker-compose up -d postgres
```

### 3. 安装依赖

**后端依赖：**
```bash
go mod download
go mod tidy
```

**前端依赖：**
```bash
cd frontend
npm install
```

### 4. 启动服务

**方式一：使用统一管理脚本（推荐）**

```bash
# 给脚本添加执行权限（首次使用）
chmod +x scripts/manage.sh

# 启动所有服务（后端和前端）
./scripts/manage.sh start

# 查看服务状态
./scripts/manage.sh status

# 停止所有服务
./scripts/manage.sh stop

# 重启所有服务
./scripts/manage.sh restart
```

**方式二：手动启动**

```bash
# 启动后端服务
go run cmd/server/main.go

# 启动前端服务（新终端）
cd frontend
npm run dev
```

**服务地址：**
- 后端服务：`http://localhost:8080`
- 前端服务：`http://localhost:5173`

## 服务管理

### 使用管理脚本

项目提供了统一的服务管理脚本 `scripts/manage.sh`，支持以下命令：

| 命令 | 说明 |
|------|------|
| `start` | 启动所有服务（后端和前端） |
| `stop` | 停止所有服务 |
| `restart` | 重启所有服务 |
| `status` | 查看所有服务状态 |
| `start-backend` | 仅启动后端服务 |
| `stop-backend` | 仅停止后端服务 |
| `start-frontend` | 仅启动前端服务 |
| `stop-frontend` | 仅停止前端服务 |

**使用示例：**

```bash
# 查看服务状态
./scripts/manage.sh status

# 启动所有服务
./scripts/manage.sh start

# 仅启动后端服务
./scripts/manage.sh start-backend

# 停止所有服务
./scripts/manage.sh stop
```

**日志文件：**
- 后端日志：`server.log`
- 前端日志：`frontend.log`
- 进程ID文件：`server.pid`、`frontend.pid`

### 数据库管理

使用 `scripts/db.sh` 管理数据库：

```bash
./scripts/db.sh start    # 启动数据库
./scripts/db.sh stop     # 停止数据库
./scripts/db.sh status   # 查看数据库状态
./scripts/db.sh logs     # 查看数据库日志
./scripts/db.sh shell    # 连接到数据库命令行
./scripts/db.sh backup   # 备份数据库
./scripts/db.sh restore  # 恢复数据库
```

## 开发

### 代码规范

本项目遵循以下代码规范：

- 遵循 Go 官方代码规范
- 使用有意义的变量和函数命名
- 保持函数简洁，单一职责
- 添加必要的注释和文档

### 目录说明

- `cmd/server/`: 应用入口点
- `internal/`: 内部包，不对外暴露
  - `api/`: API 层，处理 HTTP 请求
  - `service/`: 业务逻辑层
  - `repository/`: 数据访问层
  - `models/`: 数据模型定义
- `pkg/`: 可复用的公共包
  - `database/`: 数据库连接和配置

## API 端点

### RESTful API

- `GET /api/v1/klines` - 查询历史K线数据
- `GET /api/v1/symbols` - 获取支持的交易对列表

### WebSocket

- `ws://localhost:8080/ws` - WebSocket 连接端点

## 环境变量

| 变量名 | 说明 | 默认值（无 .env 文件时） | 默认值（有 .env 文件时） |
|--------|------|------------------------|------------------------|
| `DB_HOST` | 数据库主机 | localhost | localhost |
| `DB_PORT` | 数据库端口 | 5432 | 5432 |
| `DB_USER` | 数据库用户 | postgres | postgres |
| `DB_PASSWORD` | 数据库密码 | postgres | postgres |
| `DB_NAME` | 数据库名称 | crypto_monitor | crypto_monitor |
| `PORT` | 服务端口 | 8080 | 8080 |
| `BINANCE_API_URL` | Binance API URL | https://testnet.binance.vision | https://api.binance.com |
| `BINANCE_WS_URL` | Binance WebSocket URL | wss://stream.testnet.binance.vision/ws | wss://stream.binance.com:9443/ws |

**重要提示：**
- 如果没有 `.env` 文件，程序会自动使用 **Binance 测试网**配置
- 测试网适合开发和测试，不会产生真实交易
- 生产环境请务必创建 `.env` 文件并配置正确的生产网地址

## 许可证

个人使用项目

