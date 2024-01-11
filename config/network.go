package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/0xPolygonHermez/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

type leafType uint8
type network string

const (
	mainnet network = "mainnet"
	testnet network = "testnet"
	custom  network = "custom"
	// LeafTypeBalance specifies that leaf stores Balance
	LeafTypeBalance leafType = 0
	// LeafTypeNonce specifies that leaf stores Nonce
	LeafTypeNonce leafType = 1
	// LeafTypeCode specifies that leaf stores Code
	LeafTypeCode leafType = 2
	// LeafTypeStorage specifies that leaf stores Storage Value
	LeafTypeStorage leafType = 3
	// LeafTypeSCLength specifies that leaf stores Storage Value
	LeafTypeSCLength leafType = 4
)

// NetworkConfig is the configuration struct for the different environments
type NetworkConfig struct {
	// L1: Configuration related to L1
	L1Config etherman.L1Config `json:"l1Config"`
	// L1: Genesis of the rollup, first block number and root
	Genesis Genesis
}

// GenesisFromJSON is the config file for network_custom
type GenesisFromJSON struct {
	// L1: root hash of the genesis block
	Root string `json:"root"`
	// L1: block number of the genesis block
	GenesisBlockNum uint64 `json:"genesisBlockNumber"`
	// L2:  List of states contracts used to populate merkle tree at initial state
	Genesis []genesisAccountFromJSON `json:"genesis"`
	// L1: configuration of the network
	L1Config etherman.L1Config
}

type genesisAccountFromJSON struct {
	// Address of the account
	Balance string `json:"balance"`
	// Nonce of the account
	Nonce string `json:"nonce"`
	// Address of the contract
	Address string `json:"address"`
	// Byte code of the contract
	Bytecode string `json:"bytecode"`
	// Initial storage of the contract
	Storage map[string]string `json:"storage"`
	// Name of the contract in L1 (e.g. "PolygonZkEVMDeployer", "PolygonZkEVMBridge",...)
	ContractName string `json:"contractName"`
}

func (cfg *Config) loadNetworkConfig(ctx *cli.Context) {
	var networkJSON string
	switch ctx.String(FlagNetwork) {
	case string(mainnet):
		networkJSON = MainnetNetworkConfigJSON
	case string(testnet):
		networkJSON = TestnetNetworkConfigJSON
	case string(custom):
		var err error
		cfgPath := ctx.String(FlagCustomNetwork)
		networkJSON, err = LoadGenesisFileAsString(cfgPath)
		if err != nil {
			panic(err.Error())
		}
	default:
		log.Fatalf("unsupported --network value. Must be one of: [%s, %s, %s]", mainnet, testnet, custom)
	}
	config, err := LoadGenesisFromJSONString(networkJSON)
	if err != nil {
		panic(fmt.Errorf("failed to load genesis configuration from file. Error: %v", err))
	}
	cfg.NetworkConfig = config
}

// LoadGenesisFileAsString loads the genesis file as a string
func LoadGenesisFileAsString(cfgPath string) (string, error) {
	if cfgPath != "" {
		f, err := os.Open(cfgPath) //nolint:gosec
		if err != nil {
			return "", err
		}
		defer func() {
			err := f.Close()
			if err != nil {
				log.Error(err)
			}
		}()

		b, err := io.ReadAll(f)
		if err != nil {
			return "", err
		}
		return string(b), nil
	} else {
		return "", errors.New("custom network file not provided. Please use the custom-network-file flag")
	}
}

// LoadGenesisFromJSONString loads the genesis file from JSON string
func LoadGenesisFromJSONString(jsonStr string) (NetworkConfig, error) {
	var cfg NetworkConfig

	var cfgJSON GenesisFromJSON
	if err := json.Unmarshal([]byte(jsonStr), &cfgJSON); err != nil {
		return NetworkConfig{}, err
	}

	if len(cfgJSON.Genesis) == 0 {
		return cfg, nil
	}

	cfg.L1Config = cfgJSON.L1Config
	cfg.Genesis = Genesis{
		BlockNumber: cfgJSON.GenesisBlockNum,
		Root:        common.HexToHash(cfgJSON.Root),
		Actions:     []*GenesisAction{},
	}

	for _, account := range cfgJSON.Genesis {
		if account.Balance != "" && account.Balance != "0" {
			action := &GenesisAction{
				Address: account.Address,
				Type:    int(LeafTypeBalance),
				Value:   account.Balance,
			}
			cfg.Genesis.Actions = append(cfg.Genesis.Actions, action)
		}
		if account.Nonce != "" && account.Nonce != "0" {
			action := &GenesisAction{
				Address: account.Address,
				Type:    int(LeafTypeNonce),
				Value:   account.Nonce,
			}
			cfg.Genesis.Actions = append(cfg.Genesis.Actions, action)
		}
		if account.Bytecode != "" {
			action := &GenesisAction{
				Address:  account.Address,
				Type:     int(LeafTypeCode),
				Bytecode: account.Bytecode,
			}
			cfg.Genesis.Actions = append(cfg.Genesis.Actions, action)
		}
		if len(account.Storage) > 0 {
			for storageKey, storageValue := range account.Storage {
				action := &GenesisAction{
					Address:         account.Address,
					Type:            int(LeafTypeStorage),
					StoragePosition: storageKey,
					Value:           storageValue,
				}
				cfg.Genesis.Actions = append(cfg.Genesis.Actions, action)
			}
		}
	}

	return cfg, nil
}
