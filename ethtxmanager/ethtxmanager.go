// Package ethtxmanager handles ethereum transactions:  It makes
// calls to send and to aggregate batch, checks possible errors, like wrong nonce or gas limit too low
// and make correct adjustments to request according to it. Also, it tracks transaction receipt and status
// of tx in case tx is rejected and send signals to sequencer/aggregator to resend sequence/batch
package ethtxmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xPolygonHermez/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const failureIntervalInSeconds = 5

var (
	// ErrNotFound when the object is not found
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists when the object already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrExecutionReverted returned when trying to get the revert message
	// but the call fails without revealing the revert reason
	ErrExecutionReverted = errors.New("execution reverted")
)

// Client for eth tx manager
type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg      Config
	etherman ethermanInterface
	storage  storageInterface
}

type pending struct {
	Pending map[common.Address]map[uint64]l1Tx `json:"pending"`
}

type l1Tx struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Nonce    string `json:"nonce"`
	GasPrice string `json:"gasPrice"`
	Gas      string `json:"gas"`
	Value    string `json:"value"`
	Data     string `json:"input"`
}

// New creates new eth tx manager
// func New(cfg Config, ethMan ethermanInterface, storage storageInterface, state stateInterface) *Client {
func New(cfg Config) (*Client, error) {
	etherman, err := etherman.NewClient(cfg.Etherman)
	if err != nil {
		return nil, err
	}

	auth, err := etherman.LoadAuthFromKeyStore(cfg.PrivateKeys[0].Path, cfg.PrivateKeys[0].Password)
	if err != nil {
		return nil, err
	}

	err = etherman.AddOrReplaceAuth(*auth)
	if err != nil {
		return nil, err
	}

	client := Client{
		cfg:      cfg,
		etherman: etherman,
		storage:  NewMemStorage(),
	}

	return &client, nil
}

func pendingL1Txs(URL string, from common.Address, httpHeaders map[string]string) ([]monitoredTx, error) {
	response, err := JSONRPCCall(URL, "txpool_content", httpHeaders)
	if err != nil {
		return nil, err
	}

	var L1Txs pending
	err = json.Unmarshal(response.Result, &L1Txs)
	if err != nil {
		return nil, err
	}

	var mTxs []monitoredTx
	for _, tx := range L1Txs.Pending[from] {
		if common.HexToAddress(tx.From) == from {
			to := common.HexToAddress(tx.To)
			nonce, ok := new(big.Int).SetString(tx.Nonce, 0)
			if !ok {
				return nil, fmt.Errorf("failed to convert nonce %v to big.Int", tx.Nonce)
			}

			value, ok := new(big.Int).SetString(tx.Value, 0)
			if !ok {
				return nil, fmt.Errorf("failed to convert value %v to big.Int", tx.Value)
			}

			gas, ok := new(big.Int).SetString(tx.Gas, 0)
			if !ok {
				return nil, fmt.Errorf("failed to convert gas %v to big.Int", tx.Gas)
			}

			gasPrice, ok := new(big.Int).SetString(tx.GasPrice, 0)
			if !ok {
				return nil, fmt.Errorf("failed to convert gasPrice %v to big.Int", tx.GasPrice)
			}

			data := common.Hex2Bytes(string(tx.Data))

			mTx := monitoredTx{
				id:       types.NewTx(&types.LegacyTx{To: &to, Nonce: nonce.Uint64(), Value: value, Data: data}).Hash(),
				from:     common.HexToAddress(tx.From),
				to:       &to,
				nonce:    nonce.Uint64(),
				value:    value,
				data:     data,
				gas:      gas.Uint64(),
				gasPrice: gasPrice,
				status:   MonitoredTxStatusSent,
				history:  make(map[common.Hash]bool),
			}
			mTxs = append(mTxs, mTx)
		}
	}

	return mTxs, nil
}

