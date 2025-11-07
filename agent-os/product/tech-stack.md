# 产品技术栈

本文档定义了产品开发中使用的所有技术栈选择。**聚焦快速实现，选择最简单、最直接的技术方案。**

## 框架与运行时

- **应用框架：** Gin (Go Web框架)
- **语言/运行时：** Go 1.25.3
- **包管理器：** Go Modules

**选择原因：**
- Go语言高性能、并发能力强，适合实时数据处理
- Gin框架轻量级、简单易用，快速开发API
- Go Modules是Go官方包管理工具

## 前端

- **JavaScript 框架：** React 18+
- **构建工具：** Vite
- **CSS 框架：** Tailwind CSS（或简单CSS）
- **图表库：** TradingView Lightweight Charts
- **WebSocket客户端：** native WebSocket API
- **包管理器：** npm

**选择原因：**
- React生态成熟，快速开发
- Vite构建速度快，开发体验好
- Tailwind CSS快速开发（或直接用CSS，更快）
- TradingView Lightweight Charts专业K线图，API简洁
- 原生WebSocket足够用，无需额外库

## 数据库与存储

- **数据库：** PostgreSQL 15+（本地运行）
- **ORM/查询构建器：** GORM (Go ORM)

**选择原因：**
- PostgreSQL支持时间序列数据，适合存储K线数据
- GORM是Go生态最成熟的ORM，开发效率高
- **个人使用不需要Redis缓存**，直接查数据库即可

## 实时通信

- **WebSocket库：** Gorilla WebSocket
- **协议：** WebSocket (WS，本地开发可用HTTP)

**选择原因：**
- WebSocket提供双向实时通信
- Gorilla WebSocket是Go生态最成熟的WebSocket库
- 简单易用，文档完善

## 数据源

- **主要数据源：** Binance Public API
- **WebSocket流：** Binance WebSocket Streams
- **RESTful API：** Binance REST API

**选择原因：**
- 提供免费的公开API，无需认证
- WebSocket流支持实时价格推送
- 文档完善，易于集成

## 测试与质量（简化）

- **后端测试：** Go标准库testing（基础测试即可）
- **前端测试：** 暂不写测试（快速实现优先）
- **代码格式化：**
  - 后端：gofmt（Go自带）
  - 前端：Prettier（可选）

**选择原因：**
- MVP阶段以功能实现为主，测试后续补充
- 使用最简单的工具即可

## 部署（本地开发）

- **后端：** 本地运行（go run）
- **前端：** 本地运行（npm run dev）
- **数据库：** 本地PostgreSQL（Docker或直接安装）

**选择原因：**
- 个人使用，本地运行即可
- 无需复杂的部署流程
- 快速启动，快速开发

## 第三方服务（暂不需要）

- **监控：** 暂不需要
- **日志：** Go标准库log
- **错误追踪：** 暂不需要

**选择原因：**
- 个人使用，本地运行，无需复杂监控
- 基础日志足够排查问题

## 技术决策记录

### 后端框架选择
**决策：** Gin Web Framework
**原因：** 轻量级、简单易用，快速开发API
**日期：** 2025-11-07

### 前端图表库选择
**决策：** TradingView Lightweight Charts（优先）
**原因：** 专业K线图库，API简洁，性能优秀。备选ECharts（中文文档好）
**日期：** 2025-11-07

### 数据库选择
**决策：** PostgreSQL（本地）
**原因：** 支持时间序列数据，GORM支持好，个人使用本地即可
**日期：** 2025-11-07

### 缓存方案
**决策：** 暂不使用Redis
**原因：** 个人使用，数据量不大，直接查数据库即可，简化架构
**日期：** 2025-11-07

### 实时通信方案
**决策：** WebSocket (Gorilla WebSocket)
**原因：** 双向实时通信，成熟稳定，简单易用
**日期：** 2025-11-07

### 数据源选择
**决策：** Binance Public API
**原因：** 免费公开API，文档完善，易于集成
**日期：** 2025-11-07

### 部署方案
**决策：** 本地开发运行
**原因：** 个人使用，本地运行即可，无需复杂部署
**日期：** 2025-11-07

