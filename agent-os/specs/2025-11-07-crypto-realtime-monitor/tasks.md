# 任务列表

**创建日期：** 2025-11-07  
**状态：** 待开始  
**优先级：** P0 = 必须实现，P1 = 有时间就做，P2 = 后续考虑

## 任务分组

任务按开发顺序和依赖关系分组，建议按顺序执行。

---

## 阶段 1: 项目初始化和数据库设计

### 1.1 后端项目初始化
**优先级：** P0  
**依赖：** 无  
**预估时间：** 1小时

**任务：**
- [ ] 创建后端项目目录结构
- [ ] 初始化 Go Modules (`go mod init`)
- [ ] 创建基础目录结构：
  - `cmd/server/main.go`
  - `internal/api/handlers/`
  - `internal/api/routes.go`
  - `internal/service/`
  - `internal/repository/`
  - `internal/models/`
  - `pkg/database/`
- [ ] 安装依赖包：
  - `github.com/gin-gonic/gin`
  - `gorm.io/gorm`
  - `gorm.io/driver/postgres`
  - `github.com/gorilla/websocket`
- [ ] 创建 `.env` 文件模板
- [ ] 创建 `README.md` 说明文档

### 1.2 前端项目初始化
**优先级：** P0  
**依赖：** 无  
**预估时间：** 1小时

**任务：**
- [ ] 使用 Vite 创建 React + TypeScript 项目
- [ ] 创建基础目录结构：
  - `src/components/`
  - `src/hooks/`
  - `src/services/`
  - `src/App.tsx`
  - `src/main.tsx`
- [ ] 安装依赖包：
  - `react`、`react-dom`
  - `lightweight-charts`
  - `axios` 或 `fetch`
- [ ] 配置 Tailwind CSS（或使用简单CSS）
- [ ] 创建 `.env` 文件模板
- [ ] 创建 `README.md` 说明文档

### 1.3 数据库设计和迁移
**优先级：** P0  
**依赖：** 1.1  
**预估时间：** 2小时

**任务：**
- [ ] 设计 `klines` 表结构
- [ ] 创建 GORM 模型定义 (`internal/models/kline.go`)
- [ ] 创建数据库迁移文件
- [ ] 实现数据库连接配置 (`pkg/database/postgres.go`)
- [ ] 创建数据库索引：
  - `(symbol, interval, open_time)` 唯一索引
  - `(symbol, interval, open_time)` 查询索引
  - `(symbol, interval)` 查询索引
- [ ] 测试数据库连接
- [ ] 编写数据库初始化脚本

### 1.4 启停脚本创建
**优先级：** P0  
**依赖：** 1.1, 1.2  
**预估时间：** 1小时

**任务：**
- [ ] 创建统一启停脚本 (`scripts/manage.sh`)：
  - 支持命令：`start`, `stop`, `restart`, `status`
  - 启动后端服务：
    - 检查环境变量配置
    - 检查数据库连接（可选）
    - 后台启动后端服务（`go run cmd/server/main.go`）
    - 记录进程ID到文件（`backend.pid`）
    - 输出启动日志
  - 启动前端服务：
    - 检查环境变量配置
    - 后台启动前端开发服务器（`npm run dev`）
    - 记录进程ID到文件（`frontend.pid`）
    - 输出启动日志
  - 停止后端服务：
    - 读取 `backend.pid` 文件
    - 优雅停止后端服务（发送SIGTERM信号）
    - 清理进程ID文件
    - 输出停止日志
  - 停止前端服务：
    - 读取 `frontend.pid` 文件
    - 停止前端服务
    - 清理进程ID文件
    - 输出停止日志
  - 状态检查：
    - 检查后端服务运行状态
    - 检查前端服务运行状态
    - 显示服务状态信息
  - 重启功能：
    - 调用停止功能
    - 等待服务完全停止
    - 调用启动功能
- [ ] 为脚本添加执行权限 (`chmod +x scripts/manage.sh`)
- [ ] 测试脚本功能：
  - 测试启动后端
  - 测试启动前端
  - 测试停止后端
  - 测试停止前端
  - 测试重启功能
  - 测试状态检查
- [ ] 在项目根目录 README 中说明脚本使用方法

---

## 阶段 2: 后端核心功能

### 2.1 Binance API 集成
**优先级：** P0  
**依赖：** 1.1  
**预估时间：** 4小时

**任务：**
- [ ] 创建 Binance 服务 (`internal/service/binance.go`)
- [ ] 实现 Binance REST API 客户端
  - 获取历史K线数据：`GET /api/v3/klines`
  - 错误处理和重试机制（指数退避）
