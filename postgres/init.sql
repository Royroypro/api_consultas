-- Habilitar extensión para UUIDs seguros
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabla de Inquilinos Autorizados (SaaS)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    api_key_hash VARCHAR(64) UNIQUE NOT NULL, -- Almacenará el hash SHA-256 de la API Key entregada al cliente
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tipo ENUM para determinar el origen de los datos
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'data_source_enum') THEN
        CREATE TYPE data_source_enum AS ENUM ('LOCAL_CACHE', 'EXTERNAL_API');
    END IF;
END$$;

-- Registro de Auditoría y Métricas (api_logs)
CREATE TABLE IF NOT EXISTS api_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    endpoint VARCHAR(255) NOT NULL,
    response_time_ms INTEGER NOT NULL,
    data_source data_source_enum NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Caché para Personas (DNI)
CREATE TABLE IF NOT EXISTS cache_personas (
    dni VARCHAR(8) PRIMARY KEY,
    data JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Caché para Empresas (RUC)
CREATE TABLE IF NOT EXISTS cache_empresas (
    ruc VARCHAR(11) PRIMARY KEY,
    data JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Administradores de Infraestructura
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Panel de Administración: Dashboard (Rendimiento)
CREATE VIEW dashboard_metrics AS
SELECT
    COUNT(*) as total_requests,
    COUNT(CASE WHEN data_source = 'LOCAL_CACHE' THEN 1 END) as cache_hits,
    AVG(response_time_ms) as avg_response_time_ms
FROM api_logs;

-- Configuración Dinámica de Proveedores Externos
CREATE TABLE IF NOT EXISTS provider_configs (
    provider_name VARCHAR(50) PRIMARY KEY,
    api_key VARCHAR(255) NOT NULL,
    priority INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Proveedores por defecto
INSERT INTO provider_configs (provider_name, api_key, priority, is_active)
VALUES 
    ('apiperu', 'CaK69kijSiBldmfX1ELWzOQ9x8imvrHNz82a0OzJKqS5lpEj6wHttYVVNegR', 1, TRUE),
    ('decolecta', 'sk_15547.SnBtR5ZeI1nIw9Plku6yfa4PGohUaPQS', 2, TRUE)
ON CONFLICT (provider_name) DO NOTHING;

-- Índices de Rendimiento y Agregaciones del Dashboard
CREATE INDEX IF NOT EXISTS idx_api_logs_tenant ON api_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_logs_source ON api_logs(data_source);
CREATE INDEX IF NOT EXISTS idx_api_logs_created ON api_logs(created_at);

-- Inquilino Demo de Bootstrapping
-- API Key original: 'tc_master_key_2026'
-- SHA-256 Hash: '499208dba47c01f24bb63156c24103bc4768158f68815ffdb293f54d2e65eef7'
INSERT INTO tenants (id, name, api_key_hash, is_active)
VALUES (
    'a3b1c2d3-e4f5-6a7b-8c9d-0e1f2a3b4c5d',
    'Cliente Demo SaaS',
    '499208dba47c01f24bb63156c24103bc4768158f68815ffdb293f54d2e65eef7',
    TRUE
) ON CONFLICT (name) DO NOTHING;

-- Administrador por defecto de Bootstrapping
-- Usuario: 'admin'
-- Contraseña original: 'admin2026'
-- BCrypt Hash para 'admin2026': '$2a$10$wE1m.hVdSw82iFqf9zP65e4xIlgfG/Yy7L1gZ2o2Nq2.u/b7jYee.'
INSERT INTO admins (id, username, password_hash)
VALUES (
    'b4c2d3e4-f5a6-7b8c-9d0e-1f2a3b4c5d6e',
    'admin',
    '$2a$10$wE1m.hVdSw82iFqf9zP65e4xIlgfG/Yy7L1gZ2o2Nq2.u/b7jYee.'
) ON CONFLICT (username) DO NOTHING;
