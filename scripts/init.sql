-- 数据库初始化脚本
-- 此脚本在 Docker 容器首次启动时自动执行
-- 用于创建表结构和索引

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

-- 添加注释
COMMENT ON TABLE klines IS 'Stores cryptocurrency candlestick/K-line data';
COMMENT ON COLUMN klines.symbol IS 'Trading pair symbol (e.g., BTCUSDT)';
COMMENT ON COLUMN klines.interval IS 'Time interval (e.g., 1m, 5m, 1h)';
COMMENT ON COLUMN klines.open_time IS 'Opening time timestamp in milliseconds';
COMMENT ON COLUMN klines.close_time IS 'Closing time timestamp in milliseconds';