- [ ] 实现 Binance WebSocket 客户端
  - 连接 Binance WebSocket Streams
  - 订阅K线数据流：`<symbol>@kline_<interval>`
  - 处理消息解析
- [ ] 实现数据转换逻辑（Binance格式 → 内部模型）
- [ ] 添加日志记录
- [ ] 测试 Binance API 连接

### 2.2 数据存储逻辑
**优先级：** P0  
**依赖：** 1.3, 2.1  
**预估时间：** 3小时

**任务：**
- [ ] 创建 K线数据仓库 (`internal/repository/kline.go`)
- [ ] 实现数据存储方法：
  - `CreateKline()` - 创建K线数据
  - `CreateOrUpdateKline()` - 使用 UPSERT 逻辑
  - `GetKlines()` - 查询历史K线数据
- [ ] 实现去重逻辑（基于 `symbol, interval, open_time`）
- [ ] 实现批量插入优化
- [ ] 添加错误处理（数据库连接失败时继续运行）
- [ ] 测试数据存储功能

### 2.3 WebSocket 服务实现
**优先级：** P0  
**依赖：** 2.1, 2.2  
**预估时间：** 4小时

**任务：**
- [ ] 创建 WebSocket 服务 (`internal/service/websocket.go`)
- [ ] 实现 WebSocket 连接管理：
  - 使用 Gorilla WebSocket 管理连接
  - 维护客户端连接池
  - 处理连接断开和清理
- [ ] 实现订阅/取消订阅逻辑：
  - 处理客户端订阅消息
  - 管理订阅列表
- [ ] 实现消息广播：
  - 接收 Binance 推送的数据
  - 存储到数据库
  - 广播到所有订阅的客户端
- [ ] 实现节流更新（每秒最多推送一次）
- [ ] 添加连接状态管理
- [ ] 测试 WebSocket 服务

### 2.4 RESTful API 实现
**优先级：** P0  
**依赖：** 2.2  
**预估时间：** 3小时

**任务：**
- [ ] 创建 API 处理器 (`internal/api/handlers/kline.go`)
- [ ] 实现 `GET /api/v1/klines` 端点：
  - 查询参数验证（symbol, interval, start_time, end_time, limit）
  - 调用仓库方法查询数据
  - 返回JSON响应
- [ ] 实现 `GET /api/v1/symbols` 端点：
  - 返回支持的交易对列表
- [ ] 实现统一响应格式：
  - 成功响应：`{code: 200, message: "success", data: ...}`
  - 错误响应：`{code: 400, message: "error", data: null}`
- [ ] 添加错误处理中间件
- [ ] 配置路由 (`internal/api/routes.go`)
- [ ] 测试 API 端点

### 2.5 主程序集成
**优先级：** P0  
**依赖：** 2.1, 2.3, 2.4  
**预估时间：** 2小时

**任务：**
- [ ] 实现主程序 (`cmd/server/main.go`)
- [ ] 初始化数据库连接
- [ ] 初始化 Gin 路由
- [ ] 启动 Binance WebSocket 连接
- [ ] 启动 WebSocket 服务
- [ ] 启动 HTTP 服务器
- [ ] 实现优雅关闭
- [ ] 添加环境变量配置
- [ ] 测试完整后端流程

---

## 阶段 3: 前端核心功能

### 3.1 API Service 实现
**优先级：** P0  
**依赖：** 1.2, 2.4  
**预估时间：** 2小时

**任务：**
- [ ] 创建 API Service (`src/services/api.ts`)
- [ ] 实现 `getKlines()` 方法：
  - 查询历史K线数据
  - 支持参数：symbol, interval, startTime, endTime, limit
  - 错误处理
- [ ] 实现 `getSymbols()` 方法：
  - 获取支持的交易对列表
- [ ] 配置 API 基础URL（从环境变量读取）
- [ ] 添加请求拦截器和错误处理
- [ ] 测试 API 调用

### 3.2 WebSocket Hook 实现
**优先级：** P0  
**依赖：** 1.2, 2.3  
**预估时间：** 3小时

**任务：**
- [ ] 创建 `useWebSocket` Hook (`src/hooks/useWebSocket.ts`)
- [ ] 实现 WebSocket 连接管理：
  - 连接建立
  - 连接断开处理
  - 自动重连机制（最多5次，指数退避）
- [ ] 实现订阅/取消订阅逻辑：
  - 发送订阅消息
  - 处理订阅响应
