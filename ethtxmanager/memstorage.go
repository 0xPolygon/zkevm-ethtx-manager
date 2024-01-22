package ethtxmanager

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// MemStorage hold txs to be managed
type MemStorage struct {
	Transactions map[common.Hash]monitoredTx
}

// NewPostgresStorage creates a new instance of storage that use
// postgres to store data
func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{Transactions: make(map[common.Hash]monitoredTx)}, nil
}

// Add persist a monitored tx
func (s *MemStorage) Add(ctx context.Context, mTx monitoredTx) error {
	if _, exists := s.Transactions[mTx.id]; exists {
		return ErrAlreadyExists
	}
	s.Transactions[mTx.id] = mTx
	return nil
}

// Get loads a persisted monitored tx
func (s *MemStorage) Get(ctx context.Context, id common.Hash) (monitoredTx, error) {
	if mTx, exists := s.Transactions[id]; exists {
		return mTx, nil
	}
	return monitoredTx{}, ErrNotFound
}

// GetByStatus loads all monitored tx that match the provided status
func (s *MemStorage) GetByStatus(ctx context.Context, statuses []MonitoredTxStatus) ([]monitoredTx, error) {
	mTxs := []monitoredTx{}
	for _, mTx := range s.Transactions {
		for _, status := range statuses {
			if mTx.status == status {
				mTxs = append(mTxs, mTx)
			}
		}
	}
	return mTxs, nil
}

// GetByBlock loads all monitored tx that have the blockNumber between
// fromBlock and toBlock
func (s *MemStorage) GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]monitoredTx, error) {
	mTxs := []monitoredTx{}
	for _, mTx := range s.Transactions {
		if fromBlock != nil && mTx.blockNumber.Uint64() < *fromBlock {
			continue
		}
		if toBlock != nil && mTx.blockNumber.Uint64() > *toBlock {
			continue
		}
		mTxs = append(mTxs, mTx)
	}
	return mTxs, nil
}

// Update a persisted monitored tx
func (s *MemStorage) Update(ctx context.Context, mTx monitoredTx) error {
	if _, exists := s.Transactions[mTx.id]; !exists {
		return ErrNotFound
	}
	mTx.updatedAt = time.Now()
	s.Transactions[mTx.id] = mTx
	return nil
}
