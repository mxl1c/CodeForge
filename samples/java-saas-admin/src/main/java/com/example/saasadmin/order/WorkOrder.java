package com.example.saasadmin.order;

public class WorkOrder {
    private final String orderId;
    private final String tenantId;
    private OrderState state;

    public WorkOrder(String orderId, String tenantId, OrderState state) {
        this.orderId = orderId;
        this.tenantId = tenantId;
        this.state = state;
    }

    public String getOrderId() {
        return orderId;
    }

    public String getTenantId() {
        return tenantId;
    }

    public OrderState getState() {
        return state;
    }

    public void setState(OrderState state) {
        this.state = state;
    }
}
