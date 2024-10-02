package ethtxmanager

import (
	"context"

	"github.com/0xPolygon/zkevm-ethtx-manager/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type monitoredTxnIteration struct {
	*types.MonitoredTx
	confirmed   bool
	lastReceipt *ethtypes.Receipt
}

func (m *monitoredTxnIteration) shouldUpdateNonce(ctx context.Context, etherman types.EthermanInterface) bool {
	if m.Status == types.MonitoredTxStatusCreated {
		// transaction was not sent, so no need to check if it was mined
		// we need to update the nonce in this case
		return true
	}

	// check if any of the txs in the history was confirmed
	var lastReceiptChecked *ethtypes.Receipt
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
		if lastReceiptChecked.Status == ethtypes.ReceiptStatusSuccessful {
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
