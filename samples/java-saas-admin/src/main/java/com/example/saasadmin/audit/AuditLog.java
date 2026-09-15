package com.example.saasadmin.audit;

public class AuditLog {
    private final String tenantId;
    private final String action;
    private final String detail;

    public AuditLog(String tenantId, String action, String detail) {
        this.tenantId = tenantId;
        this.action = action;
        this.detail = detail;
    }

    public String getTenantId() {
        return tenantId;
    }

    public String getAction() {
        return action;
    }

    public String getDetail() {
        return detail;
    }
}
