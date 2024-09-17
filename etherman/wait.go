package etherman

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"time"

	"github.com/0xPolygon/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// DefaultInterval is a time interval
	DefaultInterval = 2 * time.Millisecond
	// DefaultDeadline is a time interval
	DefaultDeadline = 2 * time.Minute
	// DefaultTxMinedDeadline is a time interval
	DefaultTxMinedDeadline = 5 * time.Second
)

var (
	// ErrTimeoutReached is thrown when the timeout is reached and
	// because the condition is not matched
	ErrTimeoutReached = fmt.Errorf("timeout has been reached")
)

// Wait handles polliing until conditions are met.
type Wait struct{}

// NewWait is the Wait constructor.
func NewWait() *Wait {
	return &Wait{}
}

// Poll retries the given condition with the given interval until it succeeds
// or the given deadline expires.
func Poll(interval, deadline time.Duration, condition ConditionFunc) error {
	timeout := time.After(deadline)
	tick := time.NewTicker(interval)

	for {
		select {
		case <-timeout:
			return ErrTimeoutReached
		case <-tick.C:
			ok, err := condition()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

type ethClienter interface {
	ethereum.TransactionReader
	ethereum.ContractCaller
	bind.DeployBackend
}

// WaitTxToBeMined waits until a tx has been mined or the given timeout expires.
func WaitTxToBeMined(
	parentCtx context.Context,
	client ethClienter,
	tx *types.Transaction,
	timeout time.Duration,
) error {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, client, tx)
	if errors.Is(err, context.DeadlineExceeded) {
		return err
	} else if err != nil {
		log.Errorf("error waiting tx %s to be mined: %w", tx.Hash(), err)
		return err
	}
	if receipt.Status == types.ReceiptStatusFailed {
		// Get revert reason
		reason, reasonErr := RevertReason(ctx, client, tx, receipt.BlockNumber)
		if reasonErr != nil {
			reason = reasonErr.Error()
		}
		return fmt.Errorf("transaction has failed, reason: %s, receipt: %+v. tx: %+v, gas: %v", reason, receipt, tx, tx.Gas())
	}
	log.Debug("Transaction successfully mined: ", tx.Hash())
	return nil
}

// RevertReason returns the revert reason for a tx that has a receipt with failed status
func RevertReason(ctx context.Context, c ethClienter, tx *types.Transaction, blockNumber *big.Int) (string, error) {
	if tx == nil {
		return "", nil
	}

	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		signer := types.LatestSignerForChainID(tx.ChainId())
		from, err = types.Sender(signer, tx)
		if err != nil {
			return "", err
		}
	}
	msg := ethereum.CallMsg{
		From: from,
		To:   tx.To(),
		Gas:  tx.Gas(),

		Value: tx.Value(),
		Data:  tx.Data(),
	}
	hex, err := c.CallContract(ctx, msg, blockNumber)
	if err != nil {
		return "", err
	}

	unpackedMsg, err := abi.UnpackRevert(hex)
	if err != nil {
		log.Warnf("failed to get the revert message for tx %v: %v", tx.Hash(), err)
		return "", errors.New("execution reverted")
	}

	return unpackedMsg, nil
}

// WaitTxReceipt waits until a tx receipt is available or the given timeout expires.
func WaitTxReceipt(
	ctx context.Context,
	txHash common.Hash,
	timeout time.Duration,
	client *ethclient.Client,
) (*types.Receipt, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	var receipt *types.Receipt
	pollErr := Poll(DefaultInterval, timeout, func() (bool, error) {
		var err error
		receipt, err = client.TransactionReceipt(ctx, txHash)
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				time.Sleep(time.Second)
				return false, nil
			} else {
				return false, err
			}
		}
		return true, nil
	})
	if pollErr != nil {
		return nil, pollErr
	}
	return receipt, nil
}

// ConditionFunc is a generic function
type ConditionFunc func() (done bool, err error)

// WaitSignal blocks until an Interrupt or Kill signal is received, then it
// executes the given cleanup functions and returns.
func WaitSignal(cleanupFuncs ...func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for sig := range signals {
		switch sig {
		case os.Interrupt, os.Kill:
			log.Info("terminating application gracefully...")
			for _, cleanup := range cleanupFuncs {
				cleanup()
			}
			return
		}
	}
}
