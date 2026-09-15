package order

import (
	"errors"
	"fmt"
)

var (
	ErrCrossTenant       = errors.New("cross-tenant approve is forbidden")
	ErrMissingFinance    = errors.New("missing permission order.approve.finance")
	ErrIllegalTransit    = errors.New("illegal status transit")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Status string

const (
	StatusDraft          Status = "DRAFT"
	StatusPendingFinance Status = "PENDING_FINANCE"
	StatusApproved       Status = "APPROVED"
)

type Order struct {
	TenantID string
	SKUID    string
	Quantity int
	Status   Status
}

// Approve is the W1 stub for finance-gated order approval with inventory rollback.
func Approve(actorTenantID string, financeRole bool, o *Order, stockAvailable bool) error {
	if actorTenantID != o.TenantID {
		return ErrCrossTenant
	}
	if !financeRole {
		return ErrMissingFinance
	}
	if o.Status != StatusPendingFinance {
		return fmt.Errorf("%w: %s", ErrIllegalTransit, o.Status)
	}
	if !stockAvailable {
		o.Status = StatusPendingFinance
		return ErrInsufficientStock
	}
	o.Status = StatusApproved
	return nil
}
