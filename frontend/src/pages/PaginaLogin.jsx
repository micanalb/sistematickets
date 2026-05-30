import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authAPI } from '../services/api'
import { useAuth } from '../context/AuthContext'
import './PaginaAuth.css'

export default function PaginaLogin() {
  const navigate = useNavigate()
  const { iniciarSesion } = useAuth()

  const [form, setForm] = useState({ email: '', password: '' })
  const [error, setError] = useState('')
  const [cargando, setCargando] = useState(false)

  const handleChange = (e) => {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))
    setError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!form.email || !form.password) {
      setError('Completá todos los campos.')
      return
    }

    setCargando(true)
    setError('')

    try {
      const response = await authAPI.login(form)
      const { token, usuario } = response.data.datos
      iniciarSesion(token, usuario)
      navigate('/')
    } catch (err) {
      const msg = err.response?.data?.error || 'Credenciales inválidas.'
      setError(msg)
    } finally {
      setCargando(false)
    }
  }

  return (
    <div className="auth-pagina">
      <div className="auth-card">
        {/* Logo */}
        <div className="auth-logo">
          <span>Tickets</span><span className="texto-acento">Ya</span>
        </div>

        <h1 className="auth-titulo">Iniciar sesión</h1>
        <p className="auth-subtitulo">Ingresá a tu cuenta para gestionar tus entradas</p>

        {error && <div className="alerta alerta-error">{error}</div>}

        <form onSubmit={handleSubmit} noValidate>
          <div className="form-grupo">
            <label className="form-etiqueta" htmlFor="email">Email</label>
            <input
              id="email"
              name="email"
              type="email"
              className="form-input"
              placeholder="tu@email.com"
              value={form.email}
              onChange={handleChange}
              autoComplete="email"
              required
            />
          </div>

          <div className="form-grupo">
            <label className="form-etiqueta" htmlFor="password">Contraseña</label>
            <input
              id="password"
              name="password"
              type="password"
              className="form-input"
              placeholder="••••••••"
              value={form.password}
              onChange={handleChange}
              autoComplete="current-password"
              required
            />
          </div>

          <button
            type="submit"
            className="btn btn-primario"
            style={{ width: '100%', marginTop: 8 }}
            disabled={cargando}
          >
            {cargando
              ? <><div className="spinner" style={{ width: 18, height: 18, borderWidth: 2 }} /> Ingresando...</>
              : 'Iniciar sesión'}
          </button>
        </form>

        <p className="auth-footer">
          ¿No tenés cuenta?{' '}
          <Link to="/registro">Registrate gratis</Link>
        </p>
      </div>
    </div>
  )
}
