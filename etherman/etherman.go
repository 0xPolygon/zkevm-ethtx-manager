package etherman

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygon/zkevm-ethtx-manager/etherman/etherscan"
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman/ethgasstation"
	"github.com/0xPolygon/zkevm-ethtx-manager/log"
	signertypes "github.com/agglayer/go_signer/signer/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	// ErrNotFound is used when the object is not found
	ErrNotFound = ethereum.NotFound
	// ErrPrivateKeyNotFound used when the provided sender does not have a private key registered to be used
	ErrPrivateKeyNotFound = errors.New("can't find sender private key to sign tx")
)

// EthereumClient is an interface that combines all the ethereum client interfaces
type EthereumClient interface {
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.ContractCaller
	ethereum.GasEstimator
	ethereum.GasPricer
	ethereum.GasPricer1559
	ethereum.PendingStateReader
	ethereum.TransactionReader
	ethereum.TransactionSender
	bind.DeployBackend
}

// EthermanSigner is an interface that combines all the signer interfaces
type EthermanSigner interface {
	SignTx(ctx context.Context, sender common.Address, tx *types.Transaction) (*types.Transaction, error)
	PublicAddress() ([]common.Address, error)
}

// Client is a simple implementation of EtherMan.
type Client struct {
	EthClient    EthereumClient
	cfg          Config
	GasProviders externalGasProviders
	auth         EthermanSigner // empty in case of read-only client
}

type externalGasProviders struct {
	MultiGasProvider bool
	Providers        []ethereum.GasPricer
}

// NewClient creates a new etherman.
func NewClient(cfg Config, signersConfig []signertypes.SignerConfig) (*Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("Ethereum node URL cannot be empty")
	}

	// Connect to ethereum node
	ethClient, err := ethclient.Dial(cfg.URL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.URL, err)
		return nil, err
	}

	for key, value := range cfg.HTTPHeaders {
		ethClient.Client().SetHeader(key, value)
	}

	// Fetch chain ID if not provided
	if cfg.L1ChainID == 0 {
		chainID, err := ethClient.ChainID(context.Background())
		if err != nil {
			log.Errorf("Failed to fetch chain ID from node: %+v", err)
			return nil, err
		}
		cfg.L1ChainID = chainID.Uint64()
		log.Infof("Etherman L1ChainID set to %d from node URL", cfg.L1ChainID)
	}

	gProviders := []ethereum.GasPricer{ethClient}
	if cfg.MultiGasProvider {
		if cfg.Etherscan.ApiKey == "" {
			log.Info("No ApiKey provided for etherscan. Ignoring provider...")
		} else {
			log.Info("ApiKey detected for etherscan")
			gProviders = append(gProviders, etherscan.NewEtherscanService(cfg.Etherscan.ApiKey))
		}
		gProviders = append(gProviders, ethgasstation.NewEthGasStationService())
	}
	auth, err := NewEthermanSigners(context.Background(), cfg.L1ChainID, signersConfig)
	if err != nil {
		return nil, err
	}

	return &Client{EthClient: ethClient,
		cfg: cfg,
		GasProviders: externalGasProviders{
			MultiGasProvider: cfg.MultiGasProvider,
			Providers:        gProviders,
		},
		auth: auth,
	}, nil
}

// GetTx function get ethereum tx
func (etherMan *Client) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := etherMan.EthClient.TransactionByHash(ctx, txHash)
	return tx, isPending, translateError(err)
}

// GetTxReceipt function gets ethereum tx receipt
func (etherMan *Client) GetTxReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	recepit, err := etherMan.EthClient.TransactionReceipt(ctx, txHash)
	return recepit, translateError(err)
}

// GetLatestBlockNumber gets the latest block number from the ethereum
func (etherMan *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	number, err := etherMan.getBlockNumber(ctx, rpc.LatestBlockNumber)
	return number, translateError(err)
}

