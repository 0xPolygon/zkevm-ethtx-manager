-- +migrate Up
CREATE SCHEMA IF NOT EXISTS tx_manager;

CREATE TABLE IF NOT EXISTS tx_manager.monitored_txs (
    id CHAR(66) PRIMARY KEY,
    from_address CHAR(42) NOT NULL,
    to_address CHAR(42),
    nonce BIGINT NOT NULL,
    "value" NUMERIC,
    tx_data BLOB,
    gas BIGINT NOT NULL,
    gas_offset BIGINT,
    gas_price NUMERIC,
    blob_sidecar BLOB,
    blob_gas BIGINT,
    blob_gas_price NUMERIC,
    gas_tip_cap NUMERIC,
    "status" INT NOT NULL,
    block_number NUMERIC,
    history TEXT,                  -- JSON encoded string
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    estimate_gas INTEGER NOT NULL  -- 0 = FALSE, 1 = TRUE
);

-- Indexes
CREATE INDEX idx_monitored_txs_status ON tx_manager.monitored_txs("status");
CREATE INDEX idx_monitored_txs_created_at ON tx_manager.monitored_txs(created_at);
CREATE INDEX idx_monitored_txs_block_number ON tx_manager.monitored_txs(block_number);
CREATE INDEX idx_monitored_txs_status_created_at ON tx_manager.monitored_txs("status", created_at);

-- +migrate Down
DROP TABLE IF EXISTS monitored_txs;

DROP SCHEMA IF EXISTS tx_manager;