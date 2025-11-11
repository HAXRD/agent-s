# 加密货币实时数据监控系统 - 详细规格说明

**版本：** 1.0  
**创建日期：** 2025-11-07  
**状态：** 待实现

## 1. 概述

### 1.1 项目目标

构建一个个人使用的加密货币交易对实时数据监控系统，提供可交互的K线图展示和实时数据存储功能。系统聚焦快速实现，满足个人使用需求。

### 1.2 核心功能

- **实时K线图展示**：可交互的K线图，支持代币对选择、时间粒度切换、鼠标悬浮显示价格
- **实时数据存储**：所有实时数据自动落表存储到数据库
- **历史数据查询**：支持查询历史K线数据

### 1.3 技术栈

- **后端：** Go 1.25.3 + Gin + GORM + PostgreSQL
- **前端：** React 18+ + Vite + TradingView Lightweight Charts
- **数据源：** Binance Public API
- **实时通信：** WebSocket (Gorilla WebSocket)

## 2. 功能需求

### 2.1 前端功能

#### 2.1.1 K线图展示

**需求：**
- 使用 TradingView Lightweight Charts 展示K线图
- 主K线图占据界面大部分空间
- K线图下方显示成交量柱状图
- 成交量图与K线图共享时间轴
- 颜色方案：绿色表示上涨，红色表示下跌

**交互功能：**
- 鼠标悬浮显示K线详细信息（十字线）
- 十字线跟随鼠标移动
- 价格标签显示在图表上（跟随鼠标）
- 右侧Y轴显示当前价格（红色标签高亮）
- 当前价格水平线（虚线横跨图表）

**悬浮信息显示：**
- 当前价格
- 开盘价 (Open)
- 最高价 (High)
- 最低价 (Low)
- 收盘价 (Close)
- 成交量 (Volume)
- 时间戳 (Timestamp)

#### 2.1.2 代币对选择

**需求：**
- 下拉选择框实现代币对选择
- 支持3个交易对：BTC/USDT, ETH/USDT, BNB/USDT
- 默认显示 BTC/USDT
- 切换交易对时重新加载数据

**位置：** 顶部控制栏

#### 2.1.3 时间粒度选择

**需求：**
- 按钮组实现时间粒度选择
- 支持3种时间粒度：1m, 5m, 1h
- 当前选中的时间粒度高亮显示
- 切换时间粒度时从数据库重新查询历史数据

**位置：** 顶部左侧

#### 2.1.4 当前价格显示

**需求：**
- 在右侧Y轴上显示当前价格
- 红色标签高亮显示
- 虚线横跨图表指示当前价格水平

#### 2.1.5 实时数据更新

**需求：**
- WebSocket连接实现实时数据推送
- 节流更新（每秒最多更新一次）
- 自动重连机制
- 显示连接状态提示

#### 2.1.6 界面布局

**布局结构：**
```
┌─────────────────────────────────────────┐
│ 顶部控制栏                               │
│ [时间粒度] [交易对] [K线数据]           │
├─────────────────────────────────────────┤
│                                         │
│          主K线图区域                      │
│                                         │
├─────────────────────────────────────────┤
│          成交量图区域                     │
├─────────────────────────────────────────┤
│ 底部时间轴                               │
└─────────────────────────────────────────┘
```

**样式要求：**
- 深色主题
- 界面简洁专业
- 全屏K线图布局

### 2.2 后端功能

#### 2.2.1 Binance API 接入

**需求：**
- 接入 Binance Public API 获取实时数据
- 使用 Binance WebSocket Streams 获取实时价格推送
- 使用 Binance REST API 获取历史K线数据

**支持的时间粒度：**
- 1m (1分钟)
- 5m (5分钟)
- 1h (1小时)

**支持的交易对：**
- BTC/USDT
- ETH/USDT
- BNB/USDT

#### 2.2.2 WebSocket 服务

