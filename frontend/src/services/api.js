import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 10000,
})

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('ticketsya_token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('ticketsya_token')
      localStorage.removeItem('ticketsya_usuario')
      if (window.location.pathname !== '/login' && window.location.pathname !== '/registro') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export const authAPI = {
  registrar: (datos) => api.post('/auth/registro', datos),
  login: (datos) => api.post('/auth/login', datos),
}

export const eventosAPI = {
  // Públicos
  listar: (filtros = {}) => api.get('/eventos', { params: filtros }),
  obtener: (id) => api.get(`/eventos/${id}`),
  // Admin
  crear: (datos) => api.post('/admin/eventos', datos),
  actualizar: (id, datos) => api.put(`/admin/eventos/${id}`, datos),
  eliminar: (id) => api.delete(`/admin/eventos/${id}`),
  reporte: (id) => api.get(`/admin/eventos/${id}/reporte`),
}

export const entradasAPI = {
  comprar: (eventoID) => api.post('/entradas/comprar', { evento_id: eventoID }),
  misEntradas: () => api.get('/entradas/mis-entradas'),
  cancelar: (entradaID) => api.put(`/entradas/${entradaID}/cancelar`),
  transferir: (entradaID, emailDestinatario) =>
    api.put(`/entradas/${entradaID}/transferir`, { email_destinatario: emailDestinatario }),
}

export default api
