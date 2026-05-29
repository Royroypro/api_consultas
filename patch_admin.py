with open("backend/internal/handlers/admin.go", "a", encoding="utf-8") as f:
    f.write("""
func (h *AdminHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")

rows, err := h.db.Query("SELECT provider_name, api_key, priority, is_active FROM provider_configs ORDER BY priority ASC")
if err != nil {
ternalServerError)
te(`{"error": "Error al consultar proveedores"}`))

}
defer rows.Close()

var providers []map[string]interface{}
for rows.Next() {
name, key string
priority int
active bool
err := rows.Scan(&name, &key, &priority, &active); err != nil {
tinue
= append(providers, map[string]interface{}{
ame": name,
":       key,
":      priority,
    active,
   if providers == nil {
        providers = []map[string]interface{}{}
    }

json.NewEncoder(w).Encode(providers)
}

func (h *AdminHandler) UpdateProviders(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")

var reqs []struct {
ame string `json:"provider_name"`
       string `json:"api_key"`
     int    `json:"priority"`
    bool   `json:"is_active"`
}

if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
uest)
te(`{"error": "Payload inválido"}`))

}

tx, err := h.db.Begin()
if err != nil {
ternalServerError)
te(`{"error": "Error al iniciar transacción"}`))

}

for _, p := range reqs {
err := tx.Exec("UPDATE provider_configs SET api_key = $1, priority = $2, is_active = $3 WHERE provider_name = $4", p.APIKey, p.Priority, p.IsActive, p.ProviderName)
err != nil {
ternalServerError)
te(`{"error": "Error al actualizar proveedor"}`))

err := tx.Commit(); err != nil {
ternalServerError)
te(`{"error": "Error al guardar transacción"}`))

}

w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"message": "Proveedores actualizados"}`))
}
""")
print("Patched admin.go")
