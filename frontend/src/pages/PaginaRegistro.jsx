import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authAPI } from '../services/api'
import { useAuth } from '../context/AuthContext'
import './PaginaAuth.css'

export default function PaginaRegistro() {
  const navigate = useNavigate()
  const { iniciarSesion } = useAuth()

  const [form, setForm] = useState({
    nombre: '', apellido: '', email: '', password: '', telefono: ''
  })
  const [errores, setErrores] = useState({})
  const [errorGlobal, setErrorGlobal] = useState('')
  const [cargando, setCargando] = useState(false)

  const handleChange = (e) => {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))
    setErrores(prev => ({ ...prev, [e.target.name]: '' }))
    setErrorGlobal('')
  }

  const validar = () => {
    const nuevos = {}
    if (!form.nombre.trim()) nuevos.nombre = 'El nombre es requerido.'
    if (!form.apellido.trim()) nuevos.apellido = 'El apellido es requerido.'
    if (!form.email.trim()) nuevos.email = 'El email es requerido.'
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) nuevos.email = 'Email inválido.'
    if (!form.password) nuevos.password = 'La contraseña es requerida.'
    else if (form.password.length < 6) nuevos.password = 'Mínimo 6 caracteres.'
    return nuevos
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    const erroresValidacion = validar()
    if (Object.keys(erroresValidacion).length > 0) {
      setErrores(erroresValidacion)
      return
    }

    setCargando(true)
    setErrorGlobal('')

    try {
      const response = await authAPI.registrar(form)
      const { token, usuario } = response.data.datos
      iniciarSesion(token, usuario)
      navigate('/')
    } catch (err) {
      const msg = err.response?.data?.error || 'Error al crear la cuenta.'
      setErrorGlobal(msg)
    } finally {
      setCargando(false)
    }
  }

  return (
    <div className="auth-pagina">
      <div className="auth-card auth-card-lg">
        <div className="auth-logo">
          <span>Tickets</span><span className="texto-acento">Ya</span>
        </div>

        <h1 className="auth-titulo">Crear cuenta</h1>
        <p className="auth-subtitulo">Sumate y comprá entradas para tus eventos favoritos</p>

        {errorGlobal && <div className="alerta alerta-error">{errorGlobal}</div>}

        <form onSubmit={handleSubmit} noValidate>
          <div className="auth-fila-doble">
            <div className="form-grupo">
              <label className="form-etiqueta" htmlFor="nombre">Nombre *</label>
              <input
                id="nombre" name="nombre" type="text"
                className={`form-input ${errores.nombre ? 'form-input-error' : ''}`}
                placeholder="Juan"
                value={form.nombre} onChange={handleChange}
              />
              {errores.nombre && <span className="form-error-msg">{errores.nombre}</span>}
            </div>

            <div className="form-grupo">
              <label className="form-etiqueta" htmlFor="apellido">Apellido *</label>
              <input
                id="apellido" name="apellido" type="text"
                className={`form-input ${errores.apellido ? 'form-input-error' : ''}`}
                placeholder="Pérez"
                value={form.apellido} onChange={handleChange}
              />
              {errores.apellido && <span className="form-error-msg">{errores.apellido}</span>}
            </div>
          </div>

          <div className="form-grupo">
            <label className="form-etiqueta" htmlFor="email">Email *</label>
            <input
              id="email" name="email" type="email"
              className={`form-input ${errores.email ? 'form-input-error' : ''}`}
              placeholder="juan@email.com"
              value={form.email} onChange={handleChange}
              autoComplete="email"
            />
            {errores.email && <span className="form-error-msg">{errores.email}</span>}
          </div>

          <div className="form-grupo">
            <label className="form-etiqueta" htmlFor="password">Contraseña *</label>
            <input
              id="password" name="password" type="password"
              className={`form-input ${errores.password ? 'form-input-error' : ''}`}
              placeholder="Mínimo 6 caracteres"
              value={form.password} onChange={handleChange}
              autoComplete="new-password"
            />
            {errores.password && <span className="form-error-msg">{errores.password}</span>}
          </div>

          <div className="form-grupo">
            <label className="form-etiqueta" htmlFor="telefono">Teléfono (opcional)</label>
            <input
              id="telefono" name="telefono" type="tel"
              className="form-input"
              placeholder="+54 11 1234-5678"
              value={form.telefono} onChange={handleChange}
            />
          </div>

          <button
            type="submit"
            className="btn btn-primario"
            style={{ width: '100%', marginTop: 8 }}
            disabled={cargando}
          >
            {cargando
              ? <><div className="spinner" style={{ width: 18, height: 18, borderWidth: 2 }} /> Creando cuenta...</>
              : 'Crear mi cuenta'}
          </button>
        </form>

        <p className="auth-footer">
          ¿Ya tenés cuenta?{' '}
          <Link to="/login">Iniciá sesión</Link>
        </p>
      </div>
    </div>
  )
}
