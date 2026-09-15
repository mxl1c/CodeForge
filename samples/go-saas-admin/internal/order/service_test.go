package order

import "testing"

func TestRefundRejectsClosedOrder(t *testing.T) {
	s := NewService()
	s.Seed("TNT-星河零售", "ORD-20260915-0088", StatusClosed)

	// Failing-test hook: production currently refunds closed orders (CF-W1-001).
	if err := s.Refund("TNT-星河零售", "ORD-20260915-0088"); err == nil {
		t.Fatal("known defect: closed order refund succeeded")
	}
}

func TestGetShouldNotReturnOtherTenantOrder(t *testing.T) {
	s := NewService()
	s.Seed("TNT-星河零售", "ORD-20260915-0088", StatusPaid)

	leaked := s.Get("TNT-河图批发", "ORD-20260915-0088")
	if leaked != nil {
		t.Fatal("known defect: cross-tenant order read succeeded")
	}
}
