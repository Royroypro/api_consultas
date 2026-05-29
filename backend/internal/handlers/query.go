package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"api_consultas/internal/middleware"
	"api_consultas/internal/service"
)

type QueryHandler struct {
	db              *sql.DB
	externalService service.ExternalAPIService
}

func NewQueryHandler(db *sql.DB, ext service.ExternalAPIService) *QueryHandler {
	return &QueryHandler{
		db:              db,
		externalService: ext,
	}
}

// getMergeMode consulta la BD para saber si el Modo Fusión está activo
func (h *QueryHandler) getMergeMode() bool {
	var val string
	err := h.db.QueryRow("SELECT value FROM app_settings WHERE key = 'merge_mode'").Scan(&val)
	if err != nil {
		return false
	}
	return val == "true"
}

// HandleDNI procesa la consulta de DNI con caché y fallback
func (h *QueryHandler) HandleDNI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	dni := strings.TrimPrefix(r.URL.Path, "/api/dni/")
	dni = strings.TrimSpace(dni)

	if len(dni) != 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "El DNI debe tener exactamente 8 caracteres numéricos"}`))
		return
	}

	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	startTime := time.Now()

	// 1. Intentar buscar en la caché de la base de datos local
	var dataBytes []byte
	var updatedAt time.Time
	cacheHit := false

	err := h.db.QueryRow("SELECT data, updated_at FROM cache_personas WHERE dni = $1", dni).Scan(&dataBytes, &updatedAt)
	if err == nil {
		// Validar tiempo de caducidad fijo (30 días)
		if time.Since(updatedAt) < 30*24*time.Hour {
			cacheHit = true
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if cacheHit {
		// Cache Hit: Retornar inmediatamente
		w.WriteHeader(http.StatusOK)
		w.Write(dataBytes)

		// Loguear petición asíncronamente
		go h.logRequest(tenantID, "/api/dni/"+dni, time.Since(startTime), "LOCAL_CACHE")
		return
	}

	// Cache Miss o expirado: Realizar Fallback o Fusión según configuración
	var externalData []byte
	var fetchErr error

	if h.getMergeMode() {
		externalData, fetchErr = h.externalService.QueryDNIMerged(r.Context(), dni)
	} else {
		externalData, fetchErr = h.externalService.QueryDNI(r.Context(), dni)
	}

	if fetchErr != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf(`{"error": "No se pudo obtener información del DNI: %s"}`, fetchErr.Error())))
		go h.logRequest(tenantID, "/api/dni/"+dni, time.Since(startTime), "EXTERNAL_API")
		return
	}

	// Retornar inmediatamente los datos formateados
	w.WriteHeader(http.StatusOK)
	w.Write(externalData)

	// Persistir de forma asíncrona la respuesta en caché y registrar logs
	go func(dniVal string, rawData []byte, elapsed time.Duration) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Guardar en la base de datos (UPSERT)
		_, dbErr := h.db.ExecContext(ctx,
			`INSERT INTO cache_personas (dni, data, updated_at) 
			 VALUES ($1, $2, NOW()) 
			 ON CONFLICT (dni) DO UPDATE SET data = $2, updated_at = NOW()`,
			dniVal, rawData)
		if dbErr != nil {
			fmt.Printf("Error al guardar cache_personas: %v\n", dbErr)
		}

		h.logRequestWithCtx(ctx, tenantID, "/api/dni/"+dniVal, elapsed, "EXTERNAL_API")
	}(dni, externalData, time.Since(startTime))
}

// HandleRUC procesa la consulta de RUC con caché y fallback
func (h *QueryHandler) HandleRUC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	ruc := strings.TrimPrefix(r.URL.Path, "/api/ruc/")
	ruc = strings.TrimSpace(ruc)

	if len(ruc) != 11 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "El RUC debe tener exactamente 11 caracteres numéricos"}`))
		return
	}

	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	startTime := time.Now()

	// 1. Intentar buscar en la caché de la base de datos local
	var dataBytes []byte
	var updatedAt time.Time
	cacheHit := false

	err := h.db.QueryRow("SELECT data, updated_at FROM cache_empresas WHERE ruc = $1", ruc).Scan(&dataBytes, &updatedAt)
	if err == nil {
		// Validar tiempo de caducidad fijo (30 días)
		if time.Since(updatedAt) < 30*24*time.Hour {
			cacheHit = true
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if cacheHit {
		// Cache Hit: Retornar inmediatamente
		w.WriteHeader(http.StatusOK)
		w.Write(dataBytes)

		// Loguear petición asíncronamente
		go h.logRequest(tenantID, "/api/ruc/"+ruc, time.Since(startTime), "LOCAL_CACHE")
		return
	}

	// Cache Miss o expirado: Realizar Fallback o Fusión según configuración
	var externalData []byte
	var fetchErr error

	if h.getMergeMode() {
		externalData, fetchErr = h.externalService.QueryRUCMerged(r.Context(), ruc)
	} else {
		externalData, fetchErr = h.externalService.QueryRUC(r.Context(), ruc)
	}

	if fetchErr != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf(`{"error": "No se pudo obtener información del RUC: %s"}`, fetchErr.Error())))
		go h.logRequest(tenantID, "/api/ruc/"+ruc, time.Since(startTime), "EXTERNAL_API")
		return
	}

	// Retornar inmediatamente los datos
	w.WriteHeader(http.StatusOK)
	w.Write(externalData)

	// Persistir de forma asíncrona en caché y registrar logs
	go func(rucVal string, rawData []byte, elapsed time.Duration) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Guardar en la base de datos (UPSERT)
		_, dbErr := h.db.ExecContext(ctx,
			`INSERT INTO cache_empresas (ruc, data, updated_at) 
			 VALUES ($1, $2, NOW()) 
			 ON CONFLICT (ruc) DO UPDATE SET data = $2, updated_at = NOW()`,
			rucVal, rawData)
		if dbErr != nil {
			fmt.Printf("Error al guardar cache_empresas: %v\n", dbErr)
		}

		h.logRequestWithCtx(ctx, tenantID, "/api/ruc/"+rucVal, elapsed, "EXTERNAL_API")
	}(ruc, externalData, time.Since(startTime))
}

func (h *QueryHandler) logRequest(tenantID, endpoint string, elapsed time.Duration, source string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h.logRequestWithCtx(ctx, tenantID, endpoint, elapsed, source)
}

func (h *QueryHandler) logRequestWithCtx(ctx context.Context, tenantID, endpoint string, elapsed time.Duration, source string) {
	if tenantID == "" {
		return
	}
	responseMs := int(elapsed.Milliseconds())
	if responseMs == 0 {
		responseMs = 1
	}

	query := `INSERT INTO api_logs (tenant_id, endpoint, response_time_ms, data_source) VALUES ($1, $2, $3, $4)`
	_, err := h.db.ExecContext(ctx, query, tenantID, endpoint, responseMs, source)
	if err != nil {
		fmt.Printf("Error al insertar log de auditoría: %v\n", err)
	}
}
