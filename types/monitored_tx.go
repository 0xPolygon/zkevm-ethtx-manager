package types

import (
	"database/sql"
	"math/big"
	"time"

	localCommon "github.com/0xPolygon/zkevm-ethtx-manager/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

const (
	// MonitoredTxStatusCreated means the tx was just added to the storage
	MonitoredTxStatusCreated = MonitoredTxStatus("created")

	// MonitoredTxStatusSent means that at least a eth tx was sent to the network
	MonitoredTxStatusSent = MonitoredTxStatus("sent")

	// MonitoredTxStatusFailed means the tx was already mined and failed with an
	// error that can't be recovered automatically, ex: the data in the tx is invalid
	// and the tx gets reverted
	MonitoredTxStatusFailed = MonitoredTxStatus("failed")

	// MonitoredTxStatusMined means the tx was already mined and the receipt
	// status is Successful
	MonitoredTxStatusMined = MonitoredTxStatus("mined")

	// MonitoredTxStatusSafe means the tx was already mined N blocks ago
	MonitoredTxStatusSafe = MonitoredTxStatus("safe")

	// MonitoredTxStatusFinalized means the tx was already mined M (M > N) blocks ago
	MonitoredTxStatusFinalized = MonitoredTxStatus("finalized")
)

// MonitoredTxStatus represents the status of a monitored tx
type MonitoredTxStatus string

// String returns a string representation of the status
func (s MonitoredTxStatus) String() string {
	return string(s)
}

// MonitoredTx represents a set of information used to build tx
// plus information to monitor if the transactions was sent successfully
type MonitoredTx struct {
	// ID is the tx identifier controlled by the caller
	ID common.Hash `mapstructure:"id" meddler:"id,hash"`

	// From is the sender of the tx, used to identify which private key should be used to sign the tx
	From common.Address `mapstructure:"from" meddler:"from_address,address"`

	// To is the receiver of the tx
	To *common.Address `mapstructure:"to" meddler:"to_address,address"`

	// Nonce is used to create the tx
	Nonce uint64 `mapstructure:"nonce" meddler:"nonce"`

	// Value is the transaction value
	Value *big.Int `mapstructure:"value" meddler:"value,bigInt"`

	// Data represents the transaction data
	Data []byte `mapstructure:"data" meddler:"tx_data"`

	// Gas is the amount of gas for the transaction
	Gas uint64 `mapstructure:"gas" meddler:"gas"`

	// GasOffset is the offset applied to the gas amount
	GasOffset uint64 `mapstructure:"gasOffset" meddler:"gas_offset"`

	// GasPrice is the price per gas unit for the transaction
	GasPrice *big.Int `mapstructure:"gasPrice" meddler:"gas_price,bigInt"`

	// BlobSidecar holds sidecar data for blob transactions
	BlobSidecar *types.BlobTxSidecar `mapstructure:"blobSidecar" meddler:"blob_sidecar,json"`

	// BlobGas is the gas amount for the blob transaction
	BlobGas uint64 `mapstructure:"blobGas" meddler:"blob_gas"`

	// BlobGasPrice is the gas price for blob transactions
	BlobGasPrice *big.Int `mapstructure:"blobGasPrice" meddler:"blob_gas_price,bigInt"`

	// GasTipCap is the tip cap for the gas fee
	GasTipCap *big.Int `mapstructure:"gasTipCap" meddler:"gas_tip_cap,bigInt"`

	// Status represents the status of this monitored transaction
	Status MonitoredTxStatus `mapstructure:"status" meddler:"status"`

	// BlockNumber represents the block where the transaction was identified to be mined
	// This is used to control reorged monitored txs.
	BlockNumber *big.Int `mapstructure:"blockNumber" meddler:"block_number,bigInt"`

	// History represents all transaction hashes created using this struct and sent to the network
	History map[common.Hash]bool `mapstructure:"history" meddler:"history,json"`

	// CreatedAt is the timestamp for when the transaction was created
	CreatedAt time.Time `mapstructure:"createdAt" meddler:"created_at,timeRFC3339"`

	// UpdatedAt is the timestamp for when the transaction was last updated
	UpdatedAt time.Time `mapstructure:"updatedAt" meddler:"updated_at,timeRFC3339"`

	// EstimateGas indicates whether gas should be estimated or the last value should be reused
	EstimateGas bool `mapstructure:"estimateGas" meddler:"estimate_gas"`
}

// Tx uses the current information to build a tx
func (mTx MonitoredTx) Tx() *types.Transaction {
	var tx *types.Transaction
	if mTx.BlobSidecar == nil {
		tx = types.NewTx(&types.LegacyTx{
			To:       mTx.To,
			Nonce:    mTx.Nonce,
			Value:    mTx.Value,
			Data:     mTx.Data,
			Gas:      mTx.Gas + mTx.GasOffset,
			GasPrice: mTx.GasPrice,
		})
	} else {
		tx = types.NewTx(&types.BlobTx{
			To:         *mTx.To,
			Nonce:      mTx.Nonce,
			Value:      uint256.MustFromBig(mTx.Value),
			Data:       mTx.Data,
			GasFeeCap:  uint256.MustFromBig(mTx.GasPrice),
			GasTipCap:  uint256.MustFromBig(mTx.GasTipCap),
			Gas:        mTx.Gas + mTx.GasOffset,
			BlobFeeCap: uint256.MustFromBig(mTx.BlobGasPrice),
			BlobHashes: mTx.BlobSidecar.BlobHashes(),
			Sidecar:    mTx.BlobSidecar,
		})
	}

	return tx
}

// AddHistory adds a transaction to the monitoring history
func (mTx MonitoredTx) AddHistory(tx *types.Transaction) error {
	if _, found := mTx.History[tx.Hash()]; found {
		return ErrAlreadyExists
	}
	mTx.History[tx.Hash()] = true
	return nil
}

// HistoryHashSlice returns the current history field as a string slice
func (mTx *MonitoredTx) HistoryHashSlice() []common.Hash {
	history := make([]common.Hash, 0, len(mTx.History))
	for h := range mTx.History {
		history = append(history, h)
	}
	return history
}

// PopulateNullableStrings converts the nullable strings and populates them to MonitoredTx instance
func (mTx *MonitoredTx) PopulateNullableStrings(toAddress, blockNumber, value, gasPrice,
	blobGasPrice, gasTipCap sql.NullString) {
	if toAddress.Valid {
		addr := common.HexToAddress(toAddress.String)
		mTx.To = &addr
	}

	if blockNumber.Valid {
		mTx.BlockNumber, _ = new(big.Int).SetString(blockNumber.String, localCommon.Base10)
	}

	if value.Valid {
		mTx.Value, _ = new(big.Int).SetString(value.String, localCommon.Base10)
	}

	if gasPrice.Valid {
		mTx.GasPrice, _ = new(big.Int).SetString(gasPrice.String, localCommon.Base10)
	}

	if blobGasPrice.Valid {
		mTx.BlobGasPrice, _ = new(big.Int).SetString(blobGasPrice.String, localCommon.Base10)
	}

	if gasTipCap.Valid {
		mTx.GasTipCap, _ = new(big.Int).SetString(gasTipCap.String, localCommon.Base10)
	}
}

// MonitoredTxResult represents the result of a execution of a monitored tx
type MonitoredTxResult struct {
	ID                 common.Hash
	To                 *common.Address
	Nonce              uint64
	Value              *big.Int
	Data               []byte
	MinedAtBlockNumber *big.Int
	Status             MonitoredTxStatus
	Txs                map[common.Hash]TxResult
}

// TxResult represents the result of a execution of a ethereum transaction in the block chain
type TxResult struct {
	Tx            *types.Transaction
	Receipt       *types.Receipt
	RevertMessage string
}
