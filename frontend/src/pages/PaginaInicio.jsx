import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { eventosAPI } from '../services/api'
import './PaginaInicio.css'

// Categorías disponibles para el filtro
const CATEGORIAS = [
  { valor: '', etiqueta: 'Todas las categorías' },
  { valor: 'musica', etiqueta: '🎵 Música' },
  { valor: 'deporte', etiqueta: '⚽ Deporte' },
  { valor: 'cultura', etiqueta: '🎭 Cultura' },
  { valor: 'teatro_cine', etiqueta: '🎬 Teatro y Cine' },
  { valor: 'conferencia', etiqueta: '💡 Conferencia' },
  { valor: 'otro', etiqueta: '📌 Otro' },
]

/**
 * TarjetaEvento — Muestra la información mínima de un evento en el catálogo
 */
function TarjetaEvento({ evento }) {
  const fecha = new Date(evento.fecha_hora)
  const fechaFormateada = fecha.toLocaleDateString('es-AR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
  const hora = fecha.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })

  const disponibles = evento.capacidad_total - evento.entradas_vendidas
  const porcentajeOcupacion = (evento.entradas_vendidas / evento.capacidad_total) * 100
  const agotado = disponibles <= 0 || evento.estado !== 'activo'

  return (
    <Link to={`/eventos/${evento.id}`} className="tarjeta-evento-link">
      <article className="tarjeta-evento">
        {/* Imagen del evento */}
        <div className="tarjeta-evento-imagen">
          {evento.imagen_url ? (
            <img src={evento.imagen_url} alt={evento.titulo} loading="lazy" />
          ) : (
            <div className="tarjeta-evento-imagen-placeholder">
              <span>🎫</span>
            </div>
          )}
          <div className={`tarjeta-evento-estado ${agotado ? 'agotado' : 'disponible'}`}>
            {agotado ? 'Agotado' : `${disponibles} disponibles`}
          </div>
        </div>

        {/* Info del evento */}
        <div className="tarjeta-evento-cuerpo">
          <span className={`badge badge-${evento.categoria}`}>
            {evento.categoria}
          </span>

          <h3 className="tarjeta-evento-titulo">{evento.titulo}</h3>

          <div className="tarjeta-evento-meta">
            <span>📅 {fechaFormateada}</span>
            <span>⏰ {hora}</span>
            <span>📍 {evento.lugar}</span>
          </div>

          {/* Barra de ocupación */}
          <div className="tarjeta-evento-ocupacion">
            <div className="ocupacion-barra">
              <div
                className="ocupacion-relleno"
                style={{ width: `${Math.min(porcentajeOcupacion, 100)}%` }}
              />
            </div>
          </div>

          <div className="tarjeta-evento-precio">
            {evento.precio_base === 0 ? (
              <span className="precio-gratis">GRATIS</span>
            ) : (
              <span className="precio-valor">
                ${evento.precio_base.toLocaleString('es-AR')}
              </span>
            )}
          </div>
        </div>
      </article>
    </Link>
  )
}

/**
 * PaginaInicio — Catálogo de eventos con búsqueda y filtros
 */
export default function PaginaInicio() {
  const [eventos, setEventos] = useState([])
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')
  const [filtros, setFiltros] = useState({
    busqueda: '',
    categoria: '',
    solo_disponibles: false,
  })

  const cargarEventos = useCallback(async () => {
    setCargando(true)
    setError('')
    try {
      const params = {}
      if (filtros.busqueda) params.busqueda = filtros.busqueda
      if (filtros.categoria) params.categoria = filtros.categoria
      if (filtros.solo_disponibles) params.solo_disponibles = true

      const response = await eventosAPI.listar(params)
      setEventos(response.data.datos?.eventos || [])
    } catch (err) {
      setError('No se pudieron cargar los eventos. Intentá de nuevo más tarde.')
      console.error('Error al cargar eventos:', err)
    } finally {
      setCargando(false)
    }
  }, [filtros])

  // Cargar eventos al montar y cuando cambian los filtros
  useEffect(() => {
    const timer = setTimeout(cargarEventos, 300) // Debounce 300ms para el texto
    return () => clearTimeout(timer)
  }, [cargarEventos])

  return (
    <div className="pagina-inicio">
      {/* Hero */}
      <section className="inicio-hero">
        <div className="contenedor">
          <h1 className="inicio-titulo">
            Encontrá tu próximo <span className="texto-acento">evento</span>
          </h1>
          <p className="inicio-subtitulo">
            Música, deporte, cultura y mucho más. Todo en un solo lugar.
          </p>
        </div>
      </section>

      {/* Barra de búsqueda y filtros */}
      <section className="inicio-filtros">
        <div className="contenedor">
          <div className="filtros-contenedor">
            {/* Búsqueda de texto */}
            <div className="filtro-busqueda">
              <span className="filtro-icono">🔍</span>
              <input
                type="text"
                placeholder="Buscar eventos, artistas, lugares..."
                className="filtro-input"
                value={filtros.busqueda}
                onChange={(e) => setFiltros(prev => ({ ...prev, busqueda: e.target.value }))}
              />
            </div>

            {/* Filtro de categoría */}
            <select
              className="form-input filtro-select"
              value={filtros.categoria}
              onChange={(e) => setFiltros(prev => ({ ...prev, categoria: e.target.value }))}
            >
              {CATEGORIAS.map(cat => (
                <option key={cat.valor} value={cat.valor}>{cat.etiqueta}</option>
              ))}
            </select>

            {/* Toggle de solo disponibles */}
            <label className="filtro-toggle">
              <input
                type="checkbox"
                checked={filtros.solo_disponibles}
                onChange={(e) => setFiltros(prev => ({ ...prev, solo_disponibles: e.target.checked }))}
              />
              <span>Solo disponibles</span>
            </label>
          </div>
        </div>
      </section>

      {/* Resultados */}
      <section className="inicio-resultados">
        <div className="contenedor">
          {cargando ? (
            <div className="cargando-contenedor">
              <div className="spinner" />
              <span>Buscando eventos...</span>
            </div>
          ) : error ? (
            <div className="alerta alerta-error">{error}</div>
          ) : eventos.length === 0 ? (
            <div className="sin-resultados">
              <span className="sin-resultados-icono">🎫</span>
              <h3>No hay eventos disponibles</h3>
              <p>Probá con otros filtros o volvé más tarde.</p>
            </div>
          ) : (
            <>
              <p className="resultados-contador">
                <strong>{eventos.length}</strong> evento{eventos.length !== 1 ? 's' : ''} encontrado{eventos.length !== 1 ? 's' : ''}
              </p>
              <div className="grilla-eventos">
                {eventos.map(evento => (
                  <TarjetaEvento key={evento.id} evento={evento} />
                ))}
              </div>
            </>
          )}
        </div>
      </section>
    </div>
  )
}
