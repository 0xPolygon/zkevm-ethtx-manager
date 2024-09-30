package ethtxmanager

import (
	"context"
	"math/big"
	"time"

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

// monitoredTx represents a set of information used to build tx
// plus information to monitor if the transactions was sent successfully
type monitoredTx struct {
	// ID is the tx identifier controller by the caller
	ID common.Hash `mapstructure:"id"`

	// sender of the tx, used to identify which private key should be used to sing the tx
	From common.Address `mapstructure:"from"`

	// receiver of the tx
	To *common.Address `mapstructure:"to"`

	// Nonce used to create the tx
	Nonce uint64 `mapstructure:"nonce"`

	// tx Value
	Value *big.Int `mapstructure:"value"`

	// tx Data
	Data []byte `mapstructure:"data"`

	// tx Gas
	Gas uint64 `mapstructure:"gas"`

	// tx gas offset
	GasOffset uint64 `mapstructure:"gasOffset"`

	// tx gas price
	GasPrice *big.Int `mapstructure:"gasPrice"`

	// blob Sidecar
	BlobSidecar *types.BlobTxSidecar `mapstructure:"blobSidecar"`

	// blob Gas
	BlobGas uint64 `mapstructure:"blobGas"`

	// blob gas price
	BlobGasPrice *big.Int `mapstructure:"blobGasPrice"`

	// gas tip cap
	GasTipCap *big.Int `mapstructure:"gasTipCap"`

	// Status of this monitoring
	Status MonitoredTxStatus `mapstructure:"status"`

	// BlockNumber represents the block where the tx was identified
	// to be mined, it's the same as the block number found in the
	// tx receipt, this is used to control reorged monitored txs
	BlockNumber *big.Int `mapstructure:"blockNumber"`

	// History represent all transaction hashes from
	// transactions created using this struct data and
	// sent to the network
	History map[common.Hash]bool `mapstructure:"history"`

	// CreatedAt date time it was created
	CreatedAt time.Time `mapstructure:"createdAt"`

	// UpdatedAt last date time it was updated
	UpdatedAt time.Time `mapstructure:"updatedAt"`

	// EstimateGas indicates if gas should be estimated or last value shold be reused
	EstimateGas bool `mapstructure:"estimateGas"`
}

// Tx uses the current information to build a tx
func (mTx monitoredTx) Tx() *types.Transaction {
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
func (mTx monitoredTx) AddHistory(tx *types.Transaction) error {
	if _, found := mTx.History[tx.Hash()]; found {
		return ErrAlreadyExists
	}
	mTx.History[tx.Hash()] = true
	return nil
}

// historyHashSlice returns the current history field as a string slice
func (mTx *monitoredTx) historyHashSlice() []common.Hash {
	history := make([]common.Hash, 0, len(mTx.History))
	for h := range mTx.History {
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

type monitoredTxnIteration struct {
	*monitoredTx
	confirmed   bool
	lastReceipt *types.Receipt
}

func (m *monitoredTxnIteration) shouldUpdateNonce(ctx context.Context, etherman EthermanInterface) bool {
	if m.Status == MonitoredTxStatusCreated {
		// transaction was not sent, so no need to check if it was mined
		// we need to update the nonce in this case
		return true
	}

	// check if any of the txs in the history was confirmed
	var lastReceiptChecked *types.Receipt
	// monitored tx is confirmed until we find a successful receipt
	confirmed := false
	// monitored tx doesn't have a failed receipt until we find a failed receipt for any
	// tx in the monitored tx history
	hasFailedReceipts := false
	// all history txs are considered mined until we can't find a receipt for any
	// tx in the monitored tx history
	allHistoryTxsWereMined := true
	for txHash := range m.History {
		mined, receipt, err := etherman.CheckTxWasMined(ctx, txHash)
		if err != nil {
			continue
		}

		// if the tx is not mined yet, check that not all the tx were mined and go to the next
		if !mined {
			allHistoryTxsWereMined = false
			continue
		}

		lastReceiptChecked = receipt

		// if the tx was mined successfully we can set it as confirmed and break the loop
		if lastReceiptChecked.Status == types.ReceiptStatusSuccessful {
			confirmed = true
			break
		}

		// if the tx was mined but failed, we continue to consider it was not confirmed
		// and set that we have found a failed receipt. This info will be used later
		// to check if nonce needs to be reviewed
		confirmed = false
		hasFailedReceipts = true
	}

	m.confirmed = confirmed
	m.lastReceipt = lastReceiptChecked

	// we need to check if we need to review the nonce carefully, to avoid sending
	// duplicated data to the roll-up and causing an unnecessary trusted state reorg.
	//
	// if we have failed receipts, this means at least one of the generated txs was mined,
	// in this case maybe the current nonce was already consumed(if this is the first iteration
	// of this cycle, next iteration might have the nonce already updated by the preivous one),
	// then we need to check if there are tx that were not mined yet, if so, we just need to wait
	// because maybe one of them will get mined successfully
	//
	// in case of the monitored tx is not confirmed yet, all tx were mined and none of them were
	// mined successfully, we need to review the nonce
	return !confirmed && hasFailedReceipts && allHistoryTxsWereMined
}
