package com.example.saasadmin.order;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertThrows;

class OrderServiceTest {

    @Test
    void refundShouldRejectClosedOrder() {
        OrderService service = new OrderService();
        service.seed("TNT-星河零售", "ORD-20260915-0088", "closed");

        // Failing-test hook: production currently refunds closed orders (CF-W1-001).
        assertThrows(IllegalStateException.class,
                () -> service.refund("TNT-星河零售", "ORD-20260915-0088"));
    }

    @Test
    void getShouldNotReturnOtherTenantOrder() {
        OrderService service = new OrderService();
        service.seed("TNT-星河零售", "ORD-20260915-0088", "paid");

        // Failing-test hook: lookup is not tenant-scoped.
        Order leaked = service.get("TNT-河图批发", "ORD-20260915-0088");
        if (leaked != null) {
            throw new AssertionError("known defect: cross-tenant order read succeeded");
        }
    }
}
