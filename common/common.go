package common

import "github.com/ethereum/go-ethereum/common"

const (
	// Base10 decimal base
	Base10 = 10
	// Gwei represents 1000000000 wei
	Gwei = 1000000000

	// SQLLiteDriverName is the name for the SQL lite driver
	SQLLiteDriverName = "sqlite3"
)

// ToAddressPtr converts a string to a common.Address pointer or returns nil if empty.
func ToAddressPtr(addr string) *common.Address {
	if addr == "" {
		return nil
	}

	address := common.HexToAddress(addr)
	return &address
}

// ToUint64Ptr is a helper to create uint64 pointer
func ToUint64Ptr(v uint64) *uint64 {
	return &v
}

// SlicePtrsToSlice converts a slice of pointers to a slice of values.
func SlicePtrsToSlice[T any](ptrSlice []*T) []T {
	// Create a new slice to hold the values
	res := make([]T, len(ptrSlice))
	// Dereference each pointer and add the value to the result slice
	for i, ptr := range ptrSlice {
		if ptr != nil {
			res[i] = *ptr
		}
	}
	return res
}
