package order

import (
	"fmt"

	"github.com/mxl1c/CodeForge/samples/go-saas-admin/internal/tenant"
)

type State string

const (
	Draft                State = "DRAFT"
	PendingFinanceReview State = "PENDING_FINANCE_REVIEW"
	Settled              State = "SETTLED"
	Closed               State = "CLOSED"
)

func ParseState(s string) State {
	switch State(s) {
	case Draft, PendingFinanceReview, Settled, Closed:
		return State(s)
	default:
		return Draft
	}
}

func label(s State) string {
	switch s {
	case Draft:
		return "草稿"
	case PendingFinanceReview:
		return "待财务复核"
	case Settled:
		return "已出账"
	case Closed:
		return "已关闭"
	default:
		return string(s)
	}
}

type WorkOrder struct {
	ID       string `json:"orderId"`
	TenantID string `json:"tenantId"`
	State    State  `json:"state"`
}

func Settle(order *WorkOrder, tn tenant.Tenant) error {
	if err := tenant.AssertActive(tn); err != nil {
		return err
	}
	return transition(order, PendingFinanceReview, Settled)
}

func transition(order *WorkOrder, from, to State) error {
	if order.State != from {
		return fmt.Errorf("工单状态不允许从「%s」流转到「%s」", label(order.State), label(to))
	}
	order.State = to
	return nil
}