**需求：**
- 使用 Gorilla WebSocket 实现WebSocket服务
- 实时推送K线数据更新
- 支持多个客户端连接
- 连接断开时自动清理资源

**消息格式：**
```json
{
  "type": "kline_update",
  "symbol": "BTCUSDT",
  "interval": "1m",
  "data": {
    "open_time": 1699000000000,
    "close_time": 1699000059999,
    "open": "35000.00",
    "high": "35100.00",
    "low": "34900.00",
    "close": "35050.00",
    "volume": "100.5"
  }
}
```

#### 2.2.3 RESTful API

**需求：**
- 提供历史K线数据查询接口
- 支持按交易对、时间粒度、时间范围查询
- 遵循RESTful设计原则

**API端点：**
- `GET /api/v1/klines` - 查询历史K线数据

#### 2.2.4 实时数据存储

**需求：**
- 实时接收Binance推送的K线数据
- 自动落表存储到PostgreSQL
- 为每个时间粒度分别存储数据
- 数据永久保存

**存储策略：**
- 存储所有时间粒度的K线数据（1m, 5m, 1h分别存储）
- 避免重复存储（基于时间戳去重）

## 3. 技术架构

### 3.1 系统架构

```
┌─────────────┐         ┌─────────────┐
│   Frontend  │◄───────►│   Backend   │
│   (React)   │  WS     │   (Gin)     │
└─────────────┘         └──────┬──────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
            ┌───────▼──────┐    ┌────────▼──────┐
            │  PostgreSQL  │    │  Binance API  │
            │   Database   │    │  WebSocket   │
            └──────────────┘    └──────────────┘
```

### 3.2 数据流

1. **实时数据流：**
   - Binance WebSocket → Backend → PostgreSQL (存储)
   - Binance WebSocket → Backend → Frontend (实时显示)

2. **历史数据流：**
   - Frontend → Backend → PostgreSQL → Frontend

### 3.3 后端架构

**目录结构：**
```
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   └── routes.go
│   ├── service/
│   │   ├── binance.go
│   │   └── websocket.go
│   ├── repository/
│   │   └── kline.go
│   └── models/
│       └── kline.go
├── pkg/
│   └── database/
│       └── postgres.go
└── go.mod
```

### 3.4 前端架构

**目录结构：**
```
frontend/
├── src/
│   ├── components/
│   │   ├── KLineChart.tsx
│   │   ├── SymbolSelector.tsx
│   │   ├── IntervalSelector.tsx
│   │   └── PriceDisplay.tsx
│   ├── hooks/
│   │   └── useWebSocket.ts
│   ├── services/
│   │   └── api.ts
│   ├── App.tsx
│   └── main.tsx
├── package.json
└── vite.config.ts
```

## 4. API 设计

### 4.1 RESTful API

#### 4.1.1 查询历史K线数据

**端点：** `GET /api/v1/klines`

**查询参数：**
- `symbol` (required): 交易对，如 "BTCUSDT"
- `interval` (required): 时间粒度，如 "1m", "5m", "1h"
- `start_time` (optional): 开始时间戳（毫秒）
- `end_time` (optional): 结束时间戳（毫秒）
- `limit` (optional): 返回数量限制，默认1000

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "open_time": 1699000000000,
      "close_time": 1699000059999,
      "open": "35000.00",
      "high": "35100.00",
      "low": "34900.00",
      "close": "35050.00",
      "volume": "100.5"
    }
  ]
}
```

**错误响应：**
```json
{
  "code": 400,
  "message": "invalid symbol",
  "data": null
}
```

#### 4.1.2 获取支持的交易对列表

**端点：** `GET /api/v1/symbols`

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "symbol": "BTCUSDT",
      "base_asset": "BTC",
      "quote_asset": "USDT"
    },
    {
      "symbol": "ETHUSDT",
      "base_asset": "ETH",
      "quote_asset": "USDT"
    },
    {
      "symbol": "BNBUSDT",
      "base_asset": "BNB",
      "quote_asset": "USDT"
    }
  ]
}
```

