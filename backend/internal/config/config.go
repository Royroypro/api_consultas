package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	DecolectaKey string
	ApiPeruKey   string
	DBConnStr    string
	Port         string
	JWTSecret    string
}

func LoadConfig(filePath string) (*Config, error) {
	// Intentar cargar variables desde el archivo si existe
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// Soportar variaciones de nombres de variables en credenciales.env
				keyClean := strings.ReplaceAll(key, " ", "")
				os.Setenv(keyClean, val)
			}
		}
	}

	// Cargar configuración unificada
	// Soportamos: "CLAVE_API_decolecta.com", "CLAVE_API_apiperu.dev", "DATABASE_URL", "JWT_SECRET", etc.
	return &Config{
		DecolectaKey: getEnv("CLAVE_API_decolecta.com", ""),
		ApiPeruKey:   getEnv("CLAVE_API_apiperu.dev", ""),
		DBConnStr:    getEnv("DATABASE_URL", "postgres://postgres:postgres@api_consulta_bd_postgres:5432/api_consultas?sslmode=disable"),
		Port:         getEnv("PORT", "8080"),
		JWTSecret:    getEnv("JWT_SECRET", "super-secret-key-sehuacho-2026"),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
