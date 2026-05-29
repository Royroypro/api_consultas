import re

with open("frontend/index.html", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Add Sidebar button
old_sidebar_btn = """                    <button @click="currentTab = 'providers'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'providers' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="settings-2" class="w-4 h-4"></i>
                        Configurar APIs
                    </button>"""

new_sidebar_btn = old_sidebar_btn + """
                    <button @click="currentTab = 'cache'; fetchCache()" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'cache' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="database" class="w-4 h-4"></i>
                        Gestión de Datos
                    </button>"""
content = content.replace(old_sidebar_btn, new_sidebar_btn, 1)

# 2. Update Header Title and Description
old_title_expr = "currentTab === 'pricing' ? 'Planes de Precios' : 'Configurar Proveedores Externos' }}"
new_title_expr = "currentTab === 'pricing' ? 'Planes de Precios' : currentTab === 'providers' ? 'Configurar Proveedores Externos' : 'Base de Datos (Caché Local)' }}"
content = content.replace(old_title_expr, new_title_expr, 1)

old_desc_expr = "currentTab === 'pricing' ? 'Revisa nuestras tarifas y límites mensuales en Soles (PEN).' : 'Administra las claves y el orden de prioridad de las APIs externas.' }}"
new_desc_expr = "currentTab === 'pricing' ? 'Revisa nuestras tarifas y límites mensuales en Soles (PEN).' : currentTab === 'providers' ? 'Administra las claves y el orden de prioridad de las APIs externas.' : 'Visualiza, edita o elimina los DNIs y RUCs guardados en la plataforma.' }}"
content = content.replace(old_desc_expr, new_desc_expr, 1)

# 3. Add Cache Section UI
old_marker = "                    <!-- ==================== SECCIÓN 6: CONFIGURAR APIS ==================== -->"
cache_section = """
                    <!-- ==================== SECCIÓN 7: GESTIÓN DE CACHÉ ==================== -->
                    <div v-if="currentTab === 'cache'" class="space-y-6">
                        <div class="flex gap-4 mb-6">
                            <button @click="cacheType = 'dni'; fetchCache()" class="px-6 py-2.5 rounded-xl font-bold text-sm transition-all" :class="cacheType === 'dni' ? 'bg-violet-600 text-white' : 'bg-dark-900 text-zinc-400 hover:text-white border border-zinc-800'">
                                Ver Personas (DNI)
                            </button>
                            <button @click="cacheType = 'ruc'; fetchCache()" class="px-6 py-2.5 rounded-xl font-bold text-sm transition-all" :class="cacheType === 'ruc' ? 'bg-violet-600 text-white' : 'bg-dark-900 text-zinc-400 hover:text-white border border-zinc-800'">
                                Ver Empresas (RUC)
                            </button>
                        </div>
                        
                        <div class="glass-panel rounded-3xl border border-zinc-800 overflow-hidden">
                            <div v-if="cacheLoading" class="text-center py-20 text-zinc-500 text-sm">Cargando base de datos...</div>
                            <table v-else class="w-full text-left">
                                <thead>
                                    <tr class="border-b border-zinc-800 bg-dark-900/30">
                                        <th class="px-6 py-4 text-xs font-extrabold text-zinc-400 uppercase tracking-widest">{{ cacheType === 'dni' ? 'DNI' : 'RUC' }}</th>
                                        <th class="px-6 py-4 text-xs font-extrabold text-zinc-400 uppercase tracking-widest">Vista Previa</th>
                                        <th class="px-6 py-4 text-xs font-extrabold text-zinc-400 uppercase tracking-widest text-right">Última Consulta</th>
                                        <th class="px-6 py-4 text-xs font-extrabold text-zinc-400 uppercase tracking-widest text-center">Acciones</th>
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-zinc-800/50">
                                    <tr v-if="cacheData.length === 0">
                                        <td colspan="4" class="text-center py-10 text-zinc-500 text-sm">No hay registros guardados.</td>
                                    </tr>
                                    <tr v-for="item in cacheData" :key="item.id" class="hover:bg-zinc-800/20 transition-colors">
                                        <td class="px-6 py-4 text-sm font-bold text-white">{{ item.id }}</td>
                                        <td class="px-6 py-4 text-xs text-zinc-400 font-mono truncate max-w-xs">{{ JSON.stringify(item.data).substring(0, 50) }}...</td>
                                        <td class="px-6 py-4 text-xs text-zinc-500 text-right">{{ new Date(item.updated_at).toLocaleString() }}</td>
                                        <td class="px-6 py-4 text-center space-x-2 flex justify-center">
                                            <button @click="openEditCache(item)" class="p-2 bg-dark-900 border border-zinc-700 rounded-lg text-violet-400 hover:bg-violet-600 hover:text-white transition-all" title="Ver / Editar JSON">
                                                <i data-lucide="edit-3" class="w-4 h-4"></i>
                                            </button>
                                            <button @click="deleteCache(item.id)" class="p-2 bg-dark-900 border border-zinc-700 rounded-lg text-red-400 hover:bg-red-500 hover:text-white transition-all" title="Eliminar de Caché">
                                                <i data-lucide="trash-2" class="w-4 h-4"></i>
                                            </button>
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
"""
content = content.replace(old_marker, cache_section + "\n" + old_marker, 1)

# 4. Add Edit Modal at the end of the body (before the scripts)
edit_modal = """
    <!-- Modal Editar Caché -->
    <div v-if="editCacheModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm" @click="editCacheModalOpen = false"></div>
        <div class="relative bg-dark-950 w-full max-w-3xl rounded-3xl border border-zinc-800 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
            <div class="p-6 border-b border-zinc-800 flex items-center justify-between bg-dark-900/50">
                <div>
                    <h2 class="text-xl font-bold text-white">Editar Registro JSON</h2>
                    <p class="text-sm text-zinc-400 mt-1">ID: <span class="font-bold text-violet-400">{{ selectedCacheItem.id }}</span></p>
                </div>
                <button @click="editCacheModalOpen = false" class="text-zinc-500 hover:text-white transition-colors">
                    <i data-lucide="x" class="w-6 h-6"></i>
                </button>
            </div>
            <div class="p-6 flex-1 overflow-auto">
                <p class="text-xs text-amber-400 mb-3">⚠️ Precaución: Edita el JSON con cuidado. Debe ser un formato válido.</p>
                <textarea v-model="selectedCacheItem.jsonString" class="w-full h-[400px] bg-dark-900 border border-zinc-700 rounded-2xl p-4 text-emerald-400 font-mono text-sm focus:outline-none focus:border-violet-500" spellcheck="false"></textarea>
            </div>
            <div class="p-6 border-t border-zinc-800 flex justify-end gap-3 bg-dark-900/50">
                <button @click="editCacheModalOpen = false" class="px-6 py-2.5 rounded-xl font-bold text-sm bg-dark-900 text-zinc-300 hover:text-white border border-zinc-700 transition-all">Cancelar</button>
                <button @click="saveCache" class="px-6 py-2.5 rounded-xl font-bold text-sm bg-violet-600 hover:bg-violet-500 text-white shadow-lg transition-all">Guardar JSON</button>
            </div>
        </div>
    </div>
"""
app_close_tag = "    </div>\n\n    <!-- Scripts de Lógica del Panel -->"
content = content.replace(app_close_tag, edit_modal + "\n" + app_close_tag, 1)


# 5. Add Vue Variables
old_vars = "                const snippetTab = ref('curl');"
new_vars = """                const cacheType = ref('dni');
                const cacheData = ref([]);
                const cacheLoading = ref(false);
                const editCacheModalOpen = ref(false);
                const selectedCacheItem = ref({});
                const snippetTab = ref('curl');"""
content = content.replace(old_vars, new_vars, 1)

# 6. Add Vue Methods
old_methods = "                const fetchProviders = async () => {"
new_methods = """                const fetchCache = async () => {
                    cacheLoading.value = true;
                    try {
                        const res = await fetch(`${API_BASE}/api/admin/cache/${cacheType.value}`, { headers: getAuthHeaders() });
                        const data = await res.json();
                        cacheData.value = data || [];
                    } catch(e) { console.error(e); }
                    finally { cacheLoading.value = false; nextTick(() => lucide.createIcons()); }
                };

                const deleteCache = async (id) => {
                    if(!confirm(`¿Estás seguro de eliminar el registro ${id} de la caché?`)) return;
                    try {
                        const res = await fetch(`${API_BASE}/api/admin/cache/${cacheType.value}/${id}`, {
                            method: 'DELETE',
                            headers: getAuthHeaders()
                        });
                        if(!res.ok) throw new Error('Error al eliminar');
                        fetchCache();
                    } catch(e) { alert(e.message); }
                };

                const openEditCache = (item) => {
                    selectedCacheItem.value = {
                        id: item.id,
                        jsonString: JSON.stringify(item.data, null, 4)
                    };
                    editCacheModalOpen.value = true;
                    nextTick(() => lucide.createIcons());
                };

                const saveCache = async () => {
                    try {
                        // Validar que sea JSON
                        JSON.parse(selectedCacheItem.value.jsonString);
                    } catch(e) {
                        alert("El JSON no es válido. Revisa la sintaxis.");
                        return;
                    }
                    
                    try {
                        const res = await fetch(`${API_BASE}/api/admin/cache/${cacheType.value}/${selectedCacheItem.value.id}`, {
                            method: 'PUT',
                            headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
                            body: selectedCacheItem.value.jsonString
                        });
                        if(!res.ok) throw new Error('Error al guardar caché');
                        editCacheModalOpen.value = false;
                        fetchCache();
                    } catch(e) { alert(e.message); }
                };

""" + old_methods
content = content.replace(old_methods, new_methods, 1)


# 7. Expose methods to template
old_return = "                    executePlayground,"
new_return = """                    cacheType,
                    cacheData,
                    cacheLoading,
                    editCacheModalOpen,
                    selectedCacheItem,
                    fetchCache,
                    deleteCache,
                    openEditCache,
                    saveCache,
                    executePlayground,"""
content = content.replace(old_return, new_return, 1)


with open("frontend/index.html", "w", encoding="utf-8") as f:
    f.write(content)

print("Cache UI patched OK")
