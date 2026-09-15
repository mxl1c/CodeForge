package order_test

import (
	"testing"

	"github.com/mxl1c/CodeForge/samples/go-saas-admin/internal/order"
	"github.com/mxl1c/CodeForge/samples/go-saas-admin/internal/tenant"
)

func TestSettleRejectsIllegalTransition(t *testing.T) {
	wo := order.WorkOrder{ID: "WO-1", TenantID: "T-10086", State: order.Draft}
	err := order.Settle(&wo, tenant.Tenant{ID: "T-10086"})
	if err == nil {
		t.Fatal("expected illegal transition")
	}
}
