package etherman

import (
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman/etherscan"
	"github.com/ethereum/go-ethereum/common"
)

// Config represents the configuration of the etherman
type Config struct {
	// URL is the URL of the Ethereum node for L1
	URL string `mapstructure:"URL"`

	// allow that L1 gas price calculation use multiples sources
	MultiGasProvider bool `mapstructure:"MultiGasProvider"`
	// Configuration for use Etherscan as used as gas provider, basically it needs the API-KEY
	Etherscan etherscan.Config
	// L1ChainID is the chain ID of the L1
	L1ChainID uint64 `mapstructure:"L1ChainID"`
	// HTTPHeaders are the headers to be used in the HTTP requests
	HTTPHeaders map[string]string `mapstructure:"HTTPHeaders"`
	// X Layer
	// ZkEVMAddr Address of the L1 contract polygonZkEVMAddress
	ZkEVMAddr common.Address `mapstructure:"PolygonZkEVMAddress"`
	// RollupManagerAddr Address of the L1 contract
	RollupManagerAddr common.Address `mapstructure:"PolygonRollupManagerAddress"`
}
