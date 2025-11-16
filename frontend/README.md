# 前端项目 - 加密货币实时数据监控系统

使用 React + TypeScript + Vite 构建的前端应用，用于实时展示加密货币K线数据。

## 技术栈

- **框架：** React 18+ with TypeScript
- **构建工具：** Vite
- **样式：** Tailwind CSS
- **图表库：** TradingView Lightweight Charts
- **HTTP 客户端：** Axios

## 项目结构

```
frontend/
├── src/
│   ├── components/          # React 组件
│   │   ├── KLineChart.tsx
│   │   ├── SymbolSelector.tsx
│   │   ├── IntervalSelector.tsx
│   │   └── PriceDisplay.tsx
│   ├── hooks/              # 自定义 Hooks
│   │   └── useWebSocket.ts
│   ├── services/           # API 服务
│   │   └── api.ts
│   ├── App.tsx             # 主应用组件
│   └── main.tsx            # 应用入口
├── package.json
├── vite.config.ts
└── tailwind.config.js
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 环境配置

复制环境变量模板文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置 API 地址：

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
```

### 开发模式

```bash
npm run dev
```

应用将在 `http://localhost:5173` 启动。

### 构建生产版本

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

### 预览生产构建

```bash
npm run preview
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `VITE_API_URL` | 后端 API 地址 | http://localhost:8080 |
| `VITE_WS_URL` | WebSocket 连接地址 | ws://localhost:8080/ws |

**注意：** Vite 要求环境变量必须以 `VITE_` 开头才能在前端代码中访问。

## 开发规范

### 组件开发

- 遵循单一职责原则
- 保持组件的可复用性
- 使用 TypeScript 定义明确的 props 接口
- 保持状态管理本地化

### 样式规范

- 使用 Tailwind CSS 工具类
- 遵循深色主题设计
- 保持设计系统一致性

### 代码规范

- 使用有意义的变量和函数命名
- 保持函数简洁，单一职责
- 添加必要的注释和文档

## 主要功能

- **实时K线图展示**：使用 TradingView Lightweight Charts 展示K线图
- **代币对选择**：支持 BTC/USDT, ETH/USDT, BNB/USDT
- **时间粒度切换**：支持 1m, 5m, 1h
- **实时数据更新**：通过 WebSocket 实时接收数据
- **深色主题**：专业的深色界面设计

## 许可证

个人使用项目
