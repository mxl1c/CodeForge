package com.codeforge.saas.admin.order;

import java.math.BigDecimal;

/** W1 stub: list-price discount used by the SaaS admin order form. */
public final class DiscountPolicy {
    public BigDecimal apply(BigDecimal listPrice, BigDecimal discountRate) {
        if (listPrice == null || discountRate == null) {
            throw new IllegalArgumentException("listPrice and discountRate are required");
        }
        if (discountRate.signum() < 0) {
            throw new IllegalArgumentException("discount must not be negative");
        }
        return listPrice.multiply(BigDecimal.ONE.subtract(discountRate));
    }
}
