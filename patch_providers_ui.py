with open("frontend/index.html", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Add sidebar button after pricing tab
old_pricing_btn = '''                    <button @click="currentTab = 'pricing'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'pricing' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="credit-card" class="w-4 h-4"></i>
                        Planes y Precios
                    </button>'''

new_pricing_btn = old_pricing_btn + '''
                    <button @click="currentTab = 'providers'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'providers' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="settings-2" class="w-4 h-4"></i>
                        Configurar APIs
                    </button>'''

content = content.replace(old_pricing_btn, new_pricing_btn)

# 2. Update header title
old_title = "currentTab === 'pricing' ? 'Planes de Precios' }}"
new_title = "currentTab === 'pricing' ? 'Planes de Precios' : 'Configurar Proveedores Externos' }}"
content = content.replace(old_title, new_title)

old_desc = "currentTab === 'docs' ? 'Guías de integración, endpoints y ejemplos de código.' : 'Revisa nuestras tarifas y límites mensuales en Soles (PEN).' }}"
new_desc = "currentTab === 'docs' ? 'Guías de integración, endpoints y ejemplos de código.' : currentTab === 'pricing' ? 'Revisa nuestras tarifas y límites mensuales en Soles (PEN).' : 'Administra las claves y el orden de prioridad de las APIs externas.' }}"
content = content.replace(old_desc, new_desc)

# 3. Add providers section before closing main-content div
old_marker = "                    <!-- ==================== SECCIÓN 5: PLANES Y PRECIOS ==================== -->"
providers_section = """
                    <!-- ==================== SECCIÓN 6: CONFIGURAR APIS ==================== -->
                    <div v-if="currentTab === 'providers'" class="space-y-6">
                        <div v-if="providersLoading" class="text-center py-20 text-zinc-500 text-sm">Cargando configuración...</div>
                        <div v-else>
                            <div class="mb-4 p-4 bg-amber-500/10 border border-amber-500/30 rounded-2xl flex items-start gap-3">
                                <i data-lucide="alert-triangle" class="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5"></i>
                                <div>
                                    <p class="text-sm font-bold text-amber-400">¡Importante!</p>
                                    <p class="text-xs text-zinc-400 mt-1">Los cambios se aplican inmediatamente sin necesidad de reiniciar el servidor. El proveedor con menor número de prioridad se ejecuta primero (1 = primero).</p>
                                </div>
                            </div>
                            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                <div v-for="(p, idx) in providers" :key="p.provider_name" class="glass-panel p-6 rounded-3xl border transition-all" :class="p.is_active ? 'border-violet-500/30' : 'border-zinc-800'">
                                    <div class="flex items-center justify-between mb-4">
                                        <div class="flex items-center gap-3">
                                            <div class="w-10 h-10 rounded-2xl flex items-center justify-center text-lg font-extrabold" :class="p.is_active ? 'bg-violet-500/20 text-violet-400' : 'bg-zinc-800 text-zinc-500'">
                                                {{ idx + 1 }}
                                            </div>
                                            <div>
                                                <h3 class="text-base font-bold text-white uppercase">{{ p.provider_name }}</h3>
                                                <span class="text-xs" :class="p.is_active ? 'text-emerald-400' : 'text-red-400'">{{ p.is_active ? '● Activo' : '● Inactivo' }}</span>
                                            </div>
                                        </div>
                                        <label class="relative inline-flex items-center cursor-pointer">
                                            <input type="checkbox" v-model="p.is_active" class="sr-only peer">
                                            <div class="w-11 h-6 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:bg-violet-600 transition-all after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"></div>
                                        </label>
                                    </div>
                                    <div class="space-y-3">
                                        <div>
                                            <label class="block text-xs font-bold uppercase text-zinc-400 mb-1.5">API Key</label>
                                            <input v-model="p.api_key" type="text" class="block w-full px-4 py-3 bg-dark-900 border border-zinc-800 rounded-2xl text-white text-sm font-mono focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="Pega la API Key aquí...">
                                        </div>
                                        <div>
                                            <label class="block text-xs font-bold uppercase text-zinc-400 mb-1.5">Prioridad de ejecución</label>
                                            <select v-model.number="p.priority" class="block w-full px-4 py-3 bg-dark-900 border border-zinc-800 rounded-2xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-violet-500">
                                                <option :value="1">1 — Primero (Principal)</option>
                                                <option :value="2">2 — Segundo (Fallback)</option>
                                                <option :value="3">3 — Tercero (Último recurso)</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div class="flex justify-end mt-6">
                                <button @click="saveProviders" :disabled="providersSaving" class="flex items-center gap-2 px-8 py-3.5 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 rounded-2xl text-white font-bold shadow-lg transition-all disabled:opacity-50">
                                    <svg v-if="providersSaving" class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                                    <i v-else data-lucide="save" class="w-4 h-4"></i>
                                    {{ providersSaving ? 'Guardando...' : 'Guardar Cambios' }}
                                </button>
                            </div>
                            <p v-if="providersSaved" class="text-center text-sm text-emerald-400 mt-2">✓ Configuración guardada y aplicada en tiempo real.</p>
                        </div>
                    </div>
"""
content = content.replace(old_marker, providers_section + "\n" + old_marker)

# 4. Add Vue reactive vars
old_vars = "                const snippetTab = ref('curl');"
new_vars = """                const providers = ref([]);
                const providersLoading = ref(false);
                const providersSaving = ref(false);
                const providersSaved = ref(false);
                const snippetTab = ref('curl');"""
content = content.replace(old_vars, new_vars)

# 5. Add methods
old_methods_marker = "                const executePlayground = async () => {"
new_methods = """                const fetchProviders = async () => {
                    providersLoading.value = true;
                    try {
                        const res = await fetch(`${API_BASE}/api/admin/providers`, { headers: getAuthHeaders() });
                        const data = await res.json();
                        providers.value = data;
                    } catch(e) { console.error(e); }
                    finally { providersLoading.value = false; nextTick(() => lucide.createIcons()); }
                };

                const saveProviders = async () => {
                    providersSaving.value = true;
                    providersSaved.value = false;
                    try {
                        const res = await fetch(`${API_BASE}/api/admin/providers`, {
                            method: 'PUT',
                            headers: { ...getAuthHeaders(), 'Content-Type': 'application/json' },
                            body: JSON.stringify(providers.value)
                        });
                        if (!res.ok) throw new Error('Error al guardar');
                        providersSaved.value = true;
                        setTimeout(() => providersSaved.value = false, 3000);
                    } catch(e) { alert('Error al guardar proveedores: ' + e.message); }
                    finally { providersSaving.value = false; }
                };

""" + old_methods_marker
content = content.replace(old_methods_marker, new_methods)

# 6. Update fetchData to also call fetchProviders
old_fetch = "                const fetchData = async () => {"
new_fetch = """                const fetchData = async () => {
                    fetchProviders();
"""
content = content.replace(old_fetch + "\n", new_fetch)
# Fix if already done
if "fetchProviders();" not in content:
    content = content.replace(old_fetch, new_fetch)

# 7. Add to return object
old_return = "                    executePlayground,"
new_return = """                    providers,
                    providersLoading,
                    providersSaving,
                    providersSaved,
                    fetchProviders,
                    saveProviders,
                    executePlayground,"""
content = content.replace(old_return, new_return)

with open("frontend/index.html", "w", encoding="utf-8") as f:
    f.write(content)

print("Providers UI patched successfully")
