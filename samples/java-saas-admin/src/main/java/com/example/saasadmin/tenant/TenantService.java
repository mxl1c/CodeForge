package com.example.saasadmin.tenant;

import com.example.saasadmin.common.BizException;

public class TenantService {
    public void assertActive(Tenant tenant) {
        if (tenant == null || tenant.isFrozen()) {
            throw new BizException("TENANT_FROZEN", "租户已冻结，禁止提交工单");
        }
    }
}
