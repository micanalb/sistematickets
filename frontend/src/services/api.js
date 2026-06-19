import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'
//import.meta.env es la forma que tiene Vite de exponer variables de entorno al código del navegador. 
//Solo las que arrancan con VITE_ quedan expuestas, por seguridad — el resto del .env no llega al bundle. 
// Si no está definida, cae a /api/v1 

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 10000,
})
//Crea una instancia de axios configurada una sola vez, en vez de repetir baseURL, headers y timeout en cada llamada. 
//timeout: 10000 corta la petición si tarda más de 10 segundos.

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('ticketsya_token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error)
)
//Antes de cada request, busca el JWT guardado en localStorage y si existe lo pone en el header Authorization. 
//Esto es lo que hace que no tengas que acordarte de mandar el token a mano en cada llamada a eventosAPI, entradasAPI, etc. — queda centralizado en un solo lugar.


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
//Si cualquier respuesta del backend viene con 401 (no autorizado — típicamente porque el JWT venció o es inválido), automáticamente:
//1. Borra el token y los datos de usuario guardados.
//2. Te manda a /login, salvo que ya estés en /login o /registro (esto evita un loop infinito de redirects).

export const authAPI = {
  registrar: (datos) => api.post('/auth/registro', datos),
  login: (datos) => api.post('/auth/login', datos),
}
//En vez de escribir api.post('/auth/login', datos) repetido por toda la app, en los componentes vas a llamar authAPI.login(datos). 
//Más legible y si cambia la URL del endpoint, la cambiás en un solo lugar.


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
//Notá la separación implícita entre rutas públicas (/eventos) y rutas de admin (/admin/eventos) — coincide con lo que mencionaste sobre roles (administrador en la base). 
//La protección real de "quién puede hacer esto" no está acá (es solo el cliente armando la URL), está en el backend con el middleware de JWT/rol — este archivo asume que el backend va a rechazar con 401/403 si no corresponde.

export const entradasAPI = {
  comprar: (eventoID) => api.post('/entradas/comprar', { evento_id: eventoID }),
  misEntradas: () => api.get('/entradas/mis-entradas'),
  cancelar: (entradaID) => api.put(`/entradas/${entradaID}/cancelar`),
  transferir: (entradaID, emailDestinatario) =>
    api.put(`/entradas/${entradaID}/transferir`, { email_destinatario: emailDestinatario }),
}
//Mismo patrón para entradas: comprar, listar las propias, cancelar, transferir a otro usuario por email. 
//cancelar y transferir son PUT (modifican un recurso existente), comprar es POST (crea uno nuevo) — convención REST correcta.

export default api
