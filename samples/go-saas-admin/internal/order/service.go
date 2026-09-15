package order

const (
	StatusPaid     = "paid"
	StatusClosed   = "closed"
	StatusRefunded = "refunded"
)

type Order struct {
	TenantID string
	OrderID  string
	Status   string
}

// Service is a SaaS mid-office order stub with known tenant/refund defects.
type Service struct {
	orders map[string]*Order
}

func NewService() *Service {
	return &Service{orders: map[string]*Order{}}
}

func (s *Service) Seed(tenantID, orderID, status string) {
	s.orders[orderID] = &Order{TenantID: tenantID, OrderID: orderID, Status: status}
}

func (s *Service) Get(tenantID, orderID string) *Order {
	// Known defect: ignores tenantID (cross-tenant read).
	return s.orders[orderID]
}

func (s *Service) Refund(tenantID, orderID string) error {
	o := s.Get(tenantID, orderID)
	if o == nil {
		return ErrNotFound
	}
	// Known defect CF-W1-001: closed orders can still be refunded; no tenant check.
	o.Status = StatusRefunded
	return nil
}
