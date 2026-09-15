package com.codeforge.saas.admin.order;

/** W1 stub: order approval with tenant + finance permission checks. */
public class OrderApprovalService {
    public enum OrderStatus {
        DRAFT,
        PENDING_FINANCE,
        APPROVED
    }

    public static final class Order {
        private final String tenantId;
        private final String skuId;
        private final int quantity;
        private OrderStatus status;

        public Order(String tenantId, String skuId, int quantity, OrderStatus status) {
            this.tenantId = tenantId;
            this.skuId = skuId;
            this.quantity = quantity;
            this.status = status;
        }

        public String getTenantId() { return tenantId; }
        public String getSkuId() { return skuId; }
        public int getQuantity() { return quantity; }
        public OrderStatus getStatus() { return status; }
        public void setStatus(OrderStatus status) { this.status = status; }
    }

    public Order approve(String actorTenantId, boolean financeRole, Order order, boolean stockAvailable) {
        if (!actorTenantId.equals(order.getTenantId())) {
            throw new SecurityException("cross-tenant approve is forbidden");
        }
        if (!financeRole) {
            throw new SecurityException("missing permission order.approve.finance");
        }
        if (order.getStatus() != OrderStatus.PENDING_FINANCE) {
            throw new IllegalStateException("illegal transit " + order.getStatus());
        }
        if (!stockAvailable) {
            order.setStatus(OrderStatus.PENDING_FINANCE);
            throw new IllegalStateException("insufficient stock; order rolled back to PENDING_FINANCE");
        }
        order.setStatus(OrderStatus.APPROVED);
        return order;
    }
}
