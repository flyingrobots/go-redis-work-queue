// Copyright 2025 James Ross
package multitenantiso

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// TenantHandler provides HTTP handlers for tenant management
type TenantHandler struct {
	tenantManager *TenantManager
	config        *Config
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(redisClient *redis.Client, config *Config) *TenantHandler {
	return &TenantHandler{
		tenantManager: NewTenantManager(redisClient),
		config:        config,
	}
}

// CreateTenantHandler handles POST /tenants
func (th *TenantHandler) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string            `json:"id"`
		Name         string            `json:"name"`
		ContactEmail string            `json:"contact_email,omitempty"`
		Metadata     map[string]string `json:"metadata,omitempty"`
		Quotas       *TenantQuotas     `json:"quotas,omitempty"`
		Encryption   *TenantEncryption `json:"encryption,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		th.writeError(w, http.StatusBadRequest, "invalid JSON", err)
		return
	}

	tenantID := TenantID(req.ID)
	config := th.config.GetDefaultTenantConfig(tenantID, req.Name)
	config.ContactEmail = req.ContactEmail
	if req.Metadata != nil {
		config.Metadata = req.Metadata
	}
	if req.Quotas != nil {
		if err := th.config.ValidateTenantQuotas(req.Quotas); err != nil {
			th.writeError(w, http.StatusBadRequest, "invalid quotas", err)
			return
		}
		config.Quotas = *req.Quotas
	}
	if req.Encryption != nil {
		if !th.config.IsKEKProviderAllowed(req.Encryption.KEKProvider) {
			th.writeError(w, http.StatusBadRequest, "KEK provider not allowed", nil)
			return
		}
		config.Encryption = *req.Encryption
	}

	if err := th.tenantManager.CreateTenant(config); err != nil {
		if IsTenantNotFound(err) {
			th.writeError(w, http.StatusConflict, "tenant already exists", err)
		} else {
			th.writeError(w, http.StatusInternalServerError, "failed to create tenant", err)
		}
		return
	}

	// Log audit event
	_ = th.tenantManager.LogAuditEvent(&AuditEvent{
		TenantID:  tenantID,
		UserID:    th.getUserID(r),
		Action:    "CREATE_TENANT",
		Resource:  "tenant:" + string(tenantID),
		Details:   map[string]interface{}{"tenant_name": req.Name},
		RemoteIP:  th.getRemoteIP(r),
		UserAgent: r.UserAgent(),
		Result:    "SUCCESS",
	})

	th.writeJSON(w, http.StatusCreated, config)
}

// GetTenantHandler handles GET /tenants/{id}
func (th *TenantHandler) GetTenantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := TenantID(vars["id"])

	// TODO: Validate access when RBAC is implemented
	// if err := th.tenantManager.ValidateAccess(th.getUserID(r), tenantID, "tenant", "read"); err != nil {
	// 	if IsAccessDenied(err) {
	// 		th.writeError(w, http.StatusForbidden, "access denied", err)
	// 	} else {
	// 		th.writeError(w, http.StatusInternalServerError, "access validation failed", err)
	// 	}
	// 	return
	// }

	config, err := th.tenantManager.GetTenant(tenantID)
	if err != nil {
		if IsTenantNotFound(err) {
			th.writeError(w, http.StatusNotFound, "tenant not found", err)
		} else {
			th.writeError(w, http.StatusInternalServerError, "failed to get tenant", err)
		}
		return
	}

	th.writeJSON(w, http.StatusOK, config)
}

// UpdateTenantHandler handles PUT /tenants/{id}
func (th *TenantHandler) UpdateTenantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := TenantID(vars["id"])

	// TODO: Validate access when RBAC is implemented

	var config TenantConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		th.writeError(w, http.StatusBadRequest, "invalid JSON", err)
		return
	}

	config.ID = tenantID // Ensure ID matches URL

	if err := th.config.ValidateTenantQuotas(&config.Quotas); err != nil {
		th.writeError(w, http.StatusBadRequest, "invalid quotas", err)
		return
	}

	if err := th.tenantManager.UpdateTenant(&config); err != nil {
		if IsTenantNotFound(err) {
			th.writeError(w, http.StatusNotFound, "tenant not found", err)
		} else {
			th.writeError(w, http.StatusInternalServerError, "failed to update tenant", err)
		}
		return
	}

	// Log audit event
	_ = th.tenantManager.LogAuditEvent(&AuditEvent{
		TenantID:  tenantID,
		UserID:    th.getUserID(r),
		Action:    "UPDATE_TENANT",
		Resource:  "tenant:" + string(tenantID),
		Details:   map[string]interface{}{"updated_fields": "config"},
		RemoteIP:  th.getRemoteIP(r),
		UserAgent: r.UserAgent(),
		Result:    "SUCCESS",
	})

	th.writeJSON(w, http.StatusOK, config)
}

// DeleteTenantHandler handles DELETE /tenants/{id}
func (th *TenantHandler) DeleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := TenantID(vars["id"])

	// TODO: Validate access when RBAC is implemented

	if err := th.tenantManager.DeleteTenant(tenantID); err != nil {
		if IsTenantNotFound(err) {
			th.writeError(w, http.StatusNotFound, "tenant not found", err)
		} else {
			th.writeError(w, http.StatusInternalServerError, "failed to delete tenant", err)
		}
		return
	}

	// Log audit event
	_ = th.tenantManager.LogAuditEvent(&AuditEvent{
		TenantID:  tenantID,
		UserID:    th.getUserID(r),
		Action:    "DELETE_TENANT",
		Resource:  "tenant:" + string(tenantID),
		Details:   map[string]interface{}{},
		RemoteIP:  th.getRemoteIP(r),
		UserAgent: r.UserAgent(),
		Result:    "SUCCESS",
	})

	w.WriteHeader(http.StatusNoContent)
}

// ListTenantsHandler handles GET /tenants
func (th *TenantHandler) ListTenantsHandler(w http.ResponseWriter, r *http.Request) {
	summaries, err := th.tenantManager.ListTenants()
	if err != nil {
		th.writeError(w, http.StatusInternalServerError, "failed to list tenants", err)
		return
	}

	th.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenants": summaries,
		"count":   len(summaries),
	})
}

// GetTenantQuotaUsageHandler handles GET /tenants/{id}/quota-usage
func (th *TenantHandler) GetTenantQuotaUsageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := TenantID(vars["id"])

	// TODO: Validate access when RBAC is implemented

	usage, err := th.tenantManager.GetQuotaUsage(tenantID)
	if err != nil {
		th.writeError(w, http.StatusInternalServerError, "failed to get quota usage", err)
		return
	}

	th.writeJSON(w, http.StatusOK, usage)
}

// CheckQuotaHandler handles POST /tenants/{id}/check-quota
func (th *TenantHandler) CheckQuotaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := TenantID(vars["id"])

	var req struct {
		QuotaType string `json:"quota_type"`
		Amount    int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		th.writeError(w, http.StatusBadRequest, "invalid JSON", err)
		return
	}

	err := th.tenantManager.CheckQuota(tenantID, req.QuotaType, req.Amount)
	if err != nil {
		if IsQuotaExceeded(err) {
			th.writeJSON(w, http.StatusOK, map[string]interface{}{
				"allowed": false,
				"reason":  err.Error(),
			})
		} else {
			th.writeError(w, http.StatusInternalServerError, "failed to check quota", err)
		}
		return
	}

	th.writeJSON(w, http.StatusOK, map[string]interface{}{
		"allowed": true,
	})
}

// Helper methods
func (th *TenantHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (th *TenantHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"error":     message,
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	th.writeJSON(w, status, response)
}

func (th *TenantHandler) getUserID(r *http.Request) string {
	// In a real implementation, extract from JWT token or session
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		return "anonymous"
	}
	return userID
}

func (th *TenantHandler) getRemoteIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Fall back to X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RegisterRoutes registers all tenant management routes
func (th *TenantHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/tenants", th.CreateTenantHandler).Methods("POST")
	router.HandleFunc("/tenants", th.ListTenantsHandler).Methods("GET")
	router.HandleFunc("/tenants/{id}", th.GetTenantHandler).Methods("GET")
	router.HandleFunc("/tenants/{id}", th.UpdateTenantHandler).Methods("PUT")
	router.HandleFunc("/tenants/{id}", th.DeleteTenantHandler).Methods("DELETE")
	router.HandleFunc("/tenants/{id}/quota-usage", th.GetTenantQuotaUsageHandler).Methods("GET")
	router.HandleFunc("/tenants/{id}/check-quota", th.CheckQuotaHandler).Methods("POST")
}

// MiddlewareFunc returns a middleware that validates tenant access
func (th *TenantHandler) MiddlewareFunc() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant ID from path or header
			tenantID := th.extractTenantID(r)
			if tenantID != "" {
				// Validate tenant exists and is active
				config, err := th.tenantManager.GetTenant(TenantID(tenantID))
				if err != nil || config.Status != TenantStatusActive {
					th.writeError(w, http.StatusForbidden, "tenant not available", err)
					return
				}

				// Add tenant to request context
				r.Header.Set("X-Tenant-ID", tenantID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (th *TenantHandler) extractTenantID(r *http.Request) string {
	// Try header first
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	// Try URL path parameter
	if vars := mux.Vars(r); vars != nil {
		if tenantID := vars["tenant_id"]; tenantID != "" {
			return tenantID
		}
	}

	// Try query parameter
	return r.URL.Query().Get("tenant_id")
}