// WaitTxToBeMined waits for an L1 tx to be mined. It will return error if the tx is reverted or timeout is exceeded
func (etherMan *Client) WaitTxToBeMined(
	ctx context.Context,
	tx *types.Transaction,
	timeout time.Duration,
) (bool, error) {
	err := WaitTxToBeMined(ctx, etherMan.EthClient, tx, timeout)
	err = translateError(err)
	if errors.Is(err, context.DeadlineExceeded) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetL1GasPrice gets the l1 gas price
func (etherMan *Client) GetL1GasPrice(ctx context.Context) *big.Int {
	// Get gasPrice from providers
	gasPrice := big.NewInt(0)
	for i, prov := range etherMan.GasProviders.Providers {
		gp, err := prov.SuggestGasPrice(ctx)
		if err != nil {
			log.Warnf("error getting gas price from provider %d. Error: %s", i+1, err.Error())
		} else if gasPrice.Cmp(gp) == -1 { // gasPrice < gp
			gasPrice = gp
		}
	}
	log.Debug("gasPrice chose: ", gasPrice)
	return gasPrice
}

// SendTx sends a tx to L1
func (etherMan *Client) SendTx(ctx context.Context, tx *types.Transaction) error {
	return etherMan.EthClient.SendTransaction(ctx, tx)
}

// CurrentNonce returns the current nonce for the provided account
func (etherMan *Client) CurrentNonce(ctx context.Context, account common.Address) (uint64, error) {
	return etherMan.EthClient.NonceAt(ctx, account, nil)
}

// PendingNonce returns the pending nonce for the provided account
func (etherMan *Client) PendingNonce(ctx context.Context, account common.Address) (uint64, error) {
	return etherMan.EthClient.PendingNonceAt(ctx, account)
}

// SuggestedGasPrice returns the suggest nonce for the network at the moment
func (etherMan *Client) SuggestedGasPrice(ctx context.Context) (*big.Int, error) {
	suggestedGasPrice := etherMan.GetL1GasPrice(ctx)
	if suggestedGasPrice.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("failed to get the suggested gas price")
	}
	return suggestedGasPrice, nil
}

// EstimateGas returns the estimated gas for the tx
func (etherMan *Client) EstimateGas(
	ctx context.Context,
	from common.Address,
	to *common.Address,
	value *big.Int,
	data []byte,
) (uint64, error) {
	return etherMan.EthClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	})
}

// EstimateGasBlobTx returns the estimated gas for the blob tx
func (etherMan *Client) EstimateGasBlobTx(
	ctx context.Context,
	from common.Address,
	to *common.Address,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	value *big.Int,
	data []byte,
) (uint64, error) {
	return etherMan.EthClient.EstimateGas(ctx, ethereum.CallMsg{
		From:      from,
		To:        to,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Value:     value,
		Data:      data,
	})
}

// CheckTxWasMined check if a tx was already mined
func (etherMan *Client) CheckTxWasMined(ctx context.Context, txHash common.Hash) (bool, *types.Receipt, error) {
	receipt, err := etherMan.EthClient.TransactionReceipt(ctx, txHash)
	err = translateError(err)
	if errors.Is(err, ethereum.NotFound) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	}
	return true, receipt, nil
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if err.Error() == ethereum.NotFound.Error() {
		return ethereum.NotFound
	}
	return err
}

// GetRevertMessage tries to get a revert message of a transaction
func (etherMan *Client) GetRevertMessage(ctx context.Context, tx *types.Transaction) (string, error) {
	if tx == nil {
		return "", nil
	}

	receipt, err := etherMan.GetTxReceipt(ctx, tx.Hash())
	err = translateError(err)
	if err != nil {
		return "", err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		revertMessage, err := RevertReason(ctx, etherMan.EthClient, tx, receipt.BlockNumber)
		if err != nil {
			return "", err
		}
		return revertMessage, nil
	}
	return "", nil
}

// getBlockNumber gets the block header by the provided block number from the ethereum
func (etherMan *Client) getBlockNumber(ctx context.Context, blockNumber rpc.BlockNumber) (uint64, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil || header == nil {
		return 0, err
	}
	return header.Number.Uint64(), nil
}

// GetHeaderByNumber returns a block header from the current canonical chain.
// If number is nil the latest header is returned
func (etherMan *Client) GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, number)
	return header, err
}

// GetSuggestGasTipCap retrieves the currently suggested gas tip cap after EIP-1559 for timely transaction execution.
func (etherMan *Client) GetSuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	gasTipCap, err := etherMan.EthClient.SuggestGasTipCap(ctx)
	return gasTipCap, err
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (etherMan *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return etherMan.EthClient.HeaderByNumber(ctx, number)
}

// SignTx tries to sign a transaction accordingly to the provided sender
func (etherMan *Client) SignTx(
	ctx context.Context,
	sender common.Address,
	tx *types.Transaction,
) (*types.Transaction, error) {
	return etherMan.auth.SignTx(ctx, sender, tx)
}

// PublicAddress returns the public addresses of the signers
func (etherMan *Client) PublicAddress() ([]common.Address, error) {
	return etherMan.auth.PublicAddress()
}
