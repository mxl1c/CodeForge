package com.codeforge.saas.admin.order;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;

class DiscountPolicyTest {
    @Test
    void rejectsNegativeDiscount() {
        DiscountPolicy policy = new DiscountPolicy();
        Assertions.assertThrows(IllegalArgumentException.class,
                () -> policy.apply(new BigDecimal("100"), new BigDecimal("-0.1")));
    }
}
