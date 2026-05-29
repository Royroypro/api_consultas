package models

import "time"

type Tenant struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	APIKeyHash string    `json:"-"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type APILog struct {
	ID             int64     `json:"id"`
	TenantID       string    `json:"tenant_id"`
	TenantName     string    `json:"tenant_name,omitempty"`
	Endpoint       string    `json:"endpoint"`
	ResponseTimeMs int       `json:"response_time_ms"`
	DataSource     string    `json:"data_source"`
	CreatedAt      time.Time `json:"created_at"`
}

type Admin struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Representa una consulta genérica
type QueryResult struct {
	Source string      `json:"source"`
	Data   interface{} `json:"data"`
}
