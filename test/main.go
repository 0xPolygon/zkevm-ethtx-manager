package main

import (
	"context"
	"math/big"
	"math/rand"
	"time"

	"github.com/0xPolygon/zkevm-ethtx-manager/config/types"
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	"github.com/0xPolygon/zkevm-ethtx-manager/log"
	coreTypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

var (
	to0 = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

func main() {
	config := ethtxmanager.Config{
		FrequencyToMonitorTxs:           types.Duration{Duration: 1 * time.Second},
		WaitTxToBeMined:                 types.Duration{Duration: 2 * time.Minute},
		GetReceiptMaxTime:               types.Duration{Duration: 10 * time.Second},
		GetReceiptWaitInterval:          types.Duration{Duration: 250 * time.Millisecond},
		ForcedGas:                       0,
		GasPriceMarginFactor:            1,
		MaxGasPriceLimit:                0,
		SafeStatusL1NumberOfBlocks:      0,
		FinalizedStatusL1NumberOfBlocks: 0,
		StoragePath:                     "ethtxmanager-persistence.db",
		ReadPendingL1Txs:                false,
		Log:                             log.Config{Level: "info", Environment: "development", Outputs: []string{"stderr"}},
		PrivateKeys:                     []types.KeystoreFileConfig{{Path: "test.keystore", Password: "testonly"}},
		Etherman: etherman.Config{
			URL:              "http://localhost:8545",
			HTTPHeaders:      map[string]string{},
			MultiGasProvider: false,
			L1ChainID:        1337,
		},
	}
	log.Init(config.Log)
	log.Debug("Creating ethtxmanager")
	client, err := ethtxmanager.New(config)
	if err != nil {
		panic(err)
	}
	log.Debug("ethtxmanager created")

	ctx := context.Background()

	go client.Start()
	log.Debug("ethtxmanager started")
	// sendBlobTransaction(ctx, client, nonce)
	// nonce++

	for i := 0; i < 1; i++ {
		time.Sleep(100 * time.Millisecond)
		sendTransaction(ctx, client)
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
			if results[x].Status != coreTypes.MonitoredTxStatusFinalized {
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
	results, err := client.ResultsByStatus(ctx, []coreTypes.MonitoredTxStatus{coreTypes.MonitoredTxStatusFinalized})
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

func sendTransaction(ctx context.Context, ethtxmanager *ethtxmanager.Client) common.Hash {
	id, err := ethtxmanager.Add(ctx, &to0, big.NewInt(1), []byte{byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256))}, 0, nil)
	if err != nil {
		log.Errorf("Error sending transaction: %s", err)
	} else {
		log.Infof("Transaction sent with id %s", id)
	}
	return id
}

func sendBlobTransaction(ctx context.Context, ethtxmanager *ethtxmanager.Client) common.Hash {
	blobBytes := []byte{255, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256))}
	blob, err := ethtxmanager.EncodeBlobData(blobBytes)
	if err != nil {
		log.Errorf("Error encoding blob data")
		return common.Hash{}
	}
	blobSidecar := ethtxmanager.MakeBlobSidecar([]kzg4844.Blob{blob})

	// data := []byte{228, 103, 97, 196} // pol method
	data := []byte{}
	id, err := ethtxmanager.Add(ctx, &to0, big.NewInt(0), data, 0, blobSidecar)
	if err != nil {
		log.Errorf("Error sending Blob transaction: %s", err)
	} else {
		log.Infof("Blob Transaction sent with id %s", id)
	}
	return id
}