### 4.2 WebSocket API

#### 4.2.1 连接端点

**端点：** `ws://localhost:8080/ws`

#### 4.2.2 订阅消息

**客户端发送：**
```json
{
  "action": "subscribe",
  "symbol": "BTCUSDT",
  "interval": "1m"
}
```

**服务端响应：**
```json
{
  "type": "subscribed",
  "symbol": "BTCUSDT",
  "interval": "1m"
}
```

#### 4.2.3 取消订阅

**客户端发送：**
```json
{
  "action": "unsubscribe",
  "symbol": "BTCUSDT",
  "interval": "1m"
}
```

#### 4.2.4 K线数据更新

**服务端推送：**
```json
{
  "type": "kline_update",
  "symbol": "BTCUSDT",
  "interval": "1m",
  "data": {
    "open_time": 1699000000000,
    "close_time": 1699000059999,
    "open": "35000.00",
    "high": "35100.00",
    "low": "34900.00",
    "close": "35050.00",
    "volume": "100.5"
  }
}
```

## 5. 数据库设计

### 5.1 数据表设计

#### 5.1.1 K线数据表

**表名：** `klines`

**字段设计：**
```sql
CREATE TABLE klines (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open_time BIGINT NOT NULL,
    close_time BIGINT NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    high_price DECIMAL(20, 8) NOT NULL,
    low_price DECIMAL(20, 8) NOT NULL,
    close_price DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, interval, open_time)
);

CREATE INDEX idx_klines_symbol_interval_time ON klines(symbol, interval, open_time);
CREATE INDEX idx_klines_symbol_interval ON klines(symbol, interval);
```

**字段说明：**
- `id`: 主键，自增
- `symbol`: 交易对，如 "BTCUSDT"
- `interval`: 时间粒度，如 "1m", "5m", "1h"
- `open_time`: 开盘时间戳（毫秒）
- `close_time`: 收盘时间戳（毫秒）
- `open_price`: 开盘价
- `high_price`: 最高价
- `low_price`: 最低价
- `close_price`: 收盘价
- `volume`: 成交量
- `created_at`: 创建时间
- `updated_at`: 更新时间

**索引设计：**
- 主键索引：`id`
- 唯一索引：`(symbol, interval, open_time)` - 防止重复数据
- 查询索引：`(symbol, interval, open_time)` - 优化历史数据查询
- 查询索引：`(symbol, interval)` - 优化按交易对和时间粒度查询

### 5.2 数据存储策略

**存储粒度：**
- 为每个时间粒度（1m, 5m, 1h）分别存储K线数据
- 使用唯一约束防止重复存储

**数据保留：**
- 永久保存（个人使用，数据量可控）

**数据去重：**
- 基于 `(symbol, interval, open_time)` 唯一约束
- 使用 `ON CONFLICT` 或 `UPSERT` 逻辑

### 5.3 GORM 模型定义

```go
type Kline struct {
    ID         uint64    `gorm:"primaryKey;autoIncrement"`
    Symbol     string    `gorm:"type:varchar(20);not null;index:idx_symbol_interval"`
    Interval   string    `gorm:"type:varchar(10);not null;index:idx_symbol_interval"`
    OpenTime   int64     `gorm:"not null;index:idx_symbol_interval_time"`
    CloseTime  int64     `gorm:"not null"`
    OpenPrice  float64   `gorm:"type:decimal(20,8);not null"`
    HighPrice  float64   `gorm:"type:decimal(20,8);not null"`
    LowPrice   float64   `gorm:"type:decimal(20,8);not null"`
    ClosePrice float64   `gorm:"type:decimal(20,8);not null"`
    Volume     float64   `gorm:"type:decimal(20,8);not null"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

