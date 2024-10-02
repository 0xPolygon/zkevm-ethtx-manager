-- +migrate Up
CREATE TABLE IF NOT EXISTS monitored_txs (
    id CHAR(66) PRIMARY KEY,
    from_address CHAR(42) NOT NULL,
    to_address CHAR(42),
    nonce BIGINT NOT NULL,
    "value" TEXT,                  -- *big.Int
    tx_data BLOB,
    gas BIGINT NOT NULL,
    gas_offset BIGINT,
    gas_price TEXT,                -- *big.Int
    blob_sidecar BLOB,
    blob_gas BIGINT,
    blob_gas_price TEXT,           -- *big.Int
    gas_tip_cap TEXT,              -- *big.Int
    "status" TEXT NOT NULL,
    block_number BIGINT,
    history JSONB,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    estimate_gas INTEGER NOT NULL  -- 0 = FALSE, 1 = TRUE
);

-- Indexes
CREATE INDEX idx_monitored_txs_status ON monitored_txs("status");
CREATE INDEX idx_monitored_txs_created_at ON monitored_txs(created_at);
CREATE INDEX idx_monitored_txs_block_number ON monitored_txs(block_number);
CREATE INDEX idx_monitored_txs_status_created_at ON monitored_txs("status", created_at);

-- +migrate Down
DROP TABLE IF EXISTS monitored_txs;
