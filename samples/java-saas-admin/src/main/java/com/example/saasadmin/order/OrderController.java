package com.example.saasadmin.order;

import com.example.saasadmin.tenant.Tenant;

/** HTTP 入口骨架，对应 fixtures/failure-stack-zh.txt 中的调用栈。 */
public class OrderController {
    private final OrderService orders = new OrderService();

    public void settle(String orderId, Tenant tenant, WorkOrder order) {
        orders.settle(order, tenant);
    }
}