- [ ] 实现消息接收和处理：
  - 接收K线数据更新
  - 节流更新（每秒最多更新一次）
- [ ] 实现连接状态管理：
  - `connected` 状态
  - `error` 状态
- [ ] 测试 WebSocket 连接

### 3.3 K线图组件实现
**优先级：** P0  
**依赖：** 3.1, 3.2  
**预估时间：** 6小时

**任务：**
- [ ] 安装 TradingView Lightweight Charts
- [ ] 创建 `KLineChart` 组件 (`src/components/KLineChart.tsx`)
- [ ] 实现图表初始化：
  - 创建图表实例
  - 配置深色主题
  - 设置图表尺寸
- [ ] 实现K线图渲染：
  - 添加 Candlestick Series
  - 配置颜色方案（绿色上涨，红色下跌）
  - 渲染历史数据
- [ ] 实现成交量图：
  - 添加 Histogram Series
  - 配置在K线图下方显示
  - 共享时间轴
  - 颜色与K线涨跌对应
- [ ] 实现实时数据更新：
  - 接收 WebSocket 数据
  - 更新K线数据
  - 更新成交量数据
- [ ] 实现鼠标交互：
  - 十字线跟随鼠标
  - 价格标签显示
  - 右侧Y轴当前价格标签（红色高亮）
  - 当前价格水平线（虚线）
- [ ] 实现数据格式化：
  - 时间戳转换
  - 价格格式化
  - 成交量格式化
- [ ] 测试K线图渲染和交互

### 3.4 代币对选择组件
**优先级：** P0  
**依赖：** 3.1  
**预估时间：** 2小时

**任务：**
- [ ] 创建 `SymbolSelector` 组件 (`src/components/SymbolSelector.tsx`)
- [ ] 实现下拉选择框：
  - 显示交易对列表（BTC/USDT, ETH/USDT, BNB/USDT）
  - 默认选中 BTC/USDT
  - 切换时触发回调
- [ ] 实现样式：
  - 深色主题
  - 选中状态高亮
- [ ] 集成到主界面（顶部控制栏）
- [ ] 测试组件功能

### 3.5 时间粒度选择组件
**优先级：** P0  
**依赖：** 无  
**预估时间：** 2小时

**任务：**
- [ ] 创建 `IntervalSelector` 组件 (`src/components/IntervalSelector.tsx`)
- [ ] 实现按钮组：
  - 显示时间粒度选项（1m, 5m, 1h）
  - 当前选中高亮显示
  - 切换时触发回调
- [ ] 实现样式：
  - 深色主题
  - 选中状态高亮
- [ ] 集成到主界面（顶部左侧）
- [ ] 测试组件功能

### 3.6 价格显示组件
**优先级：** P0  
**依赖：** 3.3  
**预估时间：** 2小时

**任务：**
- [ ] 创建 `PriceDisplay` 组件 (`src/components/PriceDisplay.tsx`)
- [ ] 实现当前价格显示：
  - 显示当前价格
  - 价格变化百分比
- [ ] 实现K线详细信息显示：
  - 开盘价 (O)
  - 最高价 (H)
  - 最低价 (L)
  - 收盘价 (C)
  - 成交量 (Volume)
  - 时间戳 (Time)
- [ ] 实现悬浮信息显示：
  - 鼠标悬浮时显示详细信息
  - 格式化显示
- [ ] 集成到主界面（顶部控制栏）
- [ ] 测试组件功能

### 3.7 主应用集成
**优先级：** P0  
**依赖：** 3.3, 3.4, 3.5, 3.6  
**预估时间：** 4小时

**任务：**
- [ ] 创建主应用组件 (`src/App.tsx`)
- [ ] 实现状态管理：
  - `symbol` - 当前交易对
  - `interval` - 当前时间粒度
  - `klineData` - K线数据数组
  - `currentPrice` - 当前价格
  - `selectedKline` - 选中的K线数据
  - `wsConnected` - WebSocket连接状态
  - `loading` - 加载状态
  - `error` - 错误信息
- [ ] 实现数据加载逻辑：
  - 初始加载历史数据
  - 切换交易对时重新加载
  - 切换时间粒度时重新加载
- [ ] 实现 WebSocket 集成：
  - 连接WebSocket
  - 订阅K线数据
  - 处理实时更新
- [ ] 实现界面布局：
  - 顶部控制栏（时间粒度、交易对、价格信息）
  - 主K线图区域
  - 成交量图区域
  - 底部时间轴
- [ ] 实现深色主题样式
- [ ] 测试完整前端流程

