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

const MAX_ENTRADAS_POR_COMPRA = 10

// ── Modal éxito — SIN código QR, ahora muestra cantidad y total ───
function ModalExito({ entradas, onCerrar }) {
  const cantidad = entradas?.length || 0
  const totalPagado = entradas?.reduce((suma, e) => suma + (e.precio_pagado || 0), 0) || 0

  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal modal-exito" onClick={(e) => e.stopPropagation()}>
        <div className="modal-exito-icono">🎉</div>
        <h2>{cantidad > 1 ? '¡Entradas confirmadas!' : '¡Entrada confirmada!'}</h2>
        <p className="modal-exito-subtitulo">Tu compra se procesó exitosamente.</p>

        <div className="modal-exito-detalle">
          <div className="detalle-fila">
            <span>Cantidad</span>
            <strong>{cantidad} {cantidad === 1 ? 'entrada' : 'entradas'}</strong>
          </div>
          <div className="detalle-fila">
            <span>Estado</span>
            <span className="badge badge-activa">Activa</span>
          </div>
          <div className="detalle-fila">
            <span>Total pagado</span>
            <strong className="texto-acento">
              ${totalPagado.toLocaleString('es-AR')}
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
  const [entradasCompradas, setEntradasCompradas] = useState(null)
  const [cantidad, setCantidad] = useState(1)

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

  // Tope real: no se puede pedir más que lo disponible ni más que el máximo por compra
  const disponiblesActuales = evento ? evento.capacidad_total - evento.entradas_vendidas : 0
  const topeCompra = Math.min(MAX_ENTRADAS_POR_COMPRA, disponiblesActuales)

  const handleCambiarCantidad = (delta) => {
    setCantidad(prev => {
      const nueva = prev + delta
      if (nueva < 1) return 1
      if (nueva > topeCompra) return topeCompra
      return nueva
    })
  }

  const handleComprar = async () => {
    if (!estaAutenticado) {
      navigate('/login')
      return
    }

    setComprando(true)
    setErrorCompra('')

    try {
      const response = await entradasAPI.comprar(evento.id, cantidad)
      const entradas = response.data.datos?.entradas || []

      // Actualizar disponibilidad en pantalla sin recargar
      setEvento(prev => ({
        ...prev,
        entradas_vendidas: prev.entradas_vendidas + cantidad
      }))
      setEntradasCompradas(entradas)
      setCantidad(1) // resetear el selector para la próxima compra
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
  const precioTotal = evento.precio_base * cantidad

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
                    <span className="dato-valor">{hora} hs — duración aprox. {Math.round(evento.duracion_minutos / 60)} hs</span>
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

              {/* Selector de cantidad — solo si hay más de 1 disponible y no está agotado */}
              {!agotado && estaAutenticado && (
                <div className="selector-cantidad">
                  <span className="selector-cantidad-label">Cantidad de entradas</span>
                  <div className="selector-cantidad-controles">
                    <button
                      type="button"
                      className="btn-cantidad"
                      onClick={() => handleCambiarCantidad(-1)}
                      disabled={cantidad <= 1 || comprando}
                      aria-label="Restar una entrada"
                    >
                      −
                    </button>
                    <span className="selector-cantidad-valor">{cantidad}</span>
                    <button
                      type="button"
                      className="btn-cantidad"
                      onClick={() => handleCambiarCantidad(1)}
                      disabled={cantidad >= topeCompra || comprando}
                      aria-label="Sumar una entrada"
                    >
                      +
                    </button>
                  </div>
                  {topeCompra < MAX_ENTRADAS_POR_COMPRA && (
                    <span className="selector-cantidad-hint">
                      Máximo {topeCompra} {topeCompra === 1 ? 'entrada' : 'entradas'} (disponibilidad limitada)
                    </span>
                  )}
                </div>
              )}

              {!agotado && cantidad > 1 && evento.precio_base > 0 && (
                <div className="sidebar-total">
                  <span>Total ({cantidad} entradas)</span>
                  <strong className="texto-acento">${precioTotal.toLocaleString('es-AR')}</strong>
                </div>
              )}

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
                  ) : estaAutenticado ? (
                    cantidad > 1 ? `🎫 Comprar ${cantidad} entradas` : '🎫 Comprar entrada'
                  ) : (
                    'Iniciá sesión para comprar'
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

      {entradasCompradas && (
        <ModalExito
          entradas={entradasCompradas}
          onCerrar={() => setEntradasCompradas(null)}
        />
      )}
    </div>
  )
}
