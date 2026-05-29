package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewAdminHandler(db *sql.DB, secret string) *AdminHandler {
	return &AdminHandler{
		db:        db,
		jwtSecret: []byte(secret),
	}
}

// EnableCORS es un helper para inyectar cabeceras CORS en todas las peticiones
func (h *AdminHandler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// AuthenticateAdmin es un middleware para validar JWT tokens del panel de control
func (h *AdminHandler) AuthenticateAdmin(next http.HandlerFunc) http.HandlerFunc {
	return h.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Acceso denegado: Token ausente o incorrecto"}`))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metodo de firma invalido: %v", t.Header["alg"])
			}
			return h.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Acceso denegado: Token invalido o expirado"}`))
			return
		}

		next(w, r)
	})
}

// Login maneja la autenticación del administrador de infraestructura
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Cuerpo de peticion invalido"}`))
		return
	}

	var adminID string
	var passwordHash string

	err := h.db.QueryRow("SELECT id, password_hash FROM admins WHERE username = $1", req.Username).Scan(&adminID, &passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Usuario o contraseña incorrectos"}`))
		return
	}

	// Comparar bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Usuario o contraseña incorrectos"}`))
		return
	}

	// Crear JWT Token
	claims := jwt.MapClaims{
		"sub": adminID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Fallo al generar el token JWT"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token":    tokenString,
		"username": req.Username,
	})
}

// GetDashboardMetrics recopila métricas y ahorros económicos para el dashboard
func (h *AdminHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var totalRequests int64
	var cacheHits int64
	var apiRequests int64

	// Contar logs
	h.db.QueryRow("SELECT COUNT(*) FROM api_logs").Scan(&totalRequests)
	h.db.QueryRow("SELECT COUNT(*) FROM api_logs WHERE data_source = 'LOCAL_CACHE'").Scan(&cacheHits)
	h.db.QueryRow("SELECT COUNT(*) FROM api_logs WHERE data_source = 'EXTERNAL_API'").Scan(&apiRequests)

	// Ahorro económico (S/. 0.05 por cada hit en cache)
	economicSavings := float64(cacheHits) * 0.05

	// Tiempo de respuesta promedio
	var avgResponseTimeCache float64
	var avgResponseTimeAPI float64
	h.db.QueryRow("SELECT COALESCE(AVG(response_time_ms), 0) FROM api_logs WHERE data_source = 'LOCAL_CACHE'").Scan(&avgResponseTimeCache)
	h.db.QueryRow("SELECT COALESCE(AVG(response_time_ms), 0) FROM api_logs WHERE data_source = 'EXTERNAL_API'").Scan(&avgResponseTimeAPI)

	// Registros en tablas de cache
	var cachedPersonas int64
	var cachedEmpresas int64
	h.db.QueryRow("SELECT COUNT(*) FROM cache_personas").Scan(&cachedPersonas)
	h.db.QueryRow("SELECT COUNT(*) FROM cache_empresas").Scan(&cachedEmpresas)

	// Historial de peticiones en los últimos 7 días
	historyRows, err := h.db.Query(`
		SELECT DATE(created_at) as day, data_source, COUNT(*) as qty
		FROM api_logs 
		WHERE created_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE(created_at), data_source
		ORDER BY DATE(created_at) ASC
	`)
	
	type DailyStat struct {
		Day        string `json:"day"`
		LocalCache int64  `json:"local_cache"`
		External   int64  `json:"external"`
	}
	
	dailyStats := make([]DailyStat, 0)
	dayMap := make(map[string]*DailyStat)

	if err == nil {
		defer historyRows.Close()
		for historyRows.Next() {
			var day time.Time
			var source string
			var qty int64
			if err := historyRows.Scan(&day, &source, &qty); err == nil {
				dayStr := day.Format("2006-01-02")
				stat, ok := dayMap[dayStr]
				if !ok {
					stat = &DailyStat{Day: dayStr}
					dayMap[dayStr] = stat
					dailyStats = append(dailyStats, *stat)
				}
				if source == "LOCAL_CACHE" {
					stat.LocalCache = qty
				} else {
					stat.External = qty
				}
			}
		}
	}

	// Reconstruir lista ordenada final
	finalStats := make([]DailyStat, 0)
	for _, ds := range dailyStats {
		actual := dayMap[ds.Day]
		finalStats = append(finalStats, *actual)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_requests":          totalRequests,
		"cache_hits":              cacheHits,
		"external_api_requests":   apiRequests,
		"economic_savings_pen":    economicSavings,
		"avg_response_time_cache": avgResponseTimeCache,
		"avg_response_time_api":   avgResponseTimeAPI,
		"cached_personas":         cachedPersonas,
		"cached_empresas":         cachedEmpresas,
		"daily_stats":             finalStats,
	})
}

