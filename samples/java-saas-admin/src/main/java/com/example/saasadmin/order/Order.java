package com.example.saasadmin.order;

final class Order {
    final String tenantId;
    final String orderId;
    String status;

    Order(String tenantId, String orderId, String status) {
        this.tenantId = tenantId;
        this.orderId = orderId;
        this.status = status;
    }
}
