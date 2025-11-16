-- Rollback migration: Drop klines table
-- Created: 2025-11-07
-- Description: Drops the klines table and all related indexes

-- Drop indexes first
DROP INDEX IF EXISTS idx_klines_symbol_interval_time;
DROP INDEX IF EXISTS idx_klines_symbol_interval;

-- Drop table
DROP TABLE IF EXISTS klines;

