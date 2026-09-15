package com.example.saasadmin.order;

import java.util.HashMap;
import java.util.Map;

/**
 * SaaS 中后台订单域骨架。已知缺陷：关闭订单仍可退款，且未校验租户隔离。
 */
public class OrderService {
    private final Map<String, Order> orders = new HashMap<>();

    public void seed(String tenantId, String orderId, String status) {
        orders.put(orderId, new Order(tenantId, orderId, status));
    }

    public Order get(String tenantId, String orderId) {
        // Known defect: ignores tenantId (cross-tenant read).
        return orders.get(orderId);
    }

    public void refund(String tenantId, String orderId) {
        Order order = get(tenantId, orderId);
        if (order == null) {
            throw new IllegalArgumentException("订单不存在: " + orderId);
        }
        // Known defect CF-W1-001: closed orders can still be refunded; no tenant check.
        order.status = "refunded";
    }
}
