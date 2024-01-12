-- +migrate Down
DROP SCHEMA IF EXISTS ethtxmgr CASCADE;

-- +migrate Up
CREATE SCHEMA ethtxmgr;

CREATE TABLE ethtxmgr.monitored_txs
(
    owner      VARCHAR NOT NULL,
    id         VARCHAR NOT NULL,
    from_addr  VARCHAR NOT NULL,
    to_addr    VARCHAR,
    nonce      DECIMAL(78, 0) NOT NULL,
    value      DECIMAL(78, 0),
    data       VARCHAR,
    gas        DECIMAL(78, 0) NOT NULL,
    gas_price  DECIMAL(78, 0) NOT NULL,
    status     VARCHAR NOT NULL,
    history    VARCHAR[],
    block_num  BIGINT,
    gas_offset DECIMAL(78, 0) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (owner, id)
);
