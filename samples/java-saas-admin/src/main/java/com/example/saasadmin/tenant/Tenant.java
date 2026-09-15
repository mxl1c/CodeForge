package com.example.saasadmin.tenant;

public class Tenant {
    private final String tenantId;
    private final String displayName;
    private final boolean frozen;

    public Tenant(String tenantId, String displayName, boolean frozen) {
        this.tenantId = tenantId;
        this.displayName = displayName;
        this.frozen = frozen;
    }

    public String getTenantId() {
        return tenantId;
    }

    public String getDisplayName() {
        return displayName;
    }

    public boolean isFrozen() {
        return frozen;
    }
}
