import re

with open("frontend/index.html", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Sidebar Buttons
sidebar_btn = """                    <button @click="currentTab = 'tenants'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'tenants' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="users" class="w-4 h-4"></i>
                        Inquilinos (Tenants)
                    </button>"""

new_sidebar_btns = sidebar_btn + """
                    <button @click="currentTab = 'playground'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'playground' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="play-circle" class="w-4 h-4"></i>
                        API Playground
                    </button>
                    <button @click="currentTab = 'docs'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'docs' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="book-open" class="w-4 h-4"></i>
                        Documentación
                    </button>
                    <button @click="currentTab = 'pricing'" 
                        class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-semibold transition-all duration-200"
                        :class="currentTab === 'pricing' ? 'bg-violet-600 text-white glow-violet' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'">
                        <i data-lucide="credit-card" class="w-4 h-4"></i>
                        Planes y Precios
                    </button>"""

content = content.replace(sidebar_btn, new_sidebar_btns)

# 2. Update Headers
header_title = """{{ currentTab === 'dashboard' ? 'Métricas de Infraestructura' : 'Gestión de Inquilinos (SaaS)' }}"""
new_header_title = """{{ currentTab === 'dashboard' ? 'Métricas de Infraestructura' : currentTab === 'tenants' ? 'Gestión de Inquilinos (SaaS)' : currentTab === 'playground' ? 'API Playground' : currentTab === 'docs' ? 'Documentación de la API' : 'Planes de Precios' }}"""

header_desc = """{{ currentTab === 'dashboard' ? 'Monitoreo de optimización, latencias y ahorro de costos.' : 'Administra API Keys, estados y accesos para cada cliente.' }}"""
new_header_desc = """{{ currentTab === 'dashboard' ? 'Monitoreo de optimización, latencias y ahorro de costos.' : currentTab === 'tenants' ? 'Administra API Keys, estados y accesos para cada cliente.' : currentTab === 'playground' ? 'Prueba la API en tiempo real con credenciales de prueba.' : currentTab === 'docs' ? 'Guías de integración, endpoints y ejemplos de código.' : 'Revisa nuestras tarifas y límites mensuales en Soles (PEN).' }}"""

content = content.replace(header_title, new_header_title)
content = content.replace(header_desc, new_header_desc)

# 3. Insert Views
views_marker = """                    <!-- ==================== SECCIÓN 2: INQUILINOS (TENANTS) ==================== -->"""
views_marker_end = """                        </div>\n\n                    </div>\n\n                </div>"""

new_views = """
                    <!-- ==================== SECCIÓN 3: PLAYGROUND ==================== -->
                    <div v-if="currentTab === 'playground'" class="space-y-6">
                        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                            <!-- Formulario -->
                            <div class="glass-panel p-6 rounded-3xl relative overflow-hidden">
                                <div class="flex items-center justify-between mb-6">
                                    <div>
                                        <h3 class="text-base font-bold text-white">Probar Consultas</h3>
                                        <p class="text-xs text-zinc-400">Ejecuta peticiones reales desde el navegador.</p>
                                    </div>
                                    <div class="p-2 bg-violet-500/10 rounded-xl text-violet-400">
                                        <i data-lucide="play" class="w-4 h-4"></i>
                                    </div>
                                </div>
                                <form @submit.prevent="executePlayground" class="space-y-4">
                                    <div>
                                        <label class="block text-xs font-semibold uppercase text-zinc-400 mb-1.5">Tipo de Documento</label>
                                        <select v-model="playgroundForm.type" class="block w-full px-4 py-3 bg-dark-900 border border-zinc-800 rounded-2xl text-white focus:outline-none focus:ring-2 focus:ring-violet-500 text-sm">
                                            <option value="dni">DNI (Persona)</option>
                                            <option value="ruc">RUC (Empresa)</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label class="block text-xs font-semibold uppercase text-zinc-400 mb-1.5">Número de Documento</label>
                                        <input v-model="playgroundForm.document" type="text" class="block w-full px-4 py-3 bg-dark-900 border border-zinc-800 rounded-2xl text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-violet-500 text-sm" placeholder="Ej. 12345678" required>
                                    </div>
                                    <div>
                                        <label class="block text-xs font-semibold uppercase text-zinc-400 mb-1.5">API Key (x-api-key)</label>
                                        <input v-model="playgroundForm.apiKey" type="text" class="block w-full px-4 py-3 bg-dark-900 border border-zinc-800 rounded-2xl text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-violet-500 text-sm" placeholder="tc_..." required>
                                    </div>
                                    <button type="submit" :disabled="playgroundLoading" class="w-full flex justify-center items-center gap-2 py-3.5 border border-transparent text-sm font-semibold rounded-2xl text-white bg-violet-600 hover:bg-violet-500 transition-all disabled:opacity-50">
                                        <i data-lucide="send" class="w-4 h-4" v-if="!playgroundLoading"></i>
                                        <svg v-else class="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                                        {{ playgroundLoading ? 'Ejecutando...' : 'Enviar Petición' }}
                                    </button>
                                </form>
                            </div>
                            <!-- Resultado -->
                            <div class="glass-panel p-6 rounded-3xl flex flex-col h-[500px]">
                                <div class="flex items-center justify-between mb-4">
                                    <h3 class="text-base font-bold text-white">Respuesta JSON</h3>
                                    <span v-if="playgroundStatus" class="text-xs font-bold px-2 py-1 rounded-lg" :class="playgroundStatus >= 400 ? 'bg-red-500/20 text-red-400' : 'bg-emerald-500/20 text-emerald-400'">HTTP {{ playgroundStatus }}</span>
                                </div>
                                <div class="flex-1 bg-dark-950 border border-zinc-800/80 rounded-2xl p-4 overflow-auto custom-scrollbar">
                                    <pre class="text-xs font-mono text-zinc-300">{{ playgroundResult || '// La respuesta aparecerá aquí...' }}</pre>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- ==================== SECCIÓN 4: DOCUMENTACIÓN ==================== -->
                    <div v-if="currentTab === 'docs'" class="space-y-6">
                        <div class="glass-panel p-8 rounded-3xl">
                            <h2 class="text-xl font-bold text-white mb-6">Guía de Integración Rápida</h2>
                            
                            <div class="space-y-8">
                                <!-- Base URL -->
                                <div>
                                    <h3 class="text-sm font-bold text-violet-400 uppercase tracking-widest mb-3">1. URL Base y Autenticación</h3>
                                    <p class="text-sm text-zinc-400 mb-3">Todas las peticiones deben dirigirse a `https://apiconsulta.sehuacho.com/api` e incluir el header `x-api-key`.</p>
                                    <div class="bg-dark-950 border border-zinc-800 rounded-xl p-4 text-sm font-mono text-zinc-300">
                                        x-api-key: tu_api_key_aqui
                                    </div>
                                </div>

                                <!-- Endpoints -->
                                <div>
                                    <h3 class="text-sm font-bold text-violet-400 uppercase tracking-widest mb-3">2. Endpoints Disponibles</h3>
                                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div class="p-4 bg-dark-900 border border-zinc-800 rounded-2xl">
                                            <div class="flex items-center gap-2 mb-2">
                                                <span class="px-2 py-1 bg-emerald-500/20 text-emerald-400 text-xs font-bold rounded">GET</span>
                                                <code class="text-white text-sm">/api/dni/{numero}</code>
                                            </div>
                                            <p class="text-xs text-zinc-400">Consulta datos de personas naturales mediante su DNI (8 dígitos).</p>
                                        </div>
                                        <div class="p-4 bg-dark-900 border border-zinc-800 rounded-2xl">
                                            <div class="flex items-center gap-2 mb-2">
                                                <span class="px-2 py-1 bg-emerald-500/20 text-emerald-400 text-xs font-bold rounded">GET</span>
                                                <code class="text-white text-sm">/api/ruc/{numero}</code>
                                            </div>
                                            <p class="text-xs text-zinc-400">Consulta información de empresas o contribuyentes (11 dígitos).</p>
                                        </div>
                                    </div>
                                </div>

                                <!-- Snippets -->
                                <div>
                                    <h3 class="text-sm font-bold text-violet-400 uppercase tracking-widest mb-3">3. Ejemplos de Código</h3>
                                    <div class="flex gap-2 mb-4 overflow-x-auto">
                                        <button @click="snippetTab = 'curl'" class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all" :class="snippetTab === 'curl' ? 'bg-violet-600 text-white' : 'bg-dark-900 text-zinc-400 hover:bg-zinc-800'">cURL</button>
                                        <button @click="snippetTab = 'python'" class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all" :class="snippetTab === 'python' ? 'bg-violet-600 text-white' : 'bg-dark-900 text-zinc-400 hover:bg-zinc-800'">Python</button>
                                        <button @click="snippetTab = 'go'" class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all" :class="snippetTab === 'go' ? 'bg-violet-600 text-white' : 'bg-dark-900 text-zinc-400 hover:bg-zinc-800'">Go</button>
                                    </div>
                                    <div class="bg-dark-950 border border-zinc-800 rounded-2xl p-4 overflow-x-auto custom-scrollbar">
<pre v-if="snippetTab === 'curl'" class="text-sm font-mono text-zinc-300">curl -X GET "https://apiconsulta.sehuacho.com/api/dni/12345678" \\
     -H "x-api-key: tu_api_key_aqui"</pre>
<pre v-if="snippetTab === 'python'" class="text-sm font-mono text-zinc-300">import requests

url = "https://apiconsulta.sehuacho.com/api/dni/12345678"
headers = {"x-api-key": "tu_api_key_aqui"}

response = requests.get(url, headers=headers)
print(response.json())</pre>
<pre v-if="snippetTab === 'go'" class="text-sm font-mono text-zinc-300">package main

import (
"fmt"
"net/http"
"io/ioutil"
)

func main() {
req, _ := http.NewRequest("GET", "https://apiconsulta.sehuacho.com/api/dni/12345678", nil)
req.Header.Add("x-api-key", "tu_api_key_aqui")

res, _ := http.DefaultClient.Do(req)
defer res.Body.Close()
body, _ := ioutil.ReadAll(res.Body)

fmt.Println(string(body))
}</pre>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- ==================== SECCIÓN 5: PLANES Y PRECIOS ==================== -->
                    <div v-if="currentTab === 'pricing'" class="space-y-6">
                        <div class="text-center mb-8">
                            <h2 class="text-2xl font-bold text-white">Tarifas Transparentes en Soles</h2>
                            <p class="text-sm text-zinc-400 mt-2">Planes escalables para proyectos de cualquier tamaño.</p>
                        </div>
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                            <!-- Plan Developer -->
                            <div class="glass-panel p-8 rounded-3xl border border-zinc-800 flex flex-col hover:border-violet-500/30 transition-all">
                                <h3 class="text-lg font-bold text-zinc-200">Developer</h3>
                                <p class="text-sm text-zinc-400 mt-1 mb-6">Ideal para pruebas y proyectos pequeños.</p>
                                <div class="text-4xl font-extrabold text-white mb-6">S/. 0<span class="text-base text-zinc-500 font-normal">/mes</span></div>
                                <ul class="space-y-3 mb-8 flex-1 text-sm text-zinc-300">
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>1,000 Peticiones/mes</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Caché local rápido</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Soporte de comunidad</li>
                                </ul>
                                <button class="w-full py-3 bg-dark-900 border border-zinc-700 hover:bg-zinc-800 rounded-xl font-bold text-zinc-200 transition-all">Empezar Gratis</button>
                            </div>
                            
                            <!-- Plan Business -->
                            <div class="glass-panel p-8 rounded-3xl border border-violet-500 glow-violet flex flex-col relative transform md:-translate-y-4">
                                <div class="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-gradient-to-r from-violet-600 to-indigo-600 text-white text-xs font-bold px-4 py-1 rounded-full uppercase tracking-widest shadow-lg">Más Popular</div>
                                <h3 class="text-lg font-bold text-violet-400 mt-2">Business Elite</h3>
                                <p class="text-sm text-zinc-400 mt-1 mb-6">Para empresas con alto volumen de tráfico.</p>
                                <div class="text-4xl font-extrabold text-white mb-6">S/. 150<span class="text-base text-zinc-500 font-normal">/mes</span></div>
                                <ul class="space-y-3 mb-8 flex-1 text-sm text-zinc-300">
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>500,000 Peticiones/mes</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Sin límites de concurrencia</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Caché en tiempo real (99.9% Uptime)</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Soporte prioritario 24/7</li>
                                </ul>
                                <button class="w-full py-3 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 rounded-xl font-bold text-white shadow-lg transition-all">Suscribirse Ahora</button>
                            </div>

                            <!-- Plan Enterprise -->
                            <div class="glass-panel p-8 rounded-3xl border border-zinc-800 flex flex-col hover:border-violet-500/30 transition-all">
                                <h3 class="text-lg font-bold text-zinc-200">Enterprise SaaS</h3>
                                <p class="text-sm text-zinc-400 mt-1 mb-6">Infraestructura dedicada y SLAs a medida.</p>
                                <div class="text-4xl font-extrabold text-white mb-6">S/. 850<span class="text-base text-zinc-500 font-normal">/mes</span></div>
                                <ul class="space-y-3 mb-8 flex-1 text-sm text-zinc-300">
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Peticiones Ilimitadas</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Instancia de Base de Datos propia</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>VPN IP Dedicada</li>
                                    <li class="flex items-center gap-2"><i data-lucide="check" class="w-4 h-4 text-emerald-500"></i>Ingeniero asignado</li>
                                </ul>
                                <button class="w-full py-3 bg-dark-900 border border-zinc-700 hover:bg-zinc-800 rounded-xl font-bold text-zinc-200 transition-all">Contactar Ventas</button>
                            </div>
                        </div>
                    </div>
"""

content = content.replace(views_marker_end, views_marker_end + new_views)

# 4. Update Vue Setup
vue_vars = """                const currentTab = ref('dashboard');"""
new_vue_vars = vue_vars + """
                const snippetTab = ref('curl');
                const playgroundForm = reactive({ type: 'dni', document: '', apiKey: 'tc_master_key_2026' });
                const playgroundResult = ref('');
                const playgroundStatus = ref(null);
                const playgroundLoading = ref(false);"""
                
content = content.replace(vue_vars, new_vue_vars)

vue_methods = """                const getHitRate = () => {"""
new_vue_methods = """                const executePlayground = async () => {
                    playgroundLoading.value = true;
                    playgroundResult.value = '';
                    playgroundStatus.value = null;
                    try {
                        const res = await fetch(`${API_BASE}/api/${playgroundForm.type}/${playgroundForm.document}`, {
                            method: 'GET',
                            headers: { 'x-api-key': playgroundForm.apiKey }
                        });
                        playgroundStatus.value = res.status;
                        const data = await res.json();
                        playgroundResult.value = JSON.stringify(data, null, 2);
                    } catch (err) {
                        playgroundResult.value = String(err);
                    } finally {
                        playgroundLoading.value = false;
                    }
                };
""" + "\n" + vue_methods

content = content.replace(vue_methods, new_vue_methods)

vue_return = """                    getHitRate,"""
new_vue_return = vue_return + """
                    snippetTab,
                    playgroundForm,
                    playgroundResult,
                    playgroundStatus,
                    playgroundLoading,
                    executePlayground,"""

content = content.replace(vue_return, new_vue_return)

with open("frontend/index.html", "w", encoding="utf-8") as f:
    f.write(content)

print("Patched successfully")
