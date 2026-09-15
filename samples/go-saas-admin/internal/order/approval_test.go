package order

import (
	"errors"
	"testing"
)

func TestApplyDiscountRejectsNegative(t *testing.T) {
	if _, err := ApplyDiscount(100, -0.1); !errors.Is(err, ErrNegativeDiscount) {
		t.Fatalf("err = %v", err)
	}
}

func TestApproveRejectsCrossTenant(t *testing.T) {
	o := &Order{TenantID: "east-retail", Status: StatusPendingFinance}
	if err := Approve("west-retail", true, o, true); !errors.Is(err, ErrCrossTenant) {
		t.Fatalf("err = %v", err)
	}
}
