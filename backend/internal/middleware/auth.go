package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
)

type ContextKey string

const (
	TenantIDKey   ContextKey = "tenant_id"
	TenantNameKey ContextKey = "tenant_name"
)

type AuthMiddleware struct {
	db *sql.DB
}

func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

func (m *AuthMiddleware) Secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("x-api-key")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Cabecera x-api-key ausente"}`))
			return
		}

		// Calcular Hash SHA-256 de la API Key proporcionada
		hasher := sha256.New()
		hasher.Write([]byte(apiKey))
		apiKeyHash := hex.EncodeToString(hasher.Sum(nil))

		var tenantID string
		var tenantName string
		var isActive bool

		query := "SELECT id, name, is_active FROM tenants WHERE api_key_hash = $1"
		err := m.db.QueryRow(query, apiKeyHash).Scan(&tenantID, &tenantName, &isActive)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "x-api-key inválida o no registrada"}`))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Fallo interno de base de datos"}`))
			return
		}

		if !isActive {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "El inquilino (tenant) se encuentra suspendido"}`))
			return
		}

		// Inyectar de forma segura tenant_id y tenant_name en el Context
		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, TenantNameKey, tenantName)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
