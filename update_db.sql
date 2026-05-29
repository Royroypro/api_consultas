CREATE TABLE IF NOT EXISTS provider_configs (
    provider_name VARCHAR(50) PRIMARY KEY,
    api_key VARCHAR(255) NOT NULL,
    priority INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO provider_configs (provider_name, api_key, priority, is_active)
VALUES 
    ('apiperu', 'CaK69kijSiBldmfX1ELWzOQ9x8imvrHNz82a0OzJKqS5lpEj6wHttYVVNegR', 1, TRUE),
    ('decolecta', 'sk_15547.SnBtR5ZeI1nIw9Plku6yfa4PGohUaPQS', 2, TRUE)
ON CONFLICT (provider_name) DO UPDATE 
SET api_key = EXCLUDED.api_key, priority = EXCLUDED.priority, is_active = EXCLUDED.is_active;
