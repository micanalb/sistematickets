import axios from 'axios'

// URL base de la API del backend
const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

/**
 * Instancia de axios configurada para la API de TicketsYa.
 * Los interceptors inyectan automáticamente el token JWT en cada request.
 */
const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
})

// ── Interceptor de REQUEST: agrega el token JWT automáticamente ─────────────
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('ticketsya_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// ── Interceptor de RESPONSE: manejo centralizado de errores ─────────────────
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Si el token expiró o es inválido, limpiar la sesión
    if (error.response?.status === 401) {
      localStorage.removeItem('ticketsya_token')
      localStorage.removeItem('ticketsya_usuario')
      // Redirigir al login si no estamos ya ahí
      if (window.location.pathname !== '/login' && window.location.pathname !== '/registro') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// ── Servicios de Autenticación ───────────────────────────────────────────────
export const authAPI = {
  /**
   * Registrar un nuevo usuario
   * @param {Object} datos - { nombre, apellido, email, password, telefono }
   */
  registrar: (datos) => api.post('/auth/registro', datos),

  /**
   * Iniciar sesión
   * @param {Object} datos - { email, password }
   */
  login: (datos) => api.post('/auth/login', datos),
}

// ── Servicios de Eventos ─────────────────────────────────────────────────────
export const eventosAPI = {
  /**
   * Listar eventos con filtros opcionales
   * @param {Object} filtros - { busqueda, categoria, ciudad, solo_disponibles }
   */
  listar: (filtros = {}) => api.get('/eventos', { params: filtros }),

  /**
   * Obtener detalle de un evento
   * @param {number} id - ID del evento
   */
  obtener: (id) => api.get(`/eventos/${id}`),

  // ── Admin ──────────────────────────────────────────────────────────────────
  crear: (datos) => api.post('/admin/eventos', datos),
  actualizar: (id, datos) => api.put(`/admin/eventos/${id}`, datos),
  eliminar: (id) => api.delete(`/admin/eventos/${id}`),
  reporte: (id) => api.get(`/admin/eventos/${id}/reporte`),
}

// ── Servicios de Entradas ────────────────────────────────────────────────────
export const entradasAPI = {
  /**
   * Comprar una entrada para un evento
   * @param {number} eventoID - ID del evento
   */
  comprar: (eventoID) => api.post('/entradas/comprar', { evento_id: eventoID }),

  /**
   * Obtener mis entradas (historial del usuario autenticado)
   */
  misEntradas: () => api.get('/entradas/mis-entradas'),

  /**
   * Cancelar una entrada propia
   * @param {number} entradaID - ID de la entrada a cancelar
   */
  cancelar: (entradaID) => api.put(`/entradas/${entradaID}/cancelar`),

  /**
   * Transferir una entrada a otro usuario
   * @param {number} entradaID - ID de la entrada
   * @param {string} emailDestinatario - Email del usuario destinatario
   */
  transferir: (entradaID, emailDestinatario) =>
    api.put(`/entradas/${entradaID}/transferir`, { email_destinatario: emailDestinatario }),
}

export default api
