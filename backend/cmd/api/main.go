package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"api_consultas/internal/config"
	"api_consultas/internal/database"
	"api_consultas/internal/handlers"
	"api_consultas/internal/middleware"
	"api_consultas/internal/service"
)

func main() {
	log.Println("Iniciando microservicio de consultas multitenant...")

	// 1. Cargar Configuración buscando en múltiples ubicaciones para resiliencia
	configPaths := []string{
		"C:\\Users\\rdrf-\\OneDrive\\Documentos\\api_consultas\\credenciales.env",
		"./credenciales.env",
		"../credenciales.env",
		"/app/credenciales.env",
	}

	var cfg *config.Config
	var err error
	for _, path := range configPaths {
		if _, errExists := os.Stat(path); errExists == nil {
			log.Printf("Cargando credenciales desde: %s\n", path)
			cfg, err = config.LoadConfig(path)
			break
		}
	}

	if cfg == nil || err != nil {
		log.Println("Advertencia: No se encontro credenciales.env en rutas por defecto. Cargando desde entorno nativo.")
		cfg, err = config.LoadConfig("")
		if err != nil {
			log.Fatalf("Error crítico al inicializar configuracion: %v\n", err)
		}
	}

	// Imprimir estado de API Keys cargadas
	if cfg.ApiPeruKey != "" {
		log.Println("✓ API Key de apiperu.dev cargada satisfactoriamente.")
	} else {
		log.Println("⚠ API Key de apiperu.dev NO cargada (las consultas a esta API podrian fallar).")
	}
	if cfg.DecolectaKey != "" {
		log.Println("✓ API Key de decolecta.com cargada satisfactoriamente.")
	} else {
		log.Println("⚠ API Key de decolecta.com NO cargada (las consultas a esta API podrian fallar).")
	}

	// 2. Inicializar Conexión a Base de Datos PostgreSQL
	db, err := database.NewConnectionPool(cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos PostgreSQL: %v\n", err)
	}
	defer db.Close()
	log.Println("✓ Conexión establecida con PostgreSQL (pool inicializado).")

	// 3. Inicializar Servicios y Controladores
	extService := service.NewUnifiedAPIService(db)
	queryHandler := handlers.NewQueryHandler(db, extService)
	adminHandler := handlers.NewAdminHandler(db, cfg.JWTSecret)
	authMiddleware := middleware.NewAuthMiddleware(db)

	// Enrutador multiplexor estándar
	mux := http.NewServeMux()

	// 4. Registrar Rutas de Consulta de Datos (Multitenant mediante Middleware)
	mux.Handle("/api/dni/", authMiddleware.Secure(http.HandlerFunc(queryHandler.HandleDNI)))
	mux.Handle("/api/ruc/", authMiddleware.Secure(http.HandlerFunc(queryHandler.HandleRUC)))

	// 5. Registrar Rutas de Panel Administrativo (Con control de CORS y JWT)
	mux.HandleFunc("/api/admin/login", adminHandler.EnableCORS(adminHandler.Login))
	mux.HandleFunc("/api/admin/dashboard", adminHandler.AuthenticateAdmin(adminHandler.GetDashboardMetrics))

	// Enrutador avanzado manual para sub-rutas CRUD de Tenants
	mux.HandleFunc("/api/admin/tenants", adminHandler.AuthenticateAdmin(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminHandler.ListTenants(w, r)
		} else if r.Method == http.MethodPost {
			adminHandler.CreateTenant(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// Enrutador para operaciones con ID sobre Tenants
	mux.HandleFunc("/api/admin/tenants/", adminHandler.AuthenticateAdmin(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		// El path esperado es /api/admin/tenants/{id}/{action}
		if len(parts) < 5 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Path invalido"}`))
			return
		}

		if len(parts) == 5 && r.Method == http.MethodDelete {
			adminHandler.DeleteTenant(w, r)
			return
		}

		if len(parts) == 6 {
			action := parts[5]
			if action == "rotate-key" && r.Method == http.MethodPost {
				adminHandler.RotateTenantKey(w, r)
				return
			}
			if action == "toggle-status" && r.Method == http.MethodPost {
				adminHandler.ToggleTenantStatus(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	// Rutas de Configuración de Proveedores Externos
	mux.HandleFunc("/api/admin/providers", adminHandler.AuthenticateAdmin(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminHandler.GetProviders(w, r)
		} else if r.Method == http.MethodPut {
			adminHandler.UpdateProviders(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// Servidor HTTP
	serverAddr := ":" + cfg.Port
	log.Printf("Microservicio listo y escuchando en el puerto %s...\n", cfg.Port)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Error al arrancar el servidor HTTP: %v\n", err)
	}
}
