-- +migrate Up
CREATE SCHEMA IF NOT EXISTS tx_manager;

CREATE TABLE IF NOT EXISTS tx_manager.monitored_txs (
    id CHAR(66) PRIMARY KEY,                     -- Corresponds to common.Hash (Tx identifier)
    from_address CHAR(42) NOT NULL,              -- Corresponds to common.Address (Sender of the tx)
    to_address CHAR(42),                         -- Corresponds to *common.Address (Receiver of the tx)
    nonce BIGINT NOT NULL,                       -- Corresponds to uint64 (Nonce used to create the tx)
    "value" NUMERIC,                             -- Corresponds to *big.Int (Tx value)
    tx_data BYTEA,                               -- Corresponds to []byte (Tx data)
    gas BIGINT NOT NULL,                         -- Corresponds to uint64 (Tx gas)
    gas_offset BIGINT,                           -- Corresponds to uint64 (Tx gas offset)
    gas_price NUMERIC,                           -- Corresponds to *big.Int (Tx gas price)
    blob_sidecar BYTEA,                          -- Corresponds to *types.BlobTxSidecar (Blob sidecar data)
    blob_gas BIGINT,                             -- Corresponds to uint64 (Blob gas)
    blob_gas_price NUMERIC,                      -- Corresponds to *big.Int (Blob gas price)
    gas_tip_cap NUMERIC,                         -- Corresponds to *big.Int (Gas tip cap)
    "status" INT NOT NULL,                       -- Corresponds to MonitoredTxStatus (Status of monitoring)
    block_number NUMERIC,                        -- Corresponds to *big.Int (Block number where the tx was mined)
    history JSONB,                               -- Corresponds to map[common.Hash]bool (History of transaction hashes)
    created_at TIMESTAMPTZ NOT NULL,             -- Corresponds to time.Time (Time of creation)
    updated_at TIMESTAMPTZ NOT NULL,             -- Corresponds to time.Time (Last update time)
    estimate_gas BOOLEAN NOT NULL                -- Corresponds to bool (Whether to estimate gas)
);

-- status column index
CREATE INDEX idx_monitored_txs_status ON tx_manager.monitored_txs("status");
-- created_at column index
CREATE INDEX idx_monitored_txs_created_at ON tx_manager.monitored_txs(created_at);
-- block_number column index
CREATE INDEX idx_monitored_txs_block_number ON tx_manager.monitored_txs(block_number);
-- status and created_at composite index
CREATE INDEX idx_monitored_txs_status_created_at ON tx_manager.monitored_txs("status", created_at);

-- +migrate Down
DROP TABLE IF EXISTS monitored_txs;

DROP SCHEMA IF EXISTS tx_manager;