// CRUD Tenants: Listar todos
func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := h.db.Query(`
		SELECT t.id, t.name, t.is_active, t.created_at, COUNT(l.id) as total_queries 
		FROM tenants t
		LEFT JOIN api_logs l ON t.id = l.tenant_id
		GROUP BY t.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "Error al listar tenants: %v"}`, err)))
		return
	}
	defer rows.Close()

	type TenantWithStats struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		IsActive     bool      `json:"is_active"`
		CreatedAt    time.Time `json:"created_at"`
		TotalQueries int64     `json:"total_queries"`
	}

	tenants := make([]TenantWithStats, 0)
	for rows.Next() {
		var t TenantWithStats
		if err := rows.Scan(&t.ID, &t.Name, &t.IsActive, &t.CreatedAt, &t.TotalQueries); err == nil {
			tenants = append(tenants, t)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenants)
}

// CRUD Tenants: Crear nuevo inquilino
func (h *AdminHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Nombre de inquilino requerido"}`))
		return
	}

	// Generar API Key aleatoria de alta entropía (tc_...)
	rawKey, err := generateRandomAPIKey()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Fallo al generar API Key criptografica"}`))
		return
	}

	// Generar Hash SHA-256 de la API Key para persistir de manera segura
	hasher := sha256.New()
	hasher.Write([]byte(rawKey))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	var newID string
	query := "INSERT INTO tenants (name, api_key_hash) VALUES ($1, $2) RETURNING id"
	err = h.db.QueryRow(query, req.Name, keyHash).Scan(&newID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error": "No se pudo crear tenant (posible nombre duplicado): %v"}`, err)))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":          newID,
		"name":        req.Name,
		"raw_api_key": rawKey, // Entregada SOLO UNA VEZ para almacenamiento por parte del cliente
	})
}

// CRUD Tenants: Rotar API Key
func (h *AdminHandler) RotateTenantKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Extraer ID del path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "ID de inquilino requerido"}`))
		return
	}
	tenantID := parts[4] // Ej. /api/admin/tenants/{id}/rotate-key

	rawKey, err := generateRandomAPIKey()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error al generar la nueva clave"}`))
		return
	}

	hasher := sha256.New()
	hasher.Write([]byte(rawKey))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	query := "UPDATE tenants SET api_key_hash = $1 WHERE id = $2"
	result, err := h.db.Exec(query, keyHash, tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Fallo al rotar la clave en BD"}`))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Inquilino no encontrado"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"raw_api_key": rawKey,
	})
}

// CRUD Tenants: Activar / Desactivar Tenant
func (h *AdminHandler) ToggleTenantStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "ID de inquilino requerido"}`))
		return
	}
	tenantID := parts[4] // Ej. /api/admin/tenants/{id}/toggle-status

	// Leer estado del body
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Estado is_active requerido"}`))
		return
	}

	query := "UPDATE tenants SET is_active = $1 WHERE id = $2"
	result, err := h.db.Exec(query, req.IsActive, tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Fallo al actualizar estado en BD"}`))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Inquilino no encontrado"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"is_active": req.IsActive,
	})
}

// CRUD Tenants: Eliminar inquilino
func (h *AdminHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "Metodo no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "ID de inquilino requerido"}`))
		return
	}
	tenantID := parts[4] // Ej. /api/admin/tenants/{id}

	query := "DELETE FROM tenants WHERE id = $1"
	result, err := h.db.Exec(query, tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Fallo al eliminar inquilino en BD"}`))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Inquilino no encontrado"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func generateRandomAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "tc_" + hex.EncodeToString(bytes), nil
}


func (h *AdminHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := h.db.Query("SELECT provider_name, api_key, priority, is_active FROM provider_configs ORDER BY priority ASC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error al consultar proveedores"}`))
		return
	}
	defer rows.Close()

	var providers []map[string]interface{}
	for rows.Next() {
		var name, key string
		var priority int
		var active bool
		if err := rows.Scan(&name, &key, &priority, &active); err != nil {
			continue
		}
		providers = append(providers, map[string]interface{}{
			"provider_name": name,
			"api_key":       key,
			"priority":      priority,
			"is_active":     active,
		})
	}
	if providers == nil {
		providers = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(providers)
}

func (h *AdminHandler) UpdateProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqs []struct {
		ProviderName string `json:"provider_name"`
		APIKey       string `json:"api_key"`
		Priority     int    `json:"priority"`
		IsActive     bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Payload inválido"}`))
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error al iniciar transacción"}`))
		return
	}

	for _, p := range reqs {
		_, execErr := tx.Exec(
			"UPDATE provider_configs SET api_key = $1, priority = $2, is_active = $3 WHERE provider_name = $4",
			p.APIKey, p.Priority, p.IsActive, p.ProviderName,
		)
		if execErr != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Error al actualizar proveedor"}`))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error al guardar transacción"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Proveedores actualizados"}`))
}
