package common

const (
	// Base10 decimal base
	Base10 = 10
	// Gwei represents 1000000000 wei
	Gwei = 1000000000
)

// BoolToInteger converts the provided boolean value into integer value
func BoolToInteger(v bool) int {
	if v {
		return 1
	}

	return 0
}
