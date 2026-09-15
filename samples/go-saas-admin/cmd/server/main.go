package main

import (
	"fmt"
	"os"

	"github.com/mxl1c/CodeForge/samples/go-saas-admin/internal/order"
)

func main() {
	svc := order.NewService()
	svc.Seed("TNT-星河零售", "ORD-20260915-0088", order.StatusClosed)
	if err := svc.Refund("TNT-星河零售", "ORD-20260915-0088"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("go-saas-admin stub: refund path executed (known defect CF-W1-001)")
}