func (Kline) TableName() string {
    return "klines"
}
```

### 5.4 Docker 数据库配置

#### 5.4.1 Docker Compose 配置

**文件：** `docker-compose.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: crypto_monitor_db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: crypto_monitor
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data:
    driver: local
```

#### 5.4.2 数据库初始化脚本

**文件：** `scripts/init.sql`

```sql
-- 创建数据库（如果不存在）
-- Docker 会自动创建，这里主要用于确保

-- 创建扩展（如果需要）
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建表
CREATE TABLE IF NOT EXISTS klines (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open_time BIGINT NOT NULL,
    close_time BIGINT NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    high_price DECIMAL(20, 8) NOT NULL,
    low_price DECIMAL(20, 8) NOT NULL,
    close_price DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, interval, open_time)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval_time 
    ON klines(symbol, interval, open_time);
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval 
    ON klines(symbol, interval);
```

#### 5.4.3 Docker 管理脚本

**文件：** `scripts/db.sh`

```bash
#!/bin/bash

# 数据库 Docker 管理脚本

case "$1" in
    start)
        echo "启动 PostgreSQL 数据库..."
        docker-compose up -d postgres
        echo "等待数据库就绪..."
        sleep 5
        echo "数据库已启动"
        ;;
    stop)
        echo "停止 PostgreSQL 数据库..."
        docker-compose stop postgres
        echo "数据库已停止"
        ;;
    restart)
        echo "重启 PostgreSQL 数据库..."
        docker-compose restart postgres
        echo "数据库已重启"
        ;;
    status)
        echo "检查数据库状态..."
        docker-compose ps postgres
        ;;
    logs)
        echo "查看数据库日志..."
        docker-compose logs -f postgres
        ;;
    shell)
        echo "连接到 PostgreSQL 数据库..."
        docker-compose exec postgres psql -U postgres -d crypto_monitor
        ;;
    reset)
        echo "重置数据库（删除所有数据）..."
        read -p "确认要删除所有数据吗？(y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker-compose down -v
            docker-compose up -d postgres
            echo "数据库已重置"
        else
            echo "操作已取消"
        fi
        ;;
    backup)
        BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
        echo "备份数据库到 $BACKUP_FILE..."
        docker-compose exec -T postgres pg_dump -U postgres crypto_monitor > "$BACKUP_FILE"
        echo "备份完成: $BACKUP_FILE"
        ;;
    restore)
        if [ -z "$2" ]; then
            echo "用法: $0 restore <backup_file.sql>"
            exit 1
        fi
        echo "从 $2 恢复数据库..."
        docker-compose exec -T postgres psql -U postgres -d crypto_monitor < "$2"
        echo "恢复完成"
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|shell|reset|backup|restore}"
        echo ""
        echo "命令说明："
        echo "  start    - 启动数据库"
        echo "  stop     - 停止数据库"
        echo "  restart  - 重启数据库"
        echo "  status   - 查看数据库状态"
        echo "  logs     - 查看数据库日志"
        echo "  shell    - 连接到数据库命令行"
        echo "  reset    - 重置数据库（删除所有数据）"
        echo "  backup   - 备份数据库"
        echo "  restore  - 从备份文件恢复数据库"
        exit 1
        ;;