// Add a transaction to be sent and monitored
func (c *Client) Add(ctx context.Context, to *common.Address, forcedNonce *uint64, value *big.Int, data []byte) (common.Hash, error) {
	var nonce uint64
	var err error
	var from = c.cfg.From

	if forcedNonce == nil {
		// get next nonce
		nonce, err = c.etherman.CurrentNonce(ctx, from)
		if err != nil {
			err := fmt.Errorf("failed to get current nonce: %w", err)
			log.Errorf(err.Error())
			return common.Hash{}, err
		}
	} else {
		nonce = *forcedNonce
	}

	// get gas
	gas, err := c.etherman.EstimateGas(ctx, from, to, value, data)
	if err != nil {
		err := fmt.Errorf("failed to estimate gas: %w, data: %v", err, common.Bytes2Hex(data))
		log.Error(err.Error())
		log.Debugf("failed to estimate gas for tx: from: %v, to: %v, value: %v", from.String(), to.String(), value.String())
		if c.cfg.ForcedGas > 0 {
			gas = c.cfg.ForcedGas
		} else {
			return common.Hash{}, err
		}
	}

	// get gas price
	gasPrice, err := c.suggestedGasPrice(ctx)
	if err != nil {
		err := fmt.Errorf("failed to get suggested gas price: %w", err)
		log.Errorf(err.Error())
		return common.Hash{}, err
	}

	// Calculate id
	tx := types.NewTx(&types.LegacyTx{
		To:    to,
		Nonce: nonce,
		Value: value,
		Data:  data,
	})

	id := tx.Hash()

	// create monitored tx
	mTx := monitoredTx{
		id: id, from: from, to: to,
		nonce: nonce, value: value, data: data,
		gas: gas, gasPrice: gasPrice,
		status:  MonitoredTxStatusCreated,
		history: make(map[common.Hash]bool),
	}

	// add to storage
	err = c.storage.Add(ctx, mTx)
	if err != nil {
		err := fmt.Errorf("failed to add tx to get monitored: %w", err)
		log.Errorf(err.Error())
		return common.Hash{}, err
	}

	mTxLog := log.WithFields("monitoredTx", mTx.id, "createdAt", mTx.createdAt)
	mTxLog.Infof("created")

	return id, nil
}

// ResultsByStatus returns all the results for all the monitored txs matching the provided statuses
// if the statuses are empty, all the statuses are considered.
func (c *Client) ResultsByStatus(ctx context.Context, statuses []MonitoredTxStatus) ([]MonitoredTxResult, error) {
	mTxs, err := c.storage.GetByStatus(ctx, statuses)
	if err != nil {
		return nil, err
	}

	results := make([]MonitoredTxResult, 0, len(mTxs))

	for _, mTx := range mTxs {
		result, err := c.buildResult(ctx, mTx)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// Result returns the current result of the transaction execution with all the details
func (c *Client) Result(ctx context.Context, id common.Hash) (MonitoredTxResult, error) {
	mTx, err := c.storage.Get(ctx, id)
	if err != nil {
		return MonitoredTxResult{}, err
	}

	return c.buildResult(ctx, mTx)
}

// SetStatusDone sets the status of a monitored tx to MonitoredStatusDone.
// this method is provided to the callers to decide when a monitored tx should be
// considered done, so they can start to ignore it when querying it by Status.
func (c *Client) setStatusFinalized(ctx context.Context, id common.Hash) error {
	mTx, err := c.storage.Get(ctx, id)
	if err != nil {
		return err
	}
	mTx.status = MonitoredTxStatusFinalized
	return c.storage.Update(ctx, mTx)
}

func (c *Client) buildResult(ctx context.Context, mTx monitoredTx) (MonitoredTxResult, error) {
	history := mTx.historyHashSlice()
	txs := make(map[common.Hash]TxResult, len(history))

	for _, txHash := range history {
		tx, _, err := c.etherman.GetTx(ctx, txHash)
		if !errors.Is(err, ethereum.NotFound) && err != nil {
			return MonitoredTxResult{}, err
		}

		receipt, err := c.etherman.GetTxReceipt(ctx, txHash)
		if !errors.Is(err, ethereum.NotFound) && err != nil {
			return MonitoredTxResult{}, err
		}

		revertMessage, err := c.etherman.GetRevertMessage(ctx, tx)
		if !errors.Is(err, ethereum.NotFound) && err != nil && err.Error() != ErrExecutionReverted.Error() {
			return MonitoredTxResult{}, err
		}

		txs[txHash] = TxResult{
			Tx:            tx,
			Receipt:       receipt,
			RevertMessage: revertMessage,
		}
	}

	result := MonitoredTxResult{
		ID:                 mTx.id,
		To:                 mTx.to,
		Nonce:              mTx.nonce,
		Value:              mTx.value,
		Data:               mTx.data,
		MinedAtBlockNumber: mTx.blockNumber,
		Status:             mTx.status,
		Txs:                txs,
	}

	return result, nil
}

// Start will start the tx management, reading txs from storage,
// send then to the blockchain and keep monitoring them until they
// get mined
func (c *Client) Start() {
	// Check L1 for pending txs
	pendingTxs, err := pendingL1Txs(c.cfg.Etherman.URL, c.cfg.From, c.cfg.Etherman.HTTPHeaders)
	if err != nil {
		log.Errorf("failed to get pending txs from L1: %v", err)
	}

	log.Infof("%d L1 pending Txs found", len(pendingTxs))

	for _, mTx := range pendingTxs {
		err := c.storage.Add(context.Background(), mTx)
		if err != nil {
			log.Errorf("failed to add pending tx to storage: %v", err)
		}
	}

	// infinite loop to manage txs as they arrive
	c.ctx, c.cancel = context.WithCancel(context.Background())

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(c.cfg.FrequencyToMonitorTxs.Duration):
			err := c.monitorTxs(context.Background())
			if err != nil {
				c.logErrorAndWait("failed to monitor txs: %v", err)
			}
		}
	}
}

