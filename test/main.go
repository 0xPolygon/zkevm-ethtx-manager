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
	config := ethtxmanager.Config{
		FrequencyToMonitorTxs:             types.Duration{Duration: 1 * time.Second},
		WaitTxToBeMined:                   types.Duration{Duration: 2 * time.Minute},
		ConsolidationL1ConfirmationBlocks: 5,
		FinalizationL1ConfirmationBlocks:  10,
		PrivateKeys:                       []types.KeystoreFileConfig{{Path: "test.keystore", Password: "testonly"}},
		ForcedGas:                         0,
		GasPriceMarginFactor:              1,
		MaxGasPriceLimit:                  0,
		PersistenceFilename:               "ethtxmanager-persistence.json",
		Etherman: etherman.Config{
			URL:              "http://localhost:8545",
			HTTPHeaders:      map[string]string{},
			MultiGasProvider: false,
			L1ChainID:        1337,
		},
		Log: log.Config{Level: "info", Environment: "development", Outputs: []string{"stderr"}},
	}
	log.Init(config.Log)
	log.Debug("Creating ethtxmanager")
	client, err := ethtxmanager.New(config)
	if err != nil {
		panic(err)
	}
	log.Debug("ethtxmanager created")

	ctx := context.Background()

	// Get starting nonce
	testEtherman, err := etherman.NewClient(config.Etherman)
	if err != nil {
		log.Fatalf("Error creating etherman client: %s", err)
	}
	nonce, err := testEtherman.CurrentNonce(ctx, common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"))
	if err != nil {
		log.Fatalf("Error getting nonce: %s", err)
	}

	go client.Start()
	log.Debug("ethtxmanager started")
	sendTransaction(ctx, client, nonce)
	nonce++

	for i := 0; i < 10; i++ {
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
			if results[x].Status != ethtxmanager.MonitoredTxStatusFinalized {
				log.Debugf("Tx %s not finalized yet: %s", results[x].ID, results[x].Status)
				break
			}
			x++
		}

		if x == len(results) {
			log.Info("All txs finalized")
			break
		}
	}

	// Clean up
	results, err := client.ResultsByStatus(ctx, []ethtxmanager.MonitoredTxStatus{ethtxmanager.MonitoredTxStatusFinalized})
	if err != nil {
		log.Errorf("Error getting result: %s", err)
	}
	for _, result := range results {
		log.Infof("Removing tx %s", result.ID)
		err = client.Remove(ctx, result.ID)
		if err != nil {
			log.Errorf("Error removing tx %s: %s", result.ID, err)
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
