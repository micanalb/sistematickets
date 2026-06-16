import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { eventosAPI, entradasAPI } from '../services/api'
import { useAuth } from '../context/AuthContext'
import './PaginaDetalle.css'

const ETIQUETAS_CATEGORIA = {
  musica: '🎵 Música',
  deporte: '⚽ Deporte',
  cultura: '🎭 Cultura',
  teatro_cine: '🎬 Teatro y Cine',
  conferencia: '💡 Conferencia',
  otro: '📌 Otro',
}

// ── Modal éxito — SIN código QR ───────────────────────────────────
function ModalExito({ entrada, onCerrar }) {
  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal modal-exito" onClick={(e) => e.stopPropagation()}>
        <div className="modal-exito-icono">🎉</div>
        <h2>¡Entrada confirmada!</h2>
        <p className="modal-exito-subtitulo">Tu compra se procesó exitosamente.</p>

        <div className="modal-exito-detalle">
          <div className="detalle-fila">
            <span>Estado</span>
            <span className="badge badge-activa">Activa</span>
          </div>
          <div className="detalle-fila">
            <span>Precio pagado</span>
            <strong className="texto-acento">
              ${entrada?.precio_pagado?.toLocaleString('es-AR')}
            </strong>
          </div>
        </div>

        <div className="modal-exito-acciones">
          <Link to="/mis-entradas" className="btn btn-primario" onClick={onCerrar}>
            Ver mis entradas
          </Link>
          <button className="btn btn-secundario" onClick={onCerrar}>
            Seguir explorando
          </button>
        </div>
      </div>
    </div>
  )
}

