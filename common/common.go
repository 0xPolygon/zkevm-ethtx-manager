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

// BoolToInteger converts the provided boolean value into integer value
func BoolToInteger(v bool) int {
	if v {
		return 1
	}

	return 0
}

// ToAddressOrNil converts a string to a common.Address pointer or returns nil if empty.
func ToAddressOrNil(addr string) *common.Address {
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
