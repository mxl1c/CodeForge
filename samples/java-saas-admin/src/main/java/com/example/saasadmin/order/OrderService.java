package com.example.saasadmin.order;

import com.example.saasadmin.common.BizException;
import com.example.saasadmin.tenant.Tenant;
import com.example.saasadmin.tenant.TenantService;

/**
 * 工单状态机。允许 DRAFT → PENDING_FINANCE_REVIEW → SETTLED。
 * 禁止从「待财务复核」直接跳到「已出账」以外的非法流转。
 */
public class OrderService {
    private final TenantService tenants = new TenantService();

    public void submitForFinanceReview(WorkOrder order, Tenant tenant) {
        tenants.assertActive(tenant);
        transition(order, OrderState.DRAFT, OrderState.PENDING_FINANCE_REVIEW);
    }

    public void settle(WorkOrder order, Tenant tenant) {
        tenants.assertActive(tenant);
        transition(order, OrderState.PENDING_FINANCE_REVIEW, OrderState.SETTLED);
    }

    void transition(WorkOrder order, OrderState from, OrderState to) {
        if (order.getState() != from) {
            throw new BizException(
                    "ILLEGAL_STATE",
                    "工单状态不允许从「" + label(order.getState()) + "」流转到「" + label(to) + "」");
        }
        order.setState(to);
    }

    static String label(OrderState state) {
        switch (state) {
            case DRAFT:
                return "草稿";
            case PENDING_FINANCE_REVIEW:
                return "待财务复核";
            case SETTLED:
                return "已出账";
            case CLOSED:
                return "已关闭";
            default:
                return String.valueOf(state);
        }
    }
}
