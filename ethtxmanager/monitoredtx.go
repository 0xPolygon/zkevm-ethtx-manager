package ethtxmanager

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	// MonitoredTxStatusFinalized means the tx was set by the owner as finalized
	MonitoredTxStatusFinalized = MonitoredTxStatus("finalized")
)

// MonitoredTxStatus represents the status of a monitored tx
type MonitoredTxStatus string

// String returns a string representation of the status
func (s MonitoredTxStatus) String() string {
	return string(s)
}

// monitoredTx represents a set of information used to build tx
// plus information to monitor if the transactions was sent successfully
type monitoredTx struct {
	// id is the tx identifier controller by the caller
	id common.Hash

	// sender of the tx, used to identify which private key should be used to sing the tx
	from common.Address

	// receiver of the tx
	to *common.Address

	// nonce used to create the tx
	nonce uint64

	// tx value
	value *big.Int

	// tx data
	data []byte

	// tx gas
	gas uint64

	// tx gas offset
	gasOffset uint64

	// tx gas price
	gasPrice *big.Int

	// status of this monitoring
	status MonitoredTxStatus

	// blockNumber represents the block where the tx was identified
	// to be mined, it's the same as the block number found in the
	// tx receipt, this is used to control reorged monitored txs
	blockNumber *big.Int

	// history represent all transaction hashes from
	// transactions created using this struct data and
	// sent to the network
	history map[common.Hash]bool

	// createdAt date time it was created
	createdAt time.Time

	// updatedAt last date time it was updated
	updatedAt time.Time
}

// Tx uses the current information to build a tx
func (mTx monitoredTx) Tx() *types.Transaction {
	tx := types.NewTx(&types.LegacyTx{
		To:       mTx.to,
		Nonce:    mTx.nonce,
		Value:    mTx.value,
		Data:     mTx.data,
		Gas:      mTx.gas + mTx.gasOffset,
		GasPrice: mTx.gasPrice,
	})

	return tx
}

// AddHistory adds a transaction to the monitoring history
func (mTx monitoredTx) AddHistory(tx *types.Transaction) error {
	if _, found := mTx.history[tx.Hash()]; found {
		return ErrAlreadyExists
	}
	mTx.history[tx.Hash()] = true
	return nil
}

// historyHashSlice returns the current history field as a string slice
func (mTx *monitoredTx) historyHashSlice() []common.Hash {
	history := make([]common.Hash, 0, len(mTx.history))
	for h := range mTx.history {
		history = append(history, h)
	}
	return history
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
