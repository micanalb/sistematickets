import { createContext, useContext, useState, useEffect, useCallback } from 'react'

// Contexto de autenticación — disponible en toda la app
const AuthContext = createContext(null)

/**
 * AuthProvider envuelve la aplicación y expone el estado de autenticación.
 * Persiste el token en localStorage para mantener la sesión entre recargas.
 */
export function AuthProvider({ children }) {
  const [usuario, setUsuario] = useState(null)
  const [token, setToken] = useState(null)
  const [cargando, setCargando] = useState(true) // Mientras se hidrata desde localStorage

  // Al montar, intentar recuperar la sesión guardada
  useEffect(() => {
    const tokenGuardado = localStorage.getItem('ticketsya_token')
    const usuarioGuardado = localStorage.getItem('ticketsya_usuario')

    if (tokenGuardado && usuarioGuardado) {
      try {
        const usuarioParsed = JSON.parse(usuarioGuardado)
        setToken(tokenGuardado)
        setUsuario(usuarioParsed)
      } catch {
        // Si el JSON está corrupto, limpiar
        localStorage.removeItem('ticketsya_token')
        localStorage.removeItem('ticketsya_usuario')
      }
    }

    setCargando(false)
  }, [])

  /**
   * Iniciar sesión: guarda token y datos del usuario
   */
  const iniciarSesion = useCallback((tokenNuevo, datosUsuario) => {
    localStorage.setItem('ticketsya_token', tokenNuevo)
    localStorage.setItem('ticketsya_usuario', JSON.stringify(datosUsuario))
    setToken(tokenNuevo)
    setUsuario(datosUsuario)
  }, [])

  /**
   * Cerrar sesión: limpia todo el estado y localStorage
   */
  const cerrarSesion = useCallback(() => {
    localStorage.removeItem('ticketsya_token')
    localStorage.removeItem('ticketsya_usuario')
    setToken(null)
    setUsuario(null)
  }, [])

  const estaAutenticado = !!token && !!usuario
  const esAdmin = usuario?.rol === 'administrador'
  const esCliente = usuario?.rol === 'cliente'

  const valor = {
    usuario,
    token,
    cargando,
    estaAutenticado,
    esAdmin,
    esCliente,
    iniciarSesion,
    cerrarSesion,
  }

  return (
    <AuthContext.Provider value={valor}>
      {children}
    </AuthContext.Provider>
  )
}

/**
 * Hook personalizado para usar el contexto de auth en cualquier componente
 */
export function useAuth() {
  const contexto = useContext(AuthContext)
  if (!contexto) {
    throw new Error('useAuth debe usarse dentro de un AuthProvider')
  }
  return contexto
}
