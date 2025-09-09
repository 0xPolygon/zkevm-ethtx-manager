-- +migrate Up
ALTER TABLE monitored_txs ADD COLUMN retry_count BIGINT DEFAULT 0 NOT NULL;

-- Add index for retry_count to optimize queries
CREATE INDEX idx_monitored_txs_retry_count ON monitored_txs(retry_count);

-- +migrate Down
DROP INDEX IF EXISTS idx_monitored_txs_retry_count;
ALTER TABLE monitored_txs DROP COLUMN retry_count;
