import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'

// Páginas
import PaginaInicio from './pages/PaginaInicio'
import PaginaDetalle from './pages/PaginaDetalle'
import PaginaLogin from './pages/PaginaLogin'
import PaginaRegistro from './pages/PaginaRegistro'
import PaginaMisEntradas from './pages/PaginaMisEntradas'

// Componentes comunes
import Navbar from './components/common/Navbar'

/**
 * RutaProtegida - Wrapper para rutas que requieren autenticación.
 * Redirige al login si el usuario no está autenticado.
 */
function RutaProtegida({ children }) {
  const { estaAutenticado, cargando } = useAuth()

  if (cargando) {
    return (
      <div className="cargando-contenedor" style={{ minHeight: '100vh' }}>
        <div className="spinner" />
        <span>Cargando...</span>
      </div>
    )
  }

  if (!estaAutenticado) {
    return <Navigate to="/login" replace />
  }

  return children
}

/**
 * RutaPublicaSolo - Wrapper para rutas que solo acceden usuarios NO autenticados.
 * Redirige al inicio si el usuario ya tiene sesión activa.
 */
function RutaPublicaSolo({ children }) {
  const { estaAutenticado, cargando } = useAuth()

  if (cargando) return null

  if (estaAutenticado) {
    return <Navigate to="/" replace />
  }

  return children
}

/**
 * AppContenido - Define las rutas de la aplicación.
 * Separado del componente App para poder usar el hook useAuth.
 */
function AppContenido() {
  return (
    <BrowserRouter>
      <Navbar />
      <main>
        <Routes>
          {/* ── Rutas públicas ────────────────────────────────── */}
          <Route path="/" element={<PaginaInicio />} />
          <Route path="/eventos/:id" element={<PaginaDetalle />} />

          {/* ── Solo para usuarios NO autenticados ───────────── */}
          <Route
            path="/login"
            element={
              <RutaPublicaSolo>
                <PaginaLogin />
              </RutaPublicaSolo>
            }
          />
          <Route
            path="/registro"
            element={
              <RutaPublicaSolo>
                <PaginaRegistro />
              </RutaPublicaSolo>
            }
          />

          {/* ── Rutas protegidas (requieren login) ───────────── */}
          <Route
            path="/mis-entradas"
            element={
              <RutaProtegida>
                <PaginaMisEntradas />
              </RutaProtegida>
            }
          />

          {/* Ruta comodín: redirige al inicio */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

/**
 * App - componente raíz que envuelve la app con los providers globales
 */
export default function App() {
  return (
    <AuthProvider>
      <AppContenido />
    </AuthProvider>
  )
}
