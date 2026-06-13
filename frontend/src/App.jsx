import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'

import PaginaInicio from './pages/PaginaInicio'
import PaginaDetalle from './pages/PaginaDetalle'
import PaginaLogin from './pages/PaginaLogin'
import PaginaRegistro from './pages/PaginaRegistro'
import PaginaMisEntradas from './pages/PaginaMisEntradas'
import PaginaAdmin from './pages/PaginaAdmin'
import Navbar from './components/common/Navbar'

function RutaProtegida({ children }) {
  const { estaAutenticado, cargando } = useAuth()
  if (cargando) return (
    <div className="cargando-contenedor" style={{ minHeight: '100vh' }}>
      <div className="spinner" /><span>Cargando...</span>
    </div>
  )
  if (!estaAutenticado) return <Navigate to="/login" replace />
  return children
}

function RutaAdmin({ children }) {
  const { estaAutenticado, esAdmin, cargando } = useAuth()
  if (cargando) return null
  if (!estaAutenticado) return <Navigate to="/login" replace />
  if (!esAdmin) return <Navigate to="/" replace />
  return children
}

function RutaPublicaSolo({ children }) {
  const { estaAutenticado, cargando } = useAuth()
  if (cargando) return null
  if (estaAutenticado) return <Navigate to="/" replace />
  return children
}

function AppContenido() {
  return (
    <BrowserRouter>
      <Navbar />
      <main>
        <Routes>
          <Route path="/" element={<PaginaInicio />} />
          <Route path="/eventos/:id" element={<PaginaDetalle />} />
          <Route path="/login" element={<RutaPublicaSolo><PaginaLogin /></RutaPublicaSolo>} />
          <Route path="/registro" element={<RutaPublicaSolo><PaginaRegistro /></RutaPublicaSolo>} />
          <Route path="/mis-entradas" element={<RutaProtegida><PaginaMisEntradas /></RutaProtegida>} />
          <Route path="/admin" element={<RutaAdmin><PaginaAdmin /></RutaAdmin>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <AppContenido />
    </AuthProvider>
  )
}
