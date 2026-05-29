with open("backend/internal/handlers/admin.go", "a", encoding="utf-8") as f:
    f.write('''
func (h *AdminHandler) ListCache(w http.ResponseWriter, r *http.Request, cacheType string) {
\tw.Header().Set("Content-Type", "application/json")
\tvar tableName, idColumn string
\tif cacheType == "dni" {
\t\ttableName = "cache_personas"
\t\tidColumn = "dni"
\t} else if cacheType == "ruc" {
\t\ttableName = "cache_empresas"
\t\tidColumn = "ruc"
\t} else {
\t\tw.WriteHeader(http.StatusBadRequest)
\t\tw.Write([]byte(`{"error": "Tipo de cache invalido"}`))
\t\treturn
\t}

\tquery := fmt.Sprintf("SELECT %s, data, updated_at FROM %s ORDER BY updated_at DESC LIMIT 100", idColumn, tableName)
\trows, err := h.db.Query(query)
\tif err != nil {
\t\tw.WriteHeader(http.StatusInternalServerError)
\t\tw.Write([]byte(`{"error": "Error al consultar cache"}`))
\t\treturn
\t}
\tdefer rows.Close()

\tvar results []map[string]interface{}
\tfor rows.Next() {
\t\tvar id string
\t\tvar dataBytes []byte
\t\tvar updatedAt time.Time
\t\tif err := rows.Scan(&id, &dataBytes, &updatedAt); err != nil {
\t\t\tcontinue
\t\t}
\t\tvar dataJSON interface{}
\t\tjson.Unmarshal(dataBytes, &dataJSON)
\t\t
\t\tresults = append(results, map[string]interface{}{
\t\t\t"id": id,
\t\t\t"data": dataJSON,
\t\t\t"updated_at": updatedAt.Format(time.RFC3339),
\t\t})
\t}
\tif results == nil {
\t\tresults = []map[string]interface{}{}
\t}
\tjson.NewEncoder(w).Encode(results)
}

func (h *AdminHandler) UpdateCache(w http.ResponseWriter, r *http.Request, cacheType, id string) {
\tw.Header().Set("Content-Type", "application/json")
\tvar tableName, idColumn string
\tif cacheType == "dni" {
\t\ttableName = "cache_personas"
\t\tidColumn = "dni"
\t} else if cacheType == "ruc" {
\t\ttableName = "cache_empresas"
\t\tidColumn = "ruc"
\t} else {
\t\tw.WriteHeader(http.StatusBadRequest)
\t\tw.Write([]byte(`{"error": "Tipo de cache invalido"}`))
\t\treturn
\t}

\tvar payload interface{}
\tif err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
\t\tw.WriteHeader(http.StatusBadRequest)
\t\tw.Write([]byte(`{"error": "Payload JSON invalido"}`))
\t\treturn
\t}

\tdataBytes, _ := json.Marshal(payload)
\tquery := fmt.Sprintf("UPDATE %s SET data = $1, updated_at = NOW() WHERE %s = $2", tableName, idColumn)
\tres, err := h.db.Exec(query, dataBytes, id)
\tif err != nil {
\t\tw.WriteHeader(http.StatusInternalServerError)
\t\tw.Write([]byte(`{"error": "Error al actualizar cache"}`))
\t\treturn
\t}
\trowsAffected, _ := res.RowsAffected()
\tif rowsAffected == 0 {
\t\tw.WriteHeader(http.StatusNotFound)
\t\tw.Write([]byte(`{"error": "Registro no encontrado"}`))
\t\treturn
\t}
\tw.WriteHeader(http.StatusOK)
\tw.Write([]byte(`{"message": "Registro actualizado exitosamente"}`))
}

func (h *AdminHandler) DeleteCache(w http.ResponseWriter, r *http.Request, cacheType, id string) {
\tw.Header().Set("Content-Type", "application/json")
\tvar tableName, idColumn string
\tif cacheType == "dni" {
\t\ttableName = "cache_personas"
\t\tidColumn = "dni"
\t} else if cacheType == "ruc" {
\t\ttableName = "cache_empresas"
\t\tidColumn = "ruc"
\t} else {
\t\tw.WriteHeader(http.StatusBadRequest)
\t\tw.Write([]byte(`{"error": "Tipo de cache invalido"}`))
\t\treturn
\t}

\tquery := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tableName, idColumn)
\tres, err := h.db.Exec(query, id)
\tif err != nil {
\t\tw.WriteHeader(http.StatusInternalServerError)
\t\tw.Write([]byte(`{"error": "Error al eliminar de cache"}`))
\t\treturn
\t}
\trowsAffected, _ := res.RowsAffected()
\tif rowsAffected == 0 {
\t\tw.WriteHeader(http.StatusNotFound)
\t\tw.Write([]byte(`{"error": "Registro no encontrado"}`))
\t\treturn
\t}
\tw.WriteHeader(http.StatusOK)
\tw.Write([]byte(`{"message": "Registro eliminado exitosamente"}`))
}
''')
print("Cache handlers appended.")
