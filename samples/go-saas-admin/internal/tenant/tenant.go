package tenant

import "errors"

var ErrFrozen = errors.New("租户已冻结，禁止提交工单")

type Tenant struct {
	ID     string
	Name   string
	Frozen bool
}

func AssertActive(t Tenant) error {
	if t.Frozen {
		return ErrFrozen
	}
	return nil
}