esac
```

#### 5.4.4 目录结构

```
.
├── docker-compose.yml
├── scripts/
│   ├── init.sql
│   └── db.sh
└── ...
```

## 6. 前端设计

### 6.1 组件设计

#### 6.1.1 KLineChart 组件

**职责：** 展示K线图和成交量图

**Props：**
```typescript
interface KLineChartProps {
  symbol: string;
  interval: string;
  data: KlineData[];
  onCrosshairMove?: (data: KlineData | null) => void;
}
```

**功能：**
- 使用 TradingView Lightweight Charts 渲染K线图
- 显示成交量柱状图
- 处理鼠标交互（十字线、价格标签）
- 实时更新数据

#### 6.1.2 SymbolSelector 组件

**职责：** 代币对选择器

**Props：**
```typescript
interface SymbolSelectorProps {
  value: string;
  onChange: (symbol: string) => void;
  symbols: string[];
}
```

**功能：**
- 下拉选择框
- 显示当前选中的交易对

#### 6.1.3 IntervalSelector 组件

**职责：** 时间粒度选择器

**Props：**
```typescript
interface IntervalSelectorProps {
  value: string;
  onChange: (interval: string) => void;
  intervals: string[];
}
```

**功能：**
- 按钮组显示
- 当前选中高亮

#### 6.1.4 PriceDisplay 组件

**职责：** 显示当前价格和K线详细信息

**Props：**
```typescript
interface PriceDisplayProps {
  currentPrice: number;
  klineData: KlineData | null;
}
```

**功能：**
- 显示当前价格
- 显示K线详细信息（O, H, L, C, Volume, Time）

### 6.2 状态管理

**使用 React Hooks：**
- `useState` - 本地状态管理
- `useEffect` - 副作用处理
- `useWebSocket` - WebSocket连接管理

**状态结构：**
```typescript
interface AppState {
  symbol: string;
  interval: string;
  klineData: KlineData[];
  currentPrice: number;
  selectedKline: KlineData | null;
  wsConnected: boolean;
  loading: boolean;
  error: string | null;
}
```

### 6.3 WebSocket Hook

**useWebSocket Hook：**
```typescript
function useWebSocket(
  url: string,
  symbol: string,
  interval: string,
  onMessage: (data: KlineData) => void
): {
  connected: boolean;
  error: string | null;
  reconnect: () => void;
}
```

**功能：**
- WebSocket连接管理
- 自动重连机制
- 节流更新（每秒最多更新一次）
- 连接状态管理

### 6.4 API Service

**API Service：**
```typescript
class ApiService {
  async getKlines(
    symbol: string,
    interval: string,
    startTime?: number,
    endTime?: number,
    limit?: number
  ): Promise<KlineData[]>;
  
  async getSymbols(): Promise<Symbol[]>;
}
```

## 7. 实现细节

### 7.1 后端实现

#### 7.1.1 Binance API 集成

**WebSocket 连接：**
- 连接 Binance WebSocket Streams
- 订阅K线数据流：`<symbol>@kline_<interval>`
- 处理消息并存储到数据库
- 转发消息到前端WebSocket客户端

**REST API 调用：**
- 获取历史K线数据：`GET /api/v3/klines`
- 错误处理和重试机制

#### 7.1.2 数据存储逻辑

**存储流程：**
1. 接收Binance推送的K线数据
2. 解析数据并转换为模型
3. 使用 `ON CONFLICT` 或 `UPSERT` 逻辑存储
4. 如果数据已存在，更新 `updated_at` 字段

**去重逻辑：**
```go
// 使用 GORM 的 FirstOrCreate 或 Clauses
db.Clauses(clause.OnConflict{
    Columns: []clause.Column{{Name: "symbol"}, {Name: "interval"}, {Name: "open_time"}},
    DoUpdates: clause.AssignmentColumns([]string{"close_price", "high_price", "low_price", "volume", "updated_at"}),
}).Create(&kline)
```

#### 7.1.3 WebSocket 服务

**连接管理：**
- 使用 Gorilla WebSocket 管理连接
- 维护客户端连接池
- 处理连接断开和清理

**消息广播：**
- 接收Binance推送的数据
- 广播到所有订阅的客户端
- 节流更新（每秒最多推送一次）

### 7.2 前端实现

#### 7.2.1 TradingView Lightweight Charts 集成

**初始化：**
```typescript
import { createChart, ColorType } from 'lightweight-charts';

const chart = createChart(containerRef.current, {
  layout: {
    background: { type: ColorType.Solid, color: '#1e1e1e' },
    textColor: '#d1d5db',
  },
  width: containerRef.current.clientWidth,
  height: 600,
});