---

## 阶段 4: 集成和优化

### 4.1 前后端集成测试
**优先级：** P0  
**依赖：** 2.5, 3.7  
**预估时间：** 3小时

**任务：**
- [ ] 启动后端服务
- [ ] 启动前端服务
- [ ] 测试完整数据流：
  - 前端加载历史数据
  - WebSocket实时推送
  - 数据存储验证
  - 切换交易对
  - 切换时间粒度
- [ ] 测试错误场景：
  - WebSocket连接断开
  - API调用失败
  - 数据库连接失败
- [ ] 修复发现的bug
- [ ] 验证所有P0功能

### 4.2 错误处理完善
**优先级：** P1  
**依赖：** 4.1  
**预估时间：** 2小时

**任务：**
- [ ] 前端错误提示UI：
  - API调用失败提示
  - WebSocket连接失败提示
  - 错误消息显示
- [ ] 加载状态提示：
  - 数据加载中提示
  - WebSocket连接中提示
- [ ] 连接状态提示：
  - WebSocket连接状态显示
  - 自动重连提示
- [ ] 后端错误日志完善
- [ ] 测试错误处理场景

### 4.3 性能优化
**优先级：** P1  
**依赖：** 4.1  
**预估时间：** 2小时

**任务：**
- [ ] 前端节流更新优化：
  - 确保每秒最多更新一次
  - 优化渲染性能
- [ ] 数据库查询优化：
  - 验证索引使用
  - 优化查询语句
- [ ] 前端数据缓存：
  - 缓存历史数据
  - 避免重复请求
- [ ] 测试性能指标

### 4.4 代码优化和文档
**优先级：** P1  
**依赖：** 4.1  
**预估时间：** 2小时

**任务：**
- [ ] 代码格式化：
  - 后端：`gofmt`
  - 前端：`Prettier`（可选）
- [ ] 代码审查和重构：
  - 优化代码结构
  - 提取公共逻辑
  - 改进错误处理
- [ ] 更新文档：
  - 后端 README
  - 前端 README
  - API 文档
  - 部署说明
- [ ] 添加代码注释

---

## 任务统计

### 按优先级统计

**P0 - 必须实现：**
- 阶段1: 4个任务组（1.1, 1.2, 1.3, 1.4）
- 阶段2: 5个任务组
- 阶段3: 7个任务组
- 阶段4: 1个任务组
- **总计：17个任务组**

**P1 - 有时间就做：**
- 阶段4: 3个任务组
- **总计：3个任务组**

### 预估时间

**P0任务总时间：** 约 41-46小时（阶段1增加1小时）  
**P1任务总时间：** 约 6小时  
**总计：** 约 47-52小时

**按1-2周时间线：**
- 第1周（40小时）：完成所有P0任务
- 第2周（10小时）：完成P1任务和优化

---

## 开发顺序建议

### 第1周：核心功能

**Day 1-2: 项目初始化和数据库**
- 1.1 后端项目初始化
- 1.2 前端项目初始化
- 1.3 数据库设计和迁移
- 1.4 启停脚本创建

**Day 3-4: 后端核心功能**
- 2.1 Binance API 集成
- 2.2 数据存储逻辑
- 2.3 WebSocket 服务实现

**Day 5-7: 前端核心功能**
- 2.4 RESTful API 实现
- 2.5 主程序集成
- 3.1 API Service 实现
- 3.2 WebSocket Hook 实现
- 3.3 K线图组件实现（重点）

**Day 8-10: 前端完善**
- 3.4 代币对选择组件
- 3.5 时间粒度选择组件
- 3.6 价格显示组件
- 3.7 主应用集成

**Day 11-14: 集成和测试**
- 4.1 前后端集成测试

### 第2周：优化和完善

**Day 15-17: 优化**
- 4.2 错误处理完善
- 4.3 性能优化
- 4.4 代码优化和文档

---

## 注意事项

1. **依赖关系：** 严格按照任务依赖顺序执行
2. **测试：** 每个阶段完成后进行测试
3. **错误处理：** 优先实现P0功能，P1功能有时间再做
4. **性能：** 关注实时数据更新性能，确保节流更新正常工作
5. **数据一致性：** 确保数据库存储和实时推送的数据一致

---

## 参考资料

- [Binance API 文档](https://binance-docs.github.io/apidocs/spot/en/)
- [TradingView Lightweight Charts 文档](https://tradingview.github.io/lightweight-charts/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [Gorilla WebSocket 文档](https://github.com/gorilla/websocket)
