package order

import (
	"errors"
	"fmt"
)

var ErrNegativeDiscount = errors.New("discount must not be negative")

// ApplyDiscount is the W1 stub for SaaS admin list-price discounts.
func ApplyDiscount(listPrice, discountRate float64) (float64, error) {
	if discountRate < 0 {
		return 0, fmt.Errorf("%w", ErrNegativeDiscount)
	}
	return listPrice * (1 - discountRate), nil
}
