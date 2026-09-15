package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mxl1c/CodeForge/samples/go-saas-admin/internal/order"
	"github.com/mxl1c/CodeForge/samples/go-saas-admin/internal/tenant"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/orders/settle", settle)

	log.Println("go-saas-admin skeleton listening on :18080")
	log.Fatal(http.ListenAndServe(":18080", mux))
}

func settle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID  string `json:"orderId"`
		TenantID string `json:"tenantId"`
		Frozen   bool   `json:"frozen"`
		From     string `json:"from"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tn := tenant.Tenant{ID: req.TenantID, Frozen: req.Frozen}
	wo := order.WorkOrder{ID: req.OrderID, TenantID: req.TenantID, State: order.ParseState(req.From)}
	if err := order.Settle(&wo, tn); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	_ = json.NewEncoder(w).Encode(wo)
}
