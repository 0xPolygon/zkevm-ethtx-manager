package main

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPolygonHermez/zkevm-ethtx-manager/config/types"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/ethtxmanager"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum/common"
)

var (
	to1 = common.HexToAddress("0x0001")
)

func main() {
	log.Init(log.Config{Level: "info", Environment: "development", Outputs: []string{"stderr"}})

	config := ethtxmanager.Config{
		FrequencyToMonitorTxs: types.Duration{Duration: 1 * time.Second},
		WaitTxToBeMined:       types.Duration{Duration: 2 * time.Minute},
		PrivateKeys:           []types.KeystoreFileConfig{{Path: "test.keystore", Password: "testonly"}},
		ForcedGas:             0,
		GasPriceMarginFactor:  1,
		MaxGasPriceLimit:      0,
		From:                  common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
		Etherman: etherman.Config{
			URL:              "http://localhost:8545",
			MultiGasProvider: false,
			L1ChainID:        1337,
		},
	}

	log.Debug("Creating ethtxmanager")
	client, err := ethtxmanager.New(config)
	if err != nil {
		panic(err)
	}
	log.Debug("ethtxmanager created")

	ctx := context.Background()

	nonce := uint64(17)
	go client.Start()
	log.Debug("ethtxmanager started")
	sendTransaction(ctx, client, nonce)
	nonce++

	for i := 0; i < 500; i++ {
		time.Sleep(100 * time.Millisecond)
		sendTransaction(ctx, client, nonce)
		nonce++
	}

	for {
		time.Sleep(5 * time.Second)
		// Check all sent tx are confirmed
		results, err := client.ResultsByStatus(ctx, nil)
		if err != nil {
			log.Errorf("Error getting result: %s", err)
		}

		x := 0
		for x < len(results) {
			if results[x].Status != ethtxmanager.MonitoredTxStatusConfirmed {
				log.Debugf("Tx %s not confirmed yet: %s", results[x].ID, results[x].Status)
				break
			}
			x++
		}

		if x == len(results) {
			log.Info("All txs confirmed")
			break
		}
	}
}

func sendTransaction(ctx context.Context, ethtxmanager *ethtxmanager.Client, nonce uint64) common.Hash {
	id, err := ethtxmanager.Add(ctx, &to1, &nonce, big.NewInt(1), []byte{})
	if err != nil {
		log.Errorf("Error sending transaction: %s", err)
	} else {
		log.Infof("Transaction sent with id %s", id)
	}
	return id
}