// Stop will stops the monitored tx management
func (c *Client) Stop() {
	c.cancel()
}

// monitorTxs process all pending monitored tx
func (c *Client) monitorTxs(ctx context.Context) error {
	statusesFilter := []MonitoredTxStatus{MonitoredTxStatusCreated, MonitoredTxStatusSent, MonitoredTxStatusMined}
	mTxs, err := c.storage.GetByStatus(ctx, statusesFilter)
	if err != nil {
		return fmt.Errorf("failed to get created monitored txs: %v", err)
	}

	log.Infof("found %v monitored tx to process", len(mTxs))

	wg := sync.WaitGroup{}
	wg.Add(len(mTxs))
	for _, mTx := range mTxs {
		mTx := mTx // force variable shadowing to avoid pointer conflicts
		go func(c *Client, mTx monitoredTx) {
			mTxLogger := createMonitoredTxLogger(mTx)
			defer func(mTx monitoredTx, mTxLogger *log.Logger) {
				if err := recover(); err != nil {
					mTxLogger.Errorf("monitoring recovered from this err: %v", err)
				}
				wg.Done()
			}(mTx, mTxLogger)
			c.monitorTx(ctx, mTx, mTxLogger)
		}(c, mTx)
	}
	wg.Wait()

	return nil
}