const candlestickSeries = chart.addCandlestickSeries({
  upColor: '#26a69a',
  downColor: '#ef5350',
  borderVisible: false,
  wickUpColor: '#26a69a',
  wickDownColor: '#ef5350',
});

const volumeSeries = chart.addHistogramSeries({
  color: '#26a69a',
  priceFormat: {
    type: 'volume',
  },
  priceScaleId: 'volume',
  scaleMargins: {
    top: 0.8,
    bottom: 0,
  },
});
```

**数据更新：**
```typescript
// 更新K线数据
candlestickSeries.update({
  time: klineData.openTime,
  open: klineData.open,
  high: klineData.high,
  low: klineData.low,
  close: klineData.close,
});

// 更新成交量
volumeSeries.update({
  time: klineData.openTime,
  value: klineData.volume,
  color: klineData.close >= klineData.open ? '#26a69a' : '#ef5350',
});
```

#### 7.2.2 实时数据更新

**节流更新：**
```typescript
let lastUpdateTime = 0;
const UPDATE_INTERVAL = 1000; // 1秒

function handleWebSocketMessage(data: KlineData) {
  const now = Date.now();
  if (now - lastUpdateTime >= UPDATE_INTERVAL) {
    updateChart(data);
    lastUpdateTime = now;
  }
}
```

#### 7.2.3 自动重连

**重连逻辑：**
```typescript
function useWebSocket(url: string, onMessage: (data: any) => void) {
  const [connected, setConnected] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000;

  useEffect(() => {
    const ws = new WebSocket(url);
    
    ws.onopen = () => {
      setConnected(true);
      setReconnectAttempts(0);
    };
    
    ws.onclose = () => {
      setConnected(false);
      if (reconnectAttempts < maxReconnectAttempts) {
        setTimeout(() => {
          setReconnectAttempts(prev => prev + 1);
          // 重新连接
        }, reconnectDelay);
      }
    };
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onMessage(data);
    };
    
    return () => {
      ws.close();
    };
  }, [url, reconnectAttempts]);
}
```

## 8. 错误处理

### 8.1 后端错误处理

#### 8.1.1 API调用失败

**处理策略：**
- 显示错误提示（记录日志）
- 静默重试（指数退避）
- 记录日志

**实现：**
```go
func fetchBinanceKlines(symbol, interval string) ([]Kline, error) {
    maxRetries := 3
    retryDelay := time.Second
    
    for i := 0; i < maxRetries; i++ {
        data, err := binanceAPI.GetKlines(symbol, interval)
        if err == nil {
            return data, nil
        }
        
        log.Printf("Binance API call failed (attempt %d/%d): %v", i+1, maxRetries, err)
        if i < maxRetries-1 {
            time.Sleep(retryDelay * time.Duration(1<<i)) // 指数退避
        }
    }
    
    return nil, fmt.Errorf("failed to fetch klines after %d attempts", maxRetries)
}
```

#### 8.1.2 数据库连接失败

**处理策略：**
- 显示错误提示（记录日志）
- 继续运行（不存储数据，但实时数据仍可显示）

**实现：**
```go
func saveKline(kline *Kline) error {
    if err := db.Create(kline).Error; err != nil {
        log.Printf("Failed to save kline: %v", err)
        // 不返回错误，允许系统继续运行
        return nil
    }
    return nil
}
```

### 8.2 前端错误处理

#### 8.2.1 WebSocket连接失败

**处理：**
- 显示连接状态提示
- 自动重连（最多5次）
- 显示错误消息

#### 8.2.2 API调用失败

**处理：**
- 显示错误提示
- 记录错误日志
- 提供重试按钮

## 9. 性能优化

### 9.1 后端优化

- **数据库索引：** 在查询字段上创建索引
- **批量插入：** 使用批量插入减少数据库操作
- **连接池：** 使用数据库连接池

### 9.2 前端优化

- **节流更新：** 限制更新频率（每秒最多一次）
- **虚拟滚动：** 如果数据量大，使用虚拟滚动
- **数据缓存：** 缓存历史数据，避免重复请求

## 10. 测试要求

### 10.1 后端测试

**单元测试：**
- API接口测试
- 数据存储逻辑测试
- WebSocket服务测试

**集成测试：**
- Binance API集成测试
- 数据库操作测试

### 10.2 前端测试

**组件测试：**
- K线图组件渲染测试
- 交互功能测试

**E2E测试：**
- 完整流程测试（暂不实现，快速开发优先）

## 11. 部署

### 11.1 本地开发

**后端：**
```bash
go run cmd/server/main.go
```

**前端：**
```bash
npm run dev
```

**数据库：**
- 使用 Docker Compose 启动 PostgreSQL
```bash
# 方式1：使用管理脚本
chmod +x scripts/db.sh
./scripts/db.sh start

