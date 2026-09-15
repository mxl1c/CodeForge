# go-saas-admin (W1 stub)

Desensitized **SaaS mid-office** sample: tenant-scoped order admin (Go).

- Known defects: closed-order refund (`CF-W1-001`), missing tenant filter on get/refund.
- Failing-test hook: `internal/order/service_test.go`
- Pair with `fixtures/failure-stack-zh.txt` and `fixtures/fake-pr.diff`.

```bash
go test ./...   # expected to fail until the known defects are fixed
```

No real customer data or secrets.