// monitorTx does all the monitoring steps to the monitored tx
func (c *Client) monitorTx(ctx context.Context, mTx monitoredTx, logger *log.Logger) {
	var err error
	logger.Info("processing")
	// check if any of the txs in the history was confirmed
	var lastReceiptChecked types.Receipt
	// monitored tx is confirmed until we find a successful receipt
	confirmed := false
	// monitored tx doesn't have a failed receipt until we find a failed receipt for any
	// tx in the monitored tx history
	hasFailedReceipts := false
	// all history txs are considered mined until we can't find a receipt for any
	// tx in the monitored tx history
	allHistoryTxsWereMined := true
	for txHash := range mTx.history {
		mined, receipt, err := c.etherman.CheckTxWasMined(ctx, txHash)
		if err != nil {
			logger.Errorf("failed to check if tx %v was mined: %v", txHash.String(), err)
			continue
		}

		// if the tx is not mined yet, check that not all the tx were mined and go to the next
		if !mined {
			allHistoryTxsWereMined = false
			continue
		}

		lastReceiptChecked = *receipt

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
	if !confirmed && hasFailedReceipts && allHistoryTxsWereMined {
		logger.Infof("nonce needs to be updated")
		err := c.reviewMonitoredTxNonce(ctx, &mTx, logger)
		if err != nil {
			logger.Errorf("failed to review monitored tx nonce: %v", err)
			return
		}
		err = c.storage.Update(ctx, mTx)
		if err != nil {
			logger.Errorf("failed to update monitored tx nonce change: %v", err)
			return
		}
	}

	// if the history size reaches the max history size, this means something is really wrong with
	// this Tx and we are not able to identify automatically, so we mark this as failed to let the
	// caller know something is not right and needs to be review and to avoid to monitor this
	// tx infinitely
	// if len(mTx.history) == maxHistorySize {
	// 	mTx.status = MonitoredTxStatusFailed
	// 	mTxLogger.Infof("marked as failed because reached the history size limit: %v", err)
	// 	// update monitored tx changes into storage
	// 	err = c.storage.Update(ctx, mTx, nil)
	// 	if err != nil {
	// 		mTxLogger.Errorf("failed to update monitored tx when max history size limit reached: %v", err)
	// 		continue
	// 	}
	// }

	var signedTx *types.Transaction
	if !confirmed {
		// review tx and increase gas and gas price if needed
		if mTx.status == MonitoredTxStatusSent {
			err := c.reviewMonitoredTx(ctx, &mTx, logger)
			if err != nil {
				logger.Errorf("failed to review monitored tx: %v", err)
				return
			}
			err = c.storage.Update(ctx, mTx)
			if err != nil {
				logger.Errorf("failed to update monitored tx review change: %v", err)
				return
			}
		}

		// rebuild transaction
		tx := mTx.Tx()
		logger.Debugf("unsigned tx %v created", tx.Hash().String())

		// sign tx
		signedTx, err = c.etherman.SignTx(ctx, mTx.from, tx)
		if err != nil {
			logger.Errorf("failed to sign tx %v: %v", tx.Hash().String(), err)
			return
		}
		logger.Debugf("signed tx %v created", signedTx.Hash().String())

		// add tx to monitored tx history
		err = mTx.AddHistory(signedTx)
		if errors.Is(err, ErrAlreadyExists) {
			logger.Infof("signed tx already existed in the history")
		} else if err != nil {
			logger.Errorf("failed to add signed tx %v to monitored tx history: %v", signedTx.Hash().String(), err)
			return
		} else {
			// update monitored tx changes into storage
			err = c.storage.Update(ctx, mTx)
			if err != nil {
				logger.Errorf("failed to update monitored tx: %v", err)
				return
			}
			logger.Debugf("signed tx added to the monitored tx history")
		}

		// check if the tx is already in the network, if not, send it
		_, _, err = c.etherman.GetTx(ctx, signedTx.Hash())
		// if not found, send it tx to the network
		if errors.Is(err, ethereum.NotFound) {
			logger.Debugf("signed tx not found in the network")
			err := c.etherman.SendTx(ctx, signedTx)
			if err != nil {
				logger.Errorf("failed to send tx %v to network: %v", signedTx.Hash().String(), err)
				return
			}
			logger.Infof("signed tx sent to the network: %v", signedTx.Hash().String())
			if mTx.status == MonitoredTxStatusCreated {
				// update tx status to sent
				mTx.status = MonitoredTxStatusSent
				logger.Debugf("status changed to %v", string(mTx.status))
				// update monitored tx changes into storage
				err = c.storage.Update(ctx, mTx)
				if err != nil {
					logger.Errorf("failed to update monitored tx changes: %v", err)
					return
				}
			}
		} else {
			logger.Infof("signed tx already found in the network")
		}

		log.Infof("waiting signedTx to be mined...")

		// wait tx to get mined
		confirmed, err = c.etherman.WaitTxToBeMined(ctx, signedTx, c.cfg.WaitTxToBeMined.Duration)
		if err != nil {
			logger.Errorf("failed to wait tx to be mined: %v", err)
			return
		}
		if !confirmed {
			log.Infof("signedTx not mined yet and timeout has been reached")
			return
		}

		// get tx receipt
		var txReceipt *types.Receipt
		txReceipt, err = c.etherman.GetTxReceipt(ctx, signedTx.Hash())
		if err != nil {
			logger.Errorf("failed to get tx receipt for tx %v: %v", signedTx.Hash().String(), err)
			return
		}
		lastReceiptChecked = *txReceipt
	}

	// if mined, check receipt and mark as Failed or Confirmed
	if lastReceiptChecked.Status == types.ReceiptStatusSuccessful {
		// wait the number of l1 blocks configured to confirm the tx
		currentBlockNumber, err := c.etherman.GetLatestBlockNumber(ctx)
		if err != nil {
			logger.Errorf("failed to get latest block number: %v", err)
			return
		}

		if lastReceiptChecked.BlockNumber.Uint64()+c.cfg.L1ConfirmationBlocks > currentBlockNumber {
			mTx.status = MonitoredTxStatusMined
			mTx.blockNumber = lastReceiptChecked.BlockNumber
			logger.Info("mined")
		} else if c.shouldContinueToMonitorThisTx(ctx, lastReceiptChecked) {
			return
		}

	} else {
		// if we should continue to monitor, we move to the next one and this will
		// be reviewed in the next monitoring cycle
		if c.shouldContinueToMonitorThisTx(ctx, lastReceiptChecked) {
			return
		}
		// otherwise we understand this monitored tx has failed
		mTx.status = MonitoredTxStatusFailed
		mTx.blockNumber = lastReceiptChecked.BlockNumber
		logger.Info("failed")
	}

	// update monitored tx changes into storage
	err = c.storage.Update(ctx, mTx)
	if err != nil {
		logger.Errorf("failed to update monitored tx: %v", err)
		return
	}
}

// shouldContinueToMonitorThisTx checks the the tx receipt and decides if it should
// continue or not to monitor the monitored tx related to the tx from this receipt
func (c *Client) shouldContinueToMonitorThisTx(ctx context.Context, receipt types.Receipt) bool {
	// if the receipt has a is successful result, stop monitoring
	if receipt.Status == types.ReceiptStatusSuccessful {
		return false
	}

	tx, _, err := c.etherman.GetTx(ctx, receipt.TxHash)
	if err != nil {
		log.Errorf("failed to get tx when monitored tx identified as failed, tx : %v", receipt.TxHash.String(), err)
		return false
	}
	_, err = c.etherman.GetRevertMessage(ctx, tx)
	if err != nil {
		// if the error when getting the revert message is not identified, continue to monitor
		if err.Error() == ErrExecutionReverted.Error() {
			return true
		} else {
			log.Errorf("failed to get revert message for monitored tx identified as failed, tx %v: %v", receipt.TxHash.String(), err)
		}
	}
	// if nothing weird was found, stop monitoring
	return false
}

// reviewMonitoredTx checks if some field needs to be updated
// accordingly to the current information stored and the current
// state of the blockchain
func (c *Client) reviewMonitoredTx(ctx context.Context, mTx *monitoredTx, mTxLogger *log.Logger) error {
	mTxLogger.Debug("reviewing")
	// get gas
	gas, err := c.etherman.EstimateGas(ctx, mTx.from, mTx.to, mTx.value, mTx.data)
	if err != nil {
		err := fmt.Errorf("failed to estimate gas: %w", err)
		mTxLogger.Errorf(err.Error())
		return err
	}

	// check gas
	if gas > mTx.gas {
		mTxLogger.Infof("monitored tx gas updated from %v to %v", mTx.gas, gas)
		mTx.gas = gas
	}

	// get gas price
	gasPrice, err := c.suggestedGasPrice(ctx)
	if err != nil {
		err := fmt.Errorf("failed to get suggested gas price: %w", err)
		mTxLogger.Errorf(err.Error())
		return err
	}

	// check gas price
	if gasPrice.Cmp(mTx.gasPrice) == 1 {
		mTxLogger.Infof("monitored tx gas price updated from %v to %v", mTx.gasPrice.String(), gasPrice.String())
		mTx.gasPrice = gasPrice
	}
	return nil
}

// reviewMonitoredTxNonce checks if the nonce needs to be updated accordingly to
// the current nonce of the sender account.
//
// IMPORTANT: Nonce is reviewed apart from the other fields because it is a very
// sensible information and can make duplicated data to be sent to the blockchain,
// causing possible side effects and wasting resources.
func (c *Client) reviewMonitoredTxNonce(ctx context.Context, mTx *monitoredTx, mTxLogger *log.Logger) error {
	mTxLogger.Debug("reviewing nonce")
	nonce, err := c.etherman.CurrentNonce(ctx, mTx.from)
	if err != nil {
		err := fmt.Errorf("failed to load current nonce for acc %v: %w", mTx.from.String(), err)
		mTxLogger.Errorf(err.Error())
		return err
	}

	if nonce > mTx.nonce {
		mTxLogger.Infof("monitored tx nonce updated from %v to %v", mTx.nonce, nonce)
		mTx.nonce = nonce
	}

	return nil
}

func (c *Client) suggestedGasPrice(ctx context.Context) (*big.Int, error) {
	// get gas price
	gasPrice, err := c.etherman.SuggestedGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// adjust the gas price by the margin factor
	marginFactor := big.NewFloat(0).SetFloat64(c.cfg.GasPriceMarginFactor)
	fGasPrice := big.NewFloat(0).SetInt(gasPrice)
	adjustedGasPrice, _ := big.NewFloat(0).Mul(fGasPrice, marginFactor).Int(big.NewInt(0))

	// if there is a max gas price limit configured and the current
	// adjusted gas price is over this limit, set the gas price as the limit
	if c.cfg.MaxGasPriceLimit > 0 {
		maxGasPrice := big.NewInt(0).SetUint64(c.cfg.MaxGasPriceLimit)
		if adjustedGasPrice.Cmp(maxGasPrice) == 1 {
			adjustedGasPrice.Set(maxGasPrice)
		}
	}

	return adjustedGasPrice, nil
}

// logErrorAndWait used when an error is detected before trying again
func (c *Client) logErrorAndWait(msg string, err error) {
	log.Errorf(msg, err)
	time.Sleep(failureIntervalInSeconds * time.Second)
}

// ResultHandler used by the caller to handle results
// when processing monitored txs
type ResultHandler func(MonitoredTxResult)

// ProcessPendingMonitoredTxs will check all monitored txs
// and wait until all of them are either confirmed or failed before continuing
//
// for the confirmed and failed ones, the resultHandler will be triggered
func (c *Client) ProcessPendingMonitoredTxs(ctx context.Context, resultHandler ResultHandler) {
	statusesFilter := []MonitoredTxStatus{
		MonitoredTxStatusCreated,
		MonitoredTxStatusSent,
		MonitoredTxStatusFailed,
		MonitoredTxStatusMined,
	}
	// keep running until there are pending monitored txs
	for {
		results, err := c.ResultsByStatus(ctx, statusesFilter)
		if err != nil {
			// if something goes wrong here, we log, wait a bit and keep it in the infinite loop to not unlock the caller.
			log.Errorf("failed to get results by statuses from eth tx manager to monitored txs err: ", err)
			time.Sleep(time.Second)
			continue
		}

		if len(results) == 0 {
			// if there are not pending monitored txs, stop
			return
		}

		for _, result := range results {
			mTxResultLogger := CreateMonitoredTxResultLogger(result)

			// if the result is confirmed, we set it as done do stop looking into this monitored tx
			if result.Status == MonitoredTxStatusMined {
				err := c.setStatusFinalized(ctx, result.ID)
				if err != nil {
					mTxResultLogger.Errorf("failed to set monitored tx as done, err: %v", err)
					// if something goes wrong at this point, we skip this result and move to the next.
					// this result is going to be handled again in the next cycle by the outer loop.
					continue
				} else {
					mTxResultLogger.Info("monitored tx confirmed")
				}
				resultHandler(result)
				continue
			}

			// if the result is failed, we need to go around it and rebuild a batch verification
			if result.Status == MonitoredTxStatusFailed {
				resultHandler(result)
				continue
			}

			// if the result is either not confirmed or failed, it means we need to wait until it gets confirmed of failed.
			for {
				// wait before refreshing the result info
				time.Sleep(time.Second)

				// refresh the result info
				result, err := c.Result(ctx, result.ID)
				if err != nil {
					mTxResultLogger.Errorf("failed to get monitored tx result, err: %v", err)
					continue
				}

				// if the result status is confirmed or failed, breaks the wait loop
				if result.Status == MonitoredTxStatusMined || result.Status == MonitoredTxStatusFailed {
					break
				}

				mTxResultLogger.Infof("waiting for monitored tx to get confirmed, status: %v", result.Status.String())
			}
		}
	}
}

// createMonitoredTxLogger creates an instance of logger with all the important
// fields already set for a monitoredTx
func createMonitoredTxLogger(mTx monitoredTx) *log.Logger {
	return log.WithFields(
		"monitoredTxId", mTx.id,
		"createdAt", mTx.createdAt,
		"from", mTx.from,
		"to", mTx.to,
	)
}

// CreateLogger creates an instance of logger with all the important
// fields already set for a monitoredTx without requiring an instance of
// monitoredTx, this should be use in for callers before calling the ADD
// method
func CreateLogger(monitoredTxId common.Hash, from common.Address, to *common.Address) *log.Logger {
	return log.WithFields(
		"monitoredTxId", monitoredTxId.String(),
		"from", from,
		"to", to,
	)
}

// CreateMonitoredTxResultLogger creates an instance of logger with all the important
// fields already set for a MonitoredTxResult
func CreateMonitoredTxResultLogger(mTxResult MonitoredTxResult) *log.Logger {
	return log.WithFields(
		"monitoredTxId", mTxResult.ID.String(),
	)
}