export default function PaginaDetalle() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { estaAutenticado } = useAuth()

  const [evento, setEvento] = useState(null)
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')
  const [comprando, setComprando] = useState(false)
  const [errorCompra, setErrorCompra] = useState('')
  const [entradaComprada, setEntradaComprada] = useState(null)

  useEffect(() => {
    const cargar = async () => {
      try {
        const response = await eventosAPI.obtener(id)
        setEvento(response.data.datos)
      } catch (err) {
        if (err.response?.status === 404) {
          setError('El evento no existe o fue eliminado.')
        } else {
          setError('Error al cargar el evento.')
        }
      } finally {
        setCargando(false)
      }
    }
    cargar()
  }, [id])

  const handleComprar = async () => {
    if (!estaAutenticado) {
      navigate('/login')
      return
    }

    setComprando(true)
    setErrorCompra('')

    try {
      const response = await entradasAPI.comprar(evento.id)
      // Actualizar disponibilidad en pantalla sin recargar
      setEvento(prev => ({
        ...prev,
        entradas_vendidas: prev.entradas_vendidas + 1
      }))
      setEntradaComprada(response.data.datos)
    } catch (err) {
      const mensaje = err.response?.data?.error || 'No se pudo procesar la compra.'
      setErrorCompra(mensaje)
    } finally {
      setComprando(false)
    }
  }

  if (cargando) {
    return (
      <div className="cargando-contenedor" style={{ minHeight: 'calc(100vh - 64px)' }}>
        <div className="spinner" />
        <span>Cargando evento...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="contenedor" style={{ padding: '80px 24px', textAlign: 'center' }}>
        <div className="alerta alerta-error" style={{ maxWidth: 500, margin: '0 auto 24px' }}>
          {error}
        </div>
        <Link to="/" className="btn btn-secundario">← Volver al catálogo</Link>
      </div>
    )
  }

  if (!evento) return null

  const fecha = new Date(evento.fecha_hora)
  const fechaFormateada = fecha.toLocaleDateString('es-AR', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric'
  })
  const hora = fecha.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })
  const disponibles = evento.capacidad_total - evento.entradas_vendidas
  const porcentajeOcupacion = (evento.entradas_vendidas / evento.capacidad_total) * 100
  const agotado = disponibles <= 0 || evento.estado !== 'activo'

  return (
    <div className="pagina-detalle">
      <div className="contenedor">
        <Link to="/" className="detalle-volver">← Volver al catálogo</Link>

        <div className="detalle-layout">
          {/* Columna principal */}
          <div className="detalle-principal">
            <div className="detalle-imagen">
              {evento.imagen_url ? (
                <img src={evento.imagen_url} alt={evento.titulo} />
              ) : (
                <div className="detalle-imagen-placeholder">🎫</div>
              )}
            </div>

            <div className="detalle-info">
              <span className={`badge badge-${evento.categoria}`}>
                {ETIQUETAS_CATEGORIA[evento.categoria] || evento.categoria}
              </span>

              <h1 className="detalle-titulo">{evento.titulo}</h1>

              {evento.descripcion && (
                <p className="detalle-descripcion">{evento.descripcion}</p>
              )}

              <div className="separador" />

              <div className="detalle-datos">
                <div className="dato-item">
                  <span className="dato-icono">📅</span>
                  <div>
                    <span className="dato-etiqueta">Fecha</span>
                    <span className="dato-valor">{fechaFormateada}</span>
                  </div>
                </div>
                <div className="dato-item">
                  <span className="dato-icono">⏰</span>
                  <div>
                    <span className="dato-etiqueta">Horario</span>
                    <span className="dato-valor">{hora} hs — duración aprox. {evento.duracion_minutos} min</span>
                  </div>
                </div>
                <div className="dato-item">
                  <span className="dato-icono">📍</span>
                  <div>
                    <span className="dato-etiqueta">Lugar</span>
                    <span className="dato-valor">{evento.lugar}</span>
                  </div>
                </div>
                {evento.direccion && (
                  <div className="dato-item">
                    <span className="dato-icono">🗺️</span>
                    <div>
                      <span className="dato-etiqueta">Dirección</span>
                      <span className="dato-valor">
                        {evento.direccion}{evento.ciudad ? `, ${evento.ciudad}` : ''}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <aside className="detalle-sidebar">
            <div className="sidebar-card">
              <div className="sidebar-precio">
                {evento.precio_base === 0 ? (
                  <span className="precio-gratis">GRATIS</span>
                ) : (
                  <>
                    <span className="precio-label">Precio por entrada</span>
                    <span className="precio-valor-grande">
                      ${evento.precio_base.toLocaleString('es-AR')}
                    </span>
                  </>
                )}
              </div>

              <div className="sidebar-disponibilidad">
                <div className="disponibilidad-texto">
                  <span>
                    {agotado
                      ? '❌ Sin entradas disponibles'
                      : `✅ ${disponibles.toLocaleString('es-AR')} entradas disponibles`}
                  </span>
                  <span className="texto-secundario">
                    {evento.entradas_vendidas.toLocaleString('es-AR')} / {evento.capacidad_total.toLocaleString('es-AR')} vendidas
                  </span>
                </div>
                <div className="ocupacion-barra" style={{ height: 6 }}>
                  <div
                    className="ocupacion-relleno"
                    style={{ width: `${Math.min(porcentajeOcupacion, 100)}%` }}
                  />
                </div>
              </div>

              <div className="separador" />

              {errorCompra && (
                <div className="alerta alerta-error">{errorCompra}</div>
              )}

              {agotado ? (
                <button className="btn btn-secundario" style={{ width: '100%' }} disabled>
                  Entradas agotadas
                </button>
              ) : (
                <button
                  className="btn btn-primario"
                  style={{ width: '100%' }}
                  onClick={handleComprar}
                  disabled={comprando}
                >
                  {comprando ? (
                    <>
                      <div className="spinner" style={{ width: 18, height: 18, borderWidth: 2 }} />
                      Procesando...
                    </>
                  ) : (
                    estaAutenticado ? '🎫 Comprar entrada' : 'Iniciá sesión para comprar'
                  )}
                </button>
              )}

              {!estaAutenticado && !agotado && (
                <p className="sidebar-login-hint">
                  ¿No tenés cuenta? <Link to="/registro">Registrate gratis</Link>
                </p>
              )}
            </div>
          </aside>
        </div>
      </div>

      {entradaComprada && (
        <ModalExito
          entrada={entradaComprada}
          onCerrar={() => setEntradaComprada(null)}
        />
      )}
    </div>
  )
}
