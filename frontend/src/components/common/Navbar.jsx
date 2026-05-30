import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import './Navbar.css'

/**
 * Navbar — Barra de navegación persistente.
 * Muestra opciones distintas según el estado de autenticación del usuario.
 */
export default function Navbar() {
  const { estaAutenticado, usuario, cerrarSesion } = useAuth()
  const navigate = useNavigate()

  const handleCerrarSesion = () => {
    cerrarSesion()
    navigate('/')
  }

  return (
    <nav className="navbar">
      <div className="contenedor navbar-contenido">
        {/* Logo */}
        <Link to="/" className="navbar-logo">
          <span className="navbar-logo-tickets">Tickets</span>
          <span className="navbar-logo-ya">Ya</span>
        </Link>

        {/* Navegación */}
        <div className="navbar-acciones">
          {estaAutenticado ? (
            <>
              {/* Saludo al usuario */}
              <span className="navbar-saludo">
                Hola, <strong>{usuario?.nombre}</strong>
              </span>

              {/* Link a mis entradas */}
              <Link to="/mis-entradas" className="navbar-link">
                Mis Entradas
              </Link>

              {/* Botón cerrar sesión */}
              <button
                className="btn btn-secundario btn-sm"
                onClick={handleCerrarSesion}
              >
                Salir
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="navbar-link">
                Iniciar Sesión
              </Link>
              <Link to="/registro" className="btn btn-primario btn-sm">
                Registrarse
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}
