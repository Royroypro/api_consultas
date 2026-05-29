package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

type ExternalAPIService interface {
	QueryDNI(ctx context.Context, dni string) (json.RawMessage, error)
	QueryRUC(ctx context.Context, ruc string) (json.RawMessage, error)
}

type providerConfig struct {
	name     string
	apiKey   string
	priority int
	isActive bool
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

// QueryDNI realiza la consulta DNI usando los proveedores activos en el orden de prioridad
func (s *UnifiedAPIService) QueryDNI(ctx context.Context, dni string) (json.RawMessage, error) {
	providers, err := s.loadProviders()
	if err != nil {
		return nil, fmt.Errorf("error al cargar proveedores: %w", err)
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("ningún proveedor externo activo configurado")
	}

	var lastErr error
	for _, p := range providers {
		var data json.RawMessage
		var queryErr error
		switch p.name {
		case "apiperu":
			data, queryErr = s.queryApiPeru(ctx, "dni", dni, p.apiKey)
		case "decolecta":
			data, queryErr = s.queryDecolecta(ctx, "dni", dni, p.apiKey)
		default:
			continue
		}
		if queryErr == nil {
			return data, nil
		}
		fmt.Printf("[%s] DNI error (prioridad %d): %v, intentando siguiente...\n", p.name, p.priority, queryErr)
		lastErr = queryErr
	}

	return nil, fmt.Errorf("fallaron todos los proveedores para DNI: %w", lastErr)
}

// QueryRUC realiza la consulta RUC usando los proveedores activos en el orden de prioridad
func (s *UnifiedAPIService) QueryRUC(ctx context.Context, ruc string) (json.RawMessage, error) {
	providers, err := s.loadProviders()
	if err != nil {
		return nil, fmt.Errorf("error al cargar proveedores: %w", err)
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("ningún proveedor externo activo configurado")
	}

	var lastErr error
	for _, p := range providers {
		var data json.RawMessage
		var queryErr error
		switch p.name {
		case "apiperu":
			data, queryErr = s.queryApiPeru(ctx, "ruc", ruc, p.apiKey)
		case "decolecta":
			data, queryErr = s.queryDecolecta(ctx, "ruc", ruc, p.apiKey)
		default:
			continue
		}
		if queryErr == nil {
			return data, nil
		}
		fmt.Printf("[%s] RUC error (prioridad %d): %v, intentando siguiente...\n", p.name, p.priority, queryErr)
		lastErr = queryErr
	}

	return nil, fmt.Errorf("fallaron todos los proveedores para RUC: %w", lastErr)
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
	url := fmt.Sprintf("https://decolecta.com/api/%s/%s", queryType, value)
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
