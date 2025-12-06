package utils

import "fmt"

// FormatPrice converts a price stored as cents (100x) into a fixed two-decimal string.
func FormatPrice(price uint64) string {
	return fmt.Sprintf("%d.%02d", price/100, price%100)
}
