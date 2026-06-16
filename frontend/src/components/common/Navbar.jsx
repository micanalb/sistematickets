import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import './Navbar.css'

export default function Navbar() {
  const { estaAutenticado, usuario, esAdmin, cerrarSesion } = useAuth()
  const navigate = useNavigate()

  const handleCerrarSesion = () => {
    cerrarSesion()
    navigate('/')
  }

  return (
    <nav className="navbar">
      <div className="contenedor navbar-contenido">
        <Link to="/" className="navbar-logo">
          <span className="navbar-logo-tickets">Tickets</span>
          <span className="navbar-logo-ya">Ya</span>
        </Link>

        <div className="navbar-acciones">
          {estaAutenticado ? (
            <>
              <span className="navbar-saludo">
                Hola, <strong>{usuario?.nombre}</strong>
                {esAdmin && <span className="navbar-badge-admin">Admin</span>}
              </span>

              {esAdmin ? (
                <Link to="/admin" className="navbar-link">Panel Admin</Link>
              ) : (
                <Link to="/mis-entradas" className="navbar-link">Mis Entradas</Link>
              )}

              <button className="btn btn-secundario btn-sm" onClick={handleCerrarSesion}>
                Salir
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="navbar-link">Iniciar Sesión</Link>
              <Link to="/registro" className="btn btn-primario btn-sm">Registrarse</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}
