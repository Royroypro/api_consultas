package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ExternalAPIService interface {
	QueryDNI(ctx context.Context, dni string) (json.RawMessage, error)
	QueryRUC(ctx context.Context, ruc string) (json.RawMessage, error)
}

type UnifiedAPIService struct {
	apiPeruKey   string
	decolectaKey string
	client       *http.Client
}

func NewUnifiedAPIService(apiPeruKey, decolectaKey string) *UnifiedAPIService {
	return &UnifiedAPIService{
		apiPeruKey:   apiPeruKey,
		decolectaKey: decolectaKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueryDNI realiza la consulta externa intentando primero apiperu.dev y luego decolecta.com como fallback
func (s *UnifiedAPIService) QueryDNI(ctx context.Context, dni string) (json.RawMessage, error) {
	if s.apiPeruKey != "" {
		data, err := s.queryApiPeru(ctx, "dni", dni)
		if err == nil {
			return data, nil
		}
		// Loguear error de apiperu.dev y continuar al fallback
		fmt.Printf("apiperu.dev DNI error: %v, intentando fallback a decolecta.com\n", err)
	}

	if s.decolectaKey != "" {
		data, err := s.queryDecolecta(ctx, "dni", dni)
		if err == nil {
			return data, nil
		}
		return nil, fmt.Errorf("fallaron todas las APIs externas para DNI: %w", err)
	}

	return nil, fmt.Errorf("ninguna API externa configurada para DNI")
}

// QueryRUC realiza la consulta externa intentando primero apiperu.dev y luego decolecta.com como fallback
func (s *UnifiedAPIService) QueryRUC(ctx context.Context, ruc string) (json.RawMessage, error) {
	if s.apiPeruKey != "" {
		data, err := s.queryApiPeru(ctx, "ruc", ruc)
		if err == nil {
			return data, nil
		}
		fmt.Printf("apiperu.dev RUC error: %v, intentando fallback a decolecta.com\n", err)
	}

	if s.decolectaKey != "" {
		data, err := s.queryDecolecta(ctx, "ruc", ruc)
		if err == nil {
			return data, nil
		}
		return nil, fmt.Errorf("fallaron todas las APIs externas para RUC: %w", err)
	}

	return nil, fmt.Errorf("ninguna API externa configurada para RUC")
}

func (s *UnifiedAPIService) queryApiPeru(ctx context.Context, queryType, value string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://apiperu.dev/api/%s/%s", queryType, value)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiPeruKey)
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

func (s *UnifiedAPIService) queryDecolecta(ctx context.Context, queryType, value string) (json.RawMessage, error) {
	// Adaptador estándar para decolecta.com (ejemplo de endpoint habitual)
	url := fmt.Sprintf("https://decolecta.com/api/%s/%s", queryType, value)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.decolectaKey)
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
