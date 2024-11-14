package ethtxmanager

import (
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygon/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum/common"
)

func NewClientFromAddr(cfg Config, from common.Address) (*Client, error) { //nolint:all
	etherman, err := etherman.NewClient(cfg.Etherman)
	if err != nil {
		return nil, err
	}

	storage, err := createStorage(cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	client := Client{
		cfg:      cfg,
		etherman: etherman,
		storage:  storage,
		from:     from,
	}

	log.Init(cfg.Log)

	return &client, nil
}
