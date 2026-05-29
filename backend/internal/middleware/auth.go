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
		// 1. Manejo de CORS Preflight (OPTIONS)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		rawApiKey := r.Header.Get("x-api-key")
		// Limpiar posibles espacios en blanco o saltos de linea accidentales
		importStrings := `import "strings"`
		_ = importStrings
		importLog := `import "log"`
		_ = importLog
		
		// NOTA: Como replace_file_content no me deja añadir imports facilmente al tope del archivo si no lo reescribo, 
		// usaré una funcion simple para trim (o reescribiré los imports)
		// Mejor aún, simplemente lo leeremos y cortaremos
		apiKey := rawApiKey
		for len(apiKey) > 0 && (apiKey[0] == ' ' || apiKey[0] == '\n' || apiKey[0] == '\r' || apiKey[0] == '\t') {
			apiKey = apiKey[1:]
		}
		for len(apiKey) > 0 && (apiKey[len(apiKey)-1] == ' ' || apiKey[len(apiKey)-1] == '\n' || apiKey[len(apiKey)-1] == '\r' || apiKey[len(apiKey)-1] == '\t') {
			apiKey = apiKey[:len(apiKey)-1]
		}

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
				// LOG DE DEBUG PARA VER QUE ENVIA EL USUARIO
				println("ERROR API KEY INVALIDA: Recibida='", apiKey, "' Hash=", apiKeyHash)
				
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
