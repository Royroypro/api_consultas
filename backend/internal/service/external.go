package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type ExternalAPIService interface {
	QueryDNI(ctx context.Context, dni string) (json.RawMessage, error)
	QueryRUC(ctx context.Context, ruc string) (json.RawMessage, error)
	QueryDNIMerged(ctx context.Context, dni string) (json.RawMessage, error)
	QueryRUCMerged(ctx context.Context, ruc string) (json.RawMessage, error)
}

type providerConfig struct {
	name     string
	apiKey   string
	priority int
	isActive bool
}

type providerResult struct {
	name string
	data json.RawMessage
	err  error
}

type UnifiedAPIService struct {
	db     *sql.DB
	client *http.Client
}

func NewUnifiedAPIService(db *sql.DB) *UnifiedAPIService {
	return &UnifiedAPIService{
		db: db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// loadProviders carga los proveedores activos desde la BD, ordenados por prioridad
func (s *UnifiedAPIService) loadProviders() ([]providerConfig, error) {
	rows, err := s.db.Query("SELECT provider_name, api_key, priority, is_active FROM provider_configs ORDER BY priority ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []providerConfig
	for rows.Next() {
		var p providerConfig
		if err := rows.Scan(&p.name, &p.apiKey, &p.priority, &p.isActive); err != nil {
			continue
		}
		if p.isActive {
			providers = append(providers, p)
		}
	}

	sort.Slice(providers, func(i, j int) bool {
		return providers[i].priority < providers[j].priority
	})
	return providers, nil
}

// callProvider realiza la petición al proveedor correcto según su nombre
func (s *UnifiedAPIService) callProvider(ctx context.Context, p providerConfig, queryType, value string) (json.RawMessage, error) {
	switch p.name {
	case "apiperu":
		return s.queryApiPeru(ctx, queryType, value, p.apiKey)
	case "decolecta":
		return s.queryDecolecta(ctx, queryType, value, p.apiKey)
	default:
		return nil, fmt.Errorf("proveedor desconocido: %s", p.name)
	}
}

// QueryDNI — Modo Fallback: usa los proveedores en orden de prioridad
func (s *UnifiedAPIService) QueryDNI(ctx context.Context, dni string) (json.RawMessage, error) {
	return s.queryFallback(ctx, "dni", dni)
}

// QueryRUC — Modo Fallback: usa los proveedores en orden de prioridad
func (s *UnifiedAPIService) QueryRUC(ctx context.Context, ruc string) (json.RawMessage, error) {
	return s.queryFallback(ctx, "ruc", ruc)
}

// QueryDNIMerged — Modo Fusión: consulta todas las APIs en paralelo y fusiona
func (s *UnifiedAPIService) QueryDNIMerged(ctx context.Context, dni string) (json.RawMessage, error) {
	return s.queryMerged(ctx, "dni", dni)
}

// QueryRUCMerged — Modo Fusión: consulta todas las APIs en paralelo y fusiona
func (s *UnifiedAPIService) QueryRUCMerged(ctx context.Context, ruc string) (json.RawMessage, error) {
	return s.queryMerged(ctx, "ruc", ruc)
}

func (s *UnifiedAPIService) queryFallback(ctx context.Context, queryType, value string) (json.RawMessage, error) {
	providers, err := s.loadProviders()
	if err != nil {
		return nil, fmt.Errorf("error al cargar proveedores: %w", err)
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("ningún proveedor externo activo configurado")
	}

	var lastErr error
	for _, p := range providers {
		data, queryErr := s.callProvider(ctx, p, queryType, value)
		if queryErr == nil {
			return data, nil
		}
		fmt.Printf("[%s] %s error (prioridad %d): %v, intentando siguiente...\n", p.name, queryType, p.priority, queryErr)
		lastErr = queryErr
	}

	return nil, fmt.Errorf("fallaron todos los proveedores para %s: %w", queryType, lastErr)
}

func (s *UnifiedAPIService) queryMerged(ctx context.Context, queryType, value string) (json.RawMessage, error) {
	providers, err := s.loadProviders()
	if err != nil {
		return nil, fmt.Errorf("error al cargar proveedores: %w", err)
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("ningún proveedor externo activo configurado")
	}

	// Lanzar todas las consultas en paralelo
	results := make([]providerResult, len(providers))
	var wg sync.WaitGroup
	for i, p := range providers {
		wg.Add(1)
		go func(idx int, prov providerConfig) {
			defer wg.Done()
			data, err := s.callProvider(ctx, prov, queryType, value)
			results[idx] = providerResult{name: prov.name, data: data, err: err}
		}(i, p)
	}
	wg.Wait()

	// Recolectar los exitosos
	var successfulResults []providerResult
	for _, r := range results {
		if r.err == nil && r.data != nil {
			successfulResults = append(successfulResults, r)
		} else if r.err != nil {
			fmt.Printf("[FUSION] %s error: %v\n", r.name, r.err)
		}
	}

	if len(successfulResults) == 0 {
		return nil, fmt.Errorf("fallaron todos los proveedores en modo fusión para %s", queryType)
	}

	// Si solo uno tuvo éxito, devolver directamente sin fusión
	if len(successfulResults) == 1 {
		var base map[string]interface{}
		if err := json.Unmarshal(successfulResults[0].data, &base); err != nil {
			return successfulResults[0].data, nil
		}
		base["_sources"] = []string{successfulResults[0].name}
		return json.Marshal(base)
	}

	// Fusionar todos los resultados exitosos
	merged, sources := mergeProviderResults(successfulResults)
	merged["_sources"] = sources

	return json.Marshal(merged)
}

// mergeProviderResults fusiona los JSON de múltiples proveedores.
// Si dos proveedores tienen el mismo campo con valores iguales → un solo campo.
// Si los valores difieren → dos campos con sufijo del proveedor (ej: nombre_apiperu, nombre_decolecta).
func mergeProviderResults(results []providerResult) (map[string]interface{}, []string) {
	merged := make(map[string]interface{})
	sources := make([]string, 0, len(results))

	// Primero, parsear todos
	parsed := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		var m map[string]interface{}
		if err := json.Unmarshal(r.data, &m); err == nil {
			parsed = append(parsed, m)
			sources = append(sources, r.name)
		}
	}

	// Construir un mapa de qué proveedor tiene qué campo
	fieldProviders := make(map[string][]struct {
		name  string
		value interface{}
	})

	for i, m := range parsed {
		provName := results[i].name
		for k, v := range m {
			fieldProviders[k] = append(fieldProviders[k], struct {
				name  string
				value interface{}
			}{name: provName, value: v})
		}
	}

	// Fusionar campo por campo
	for field, entries := range fieldProviders {
		if len(entries) == 1 {
			// Solo un proveedor tiene este campo
			merged[field] = entries[0].value
		} else {
			// Múltiples proveedores — comparar valores
			allEqual := true
			for i := 1; i < len(entries); i++ {
				v1, _ := json.Marshal(entries[0].value)
				v2, _ := json.Marshal(entries[i].value)
				if string(v1) != string(v2) {
					allEqual = false
					break
				}
			}

			if allEqual {
				// Valores idénticos → un solo campo
				merged[field] = entries[0].value
			} else {
				// Valores distintos → campos separados con sufijo del proveedor
				for _, entry := range entries {
					merged[field+"_"+entry.name] = entry.value
				}
			}
		}
	}

	return merged, sources
}

func (s *UnifiedAPIService) queryApiPeru(ctx context.Context, queryType, value, apiKey string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://apiperu.dev/api/%s/%s", queryType, value)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apiperu dev retorno estatus: %d", resp.StatusCode)
	}

	var payload struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if !payload.Success {
		return nil, fmt.Errorf("apiperu retorno error en el payload success=false")
	}

	return payload.Data, nil
}

func (s *UnifiedAPIService) queryDecolecta(ctx context.Context, queryType, value, apiKey string) (json.RawMessage, error) {
	var url string
	if queryType == "dni" {
		url = fmt.Sprintf("https://api.decolecta.com/v1/reniec/dni?numero=%s", value)
	} else if queryType == "ruc" {
		url = fmt.Sprintf("https://api.decolecta.com/v1/sunat/ruc/full?numero=%s", value)
	} else {
		return nil, fmt.Errorf("tipo de consulta no soportado para decolecta: %s", queryType)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("decolecta retorno estatus: %d", resp.StatusCode)
	}

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