# 方式2：直接使用 docker-compose
docker-compose up -d postgres
```

### 11.2 Docker 数据库使用

#### 11.2.1 快速启动

**首次启动：**
```bash
# 启动数据库（会自动创建表结构）
docker-compose up -d postgres

# 查看日志确认启动成功
docker-compose logs postgres
```

**使用管理脚本：**
```bash
# 给脚本添加执行权限（首次使用）
chmod +x scripts/db.sh

# 启动数据库
./scripts/db.sh start

# 查看状态
./scripts/db.sh status

# 查看日志
./scripts/db.sh logs

# 连接到数据库
./scripts/db.sh shell
```

#### 11.2.2 数据库操作

**备份数据库：**
```bash
./scripts/db.sh backup
# 或手动备份
docker-compose exec -T postgres pg_dump -U postgres crypto_monitor > backup.sql
```

**恢复数据库：**
```bash
./scripts/db.sh restore backup.sql
# 或手动恢复
docker-compose exec -T postgres psql -U postgres -d crypto_monitor < backup.sql
```

**重置数据库：**
```bash
./scripts/db.sh reset
# 或手动重置
docker-compose down -v
docker-compose up -d postgres
```

#### 11.2.3 环境变量配置

可以通过环境变量或 `.env` 文件自定义数据库配置：

**`.env` 文件示例：**
```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=crypto_monitor
POSTGRES_PORT=5432
```

**更新 `docker-compose.yml` 使用环境变量：**
```yaml
environment:
  POSTGRES_USER: ${POSTGRES_USER:-postgres}
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres}
  POSTGRES_DB: ${POSTGRES_DB:-crypto_monitor}
```

### 11.3 环境配置

**后端环境变量：**
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=crypto_monitor
BINANCE_WS_URL=wss://stream.binance.com:9443/ws
BINANCE_API_URL=https://api.binance.com
```

**前端环境变量：**
```
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
```

## 12. 开发优先级

### 12.1 P0 - 必须实现

- ✅ 后端API接入Binance
- ✅ 数据库设计和迁移
- ✅ WebSocket实时推送
- ✅ 实时数据落表存储
- ✅ 前端K线图展示
- ✅ 代币对选择
- ✅ 时间粒度选择
- ✅ 鼠标悬浮显示价格
- ✅ 历史数据查询接口

### 12.2 P1 - 有时间就做

- ⏳ 错误提示UI
- ⏳ 加载状态提示
- ⏳ 连接状态提示

### 12.3 P2 - 后续考虑

- ❌ 更多交易对
- ❌ 更多时间粒度
- ❌ 技术指标
- ❌ 响应式布局

## 13. 参考资料

- [Binance API 文档](https://binance-docs.github.io/apidocs/spot/en/)
- [TradingView Lightweight Charts 文档](https://tradingview.github.io/lightweight-charts/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)

