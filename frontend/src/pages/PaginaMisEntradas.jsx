import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { entradasAPI, transporteAPI } from '../services/api'
import './PaginaMisEntradas.css'

const ETIQUETA_ESTADO = {
  activa: { label: 'Activa', clase: 'badge-activa' },
  cancelada: { label: 'Cancelada', clase: 'badge-cancelada' },
  usada: { label: 'Usada', clase: 'badge-usada' },
  transferida: { label: 'Transferida', clase: 'badge-transferida' },
}

/* ── Modal Cancelar ─────────────────────────────────────────────── */
function ModalCancelar({ entrada, onConfirmar, onCerrar, cargando }) {
  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2 style={{ marginBottom: 12 }}>Cancelar entrada</h2>
        <p className="texto-secundario" style={{ marginBottom: 20 }}>
          ¿Estás seguro que querés cancelar tu entrada para{' '}
          <strong style={{ color: 'var(--color-texto)' }}>{entrada?.evento?.titulo}</strong>?
          Esta acción no se puede deshacer.
        </p>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
          <button className="btn btn-secundario" onClick={onCerrar} disabled={cargando}>
            No, mantener
          </button>
          <button className="btn btn-peligro" onClick={onConfirmar} disabled={cargando}>
            {cargando
              ? <><div className="spinner" style={{ width: 16, height: 16, borderWidth: 2 }} /> Cancelando...</>
              : 'Sí, cancelar entrada'}
          </button>
        </div>
      </div>
    </div>
  )
}

/* ── Modal Transferir ───────────────────────────────────────────── */
function ModalTransferir({ entrada, onConfirmar, onCerrar, cargando }) {
  const [email, setEmail] = useState('')
  const [errorEmail, setErrorEmail] = useState('')

  const handleConfirmar = () => {
    if (!email.trim()) { setErrorEmail('Ingresá un email.'); return }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { setErrorEmail('Email inválido.'); return }
    onConfirmar(email)
  }

  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2 style={{ marginBottom: 8 }}>Transferir entrada</h2>
        <p className="texto-secundario" style={{ marginBottom: 20 }}>
          Transferís tu entrada para{' '}
          <strong style={{ color: 'var(--color-texto)' }}>{entrada?.evento?.titulo}</strong>.
          El destinatario debe estar registrado en el sistema.
        </p>
        <div className="form-grupo">
          <label className="form-etiqueta">Email del destinatario</label>
          <input
            type="email"
            className={`form-input ${errorEmail ? 'form-input-error' : ''}`}
            placeholder="destinatario@email.com"
            value={email}
            onChange={e => { setEmail(e.target.value); setErrorEmail('') }}
          />
          {errorEmail && <span className="form-error-msg">{errorEmail}</span>}
        </div>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
          <button className="btn btn-secundario" onClick={onCerrar} disabled={cargando}>
            Cancelar
          </button>
          <button className="btn btn-primario" onClick={handleConfirmar} disabled={cargando}>
            {cargando
              ? <><div className="spinner" style={{ width: 16, height: 16, borderWidth: 2 }} /> Transfiriendo...</>
              : 'Confirmar transferencia'}
          </button>
        </div>
      </div>
    </div>
  )
}

/* ── Modal Asistente de Transporte ──────────────────────────────────
   Flujo:
   1. Si la entrada ya tiene transporte configurado, se carga directo
      en la vista de detalle (colectivo o auto_propio).
   2. Si no tiene nada, se muestra el selector de 3 modos.
   3. Colectivo: elegir línea -> guardar -> ver línea elegida + link horarios.
   4. Auto propio: elegir si comparte o no -> guardar -> ver estacionamientos.
   El modo "compartido" (buscar auto de otro) se agrega en la próxima parte;
   por ahora ese botón usa el mismo flujo de auto propio para no dejar un
   estado roto en la UI mientras se construye esa parte.
──────────────────────────────────────────────────────────────────── */
function ModalAsistenteTransporte({ entrada, onCerrar }) {
  const [cargando, setCargando] = useState(true)
  const [guardando, setGuardando] = useState(false)
  const [error, setError] = useState('')
  const [asistente, setAsistente] = useState(null)
  const [lineasColectivo] = useState([
    { linea: 'Línea 1', recorrido: 'Centro - Estadio', url_horario: 'https://cordoba.gob.ar/tu-bondi/' },
    { linea: 'Línea 7', recorrido: 'Terminal - Centro', url_horario: 'https://cordoba.gob.ar/tu-bondi/' },
    { linea: 'Línea 14', recorrido: 'Barrio Norte - Estadio', url_horario: 'https://cordoba.gob.ar/tu-bondi/' },
    { linea: 'Línea 20', recorrido: 'Circunvalación', url_horario: 'https://cordoba.gob.ar/tu-bondi/' },
  ])
  const [estacionamientos] = useState([
    { nombre: 'Parking Centro', direccion: 'Av. Principal 450', distancia: '200m del lugar' },
    { nombre: 'Estacionamiento Plaza', direccion: 'San Martín 120', distancia: '350m del lugar' },
    { nombre: 'Parking Estadio', direccion: 'Av. del Estadio 10', distancia: '100m del lugar' },
  ])
  const [lineaSeleccionada, setLineaSeleccionada] = useState('')

  const cargarAsistente = async () => {
    setCargando(true)
    setError('')
    try {
      const res = await transporteAPI.obtenerPorEntrada(entrada.id)
      const datos = res.data.datos
      setAsistente(datos?.asistente || null)
    } catch {
      setAsistente(null)
    } finally {
      setCargando(false)
    }
  }

  useEffect(() => { cargarAsistente() }, [])

  const guardarModo = async (datosExtra) => {
    setGuardando(true)
    setError('')
    try {
      const res = await transporteAPI.configurar({
        entrada_id: entrada.id,
        ...datosExtra,
      })
      const datos = res.data.datos
      setAsistente(datos?.asistente || null)
    } catch (err) {
      setError(err.response?.data?.error || 'No se pudo guardar la configuración.')
    } finally {
      setGuardando(false)
    }
  }

  const elegirColectivo = () => {
    // No llamamos al backend todavía -- el modo "colectivo" recién se guarda
    // cuando el usuario elige una línea específica (el backend exige
    // linea_colectivo si modo=colectivo). Mientras tanto, mostramos el
    // selector de líneas con un estado "en memoria" temporal.
    setAsistente({ modo: 'colectivo', linea_colectivo: '' })
  }
  const elegirAutoPropio = () => guardarModo({ modo: 'auto_propio', comparte_auto: false })

  const confirmarLinea = () => {
    if (!lineaSeleccionada) { setError('Elegí una línea de colectivo.'); return }
    guardarModo({ modo: 'colectivo', linea_colectivo: lineaSeleccionada })
  }

  const cambiarComparteAuto = (comparte) => {
    guardarModo({ modo: 'auto_propio', comparte_auto: comparte })
  }

  const handleCambiarModo = () => {
    setAsistente(null)
    setLineaSeleccionada('')
    setError('')
  }

  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal modal-transporte" onClick={e => e.stopPropagation()}>
        <div className="modal-admin-header">
          <h2>Asistente de transporte</h2>
          <button className="btn-cerrar" onClick={onCerrar}>X</button>
        </div>
        <p className="texto-secundario" style={{ marginBottom: 16 }}>
          {entrada.evento?.titulo}
        </p>

        {error && <div className="alerta alerta-error" style={{ marginBottom: 16 }}>{error}</div>}

        {cargando ? (
          <div className="cargando-contenedor" style={{ minHeight: 120 }}>
            <div className="spinner" /><span>Cargando...</span>
          </div>
        ) : !asistente ? (
          <div className="transporte-opciones">
            <button className="transporte-opcion" onClick={elegirColectivo} disabled={guardando}>
              <span className="transporte-opcion-icono">Bus</span>
              <span className="transporte-opcion-titulo">Colectivo</span>
              <span className="transporte-opcion-desc">Ver líneas y horarios</span>
            </button>
            <button className="transporte-opcion" onClick={elegirAutoPropio} disabled={guardando}>
              <span className="transporte-opcion-icono">Auto</span>
              <span className="transporte-opcion-titulo">Auto propio</span>
              <span className="transporte-opcion-desc">Estacionamientos cercanos</span>
            </button>
            <button className="transporte-opcion" onClick={elegirAutoPropio} disabled={guardando}>
              <span className="transporte-opcion-icono">Compartir</span>
              <span className="transporte-opcion-titulo">Compartido</span>
              <span className="transporte-opcion-desc">Viajar con otro asistente</span>
            </button>
          </div>
        ) : asistente.modo === 'colectivo' && !asistente.linea_colectivo ? (
          <div>
            <h3 className="transporte-subtitulo">Elegí tu línea</h3>
            <div className="transporte-lista-lineas">
              {lineasColectivo.map(l => (
                <label key={l.linea} className="transporte-linea-item">
                  <input
                    type="radio"
                    name="linea"
                    value={l.linea}
                    checked={lineaSeleccionada === l.linea}
                    onChange={() => setLineaSeleccionada(l.linea)}
                  />
                  <div>
                    <strong>{l.linea}</strong>
                    <span className="texto-secundario"> — {l.recorrido}</span>
                  </div>
                </label>
              ))}
            </div>
            <div className="modal-admin-footer">
              <button className="btn btn-secundario" onClick={handleCambiarModo} disabled={guardando}>
                Volver
              </button>
              <button className="btn btn-primario" onClick={confirmarLinea} disabled={guardando}>
                {guardando ? 'Guardando...' : 'Confirmar línea'}
              </button>
            </div>
          </div>
        ) : asistente.modo === 'colectivo' ? (
          <div>
            <div className="transporte-resumen">
              <span className="transporte-resumen-icono">Bus</span>
              <div>
                <strong>{asistente.linea_colectivo}</strong>
                <p className="texto-secundario" style={{ margin: 0 }}>
                  {lineasColectivo.find(l => l.linea === asistente.linea_colectivo)?.recorrido}
                </p>
              </div>
            </div>
            <a
              className="btn btn-primario"
              style={{ width: '100%', textAlign: 'center', display: 'block', textDecoration: 'none' }}
              href={lineasColectivo.find(l => l.linea === asistente.linea_colectivo)?.url_horario || '#'}
              target="_blank"
              rel="noopener noreferrer"
            >
              Ver horarios de la línea
            </a>
            <button className="btn btn-secundario" style={{ width: '100%', marginTop: 12 }} onClick={handleCambiarModo}>
              Cambiar de medio de transporte
            </button>
          </div>
        ) : (
          <div>
            <h3 className="transporte-subtitulo">¿Querés compartir tu viaje?</h3>
            <p className="texto-secundario" style={{ marginBottom: 16 }}>
              Si compartís, otros asistentes van a poder solicitarte unirse al viaje.
            </p>
            <div className="transporte-toggle-comparte">
              <button
                className={`btn ${asistente.comparte_auto ? 'btn-primario' : 'btn-secundario'}`}
                onClick={() => cambiarComparteAuto(true)}
                disabled={guardando}
              >
                Sí, compartir
              </button>
              <button
                className={`btn ${!asistente.comparte_auto ? 'btn-primario' : 'btn-secundario'}`}
                onClick={() => cambiarComparteAuto(false)}
                disabled={guardando}
              >
                No, voy solo/a
              </button>
            </div>

            {asistente.comparte_auto && (
              <div className="alerta alerta-exito" style={{ marginTop: 16 }}>
                Tu auto está disponible para compartir. Cuando alguien solicite
                unirse, vas a ver la solicitud acá mismo para aprobarla o rechazarla.
              </div>
            )}

            <h3 className="transporte-subtitulo" style={{ marginTop: 20 }}>Estacionamientos cercanos</h3>
            <div className="transporte-lista-estacionamientos">
              {estacionamientos.map(e => (
                <div key={e.nombre} className="transporte-estacionamiento-item">
                  <strong>{e.nombre}</strong>
                  <span className="texto-secundario">{e.direccion} — {e.distancia}</span>
                </div>
              ))}
            </div>

            <button className="btn btn-secundario" style={{ width: '100%', marginTop: 16 }} onClick={handleCambiarModo}>
              Cambiar de medio de transporte
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

/* ── Tarjeta Entrada ─────────────────────────────────────────────── */
function TarjetaEntrada({ entrada, onCancelar, onTransferir, onTransporte }) {
  const fechaCompra = new Date(entrada.fecha_compra).toLocaleDateString('es-AR', {
    day: 'numeric', month: 'short', year: 'numeric'
  })
  const estado = ETIQUETA_ESTADO[entrada.estado] || { label: entrada.estado, clase: '' }
  const puedeAccionar = entrada.estado === 'activa'

  const fechaEvento = entrada.evento ? new Date(entrada.evento.fecha_hora) : null
  const fechaEventoStr = fechaEvento
    ? fechaEvento.toLocaleDateString('es-AR', {
      weekday: 'short', day: 'numeric', month: 'short', year: 'numeric'
    })
    : '—'
  const horaEvento = fechaEvento
    ? fechaEvento.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit' })
    : ''

  return (
    <article className="entrada-card">
      <div className="entrada-card-header">
        <div>
          <h3 className="entrada-titulo">
            {entrada.evento?.titulo || `Evento #${entrada.evento_id}`}
          </h3>
          <p className="entrada-lugar">{entrada.evento?.lugar}</p>
        </div>
        <span className={`badge ${estado.clase}`}>{estado.label}</span>
      </div>

      <div className="separador" />

      <div className="entrada-detalles">
        <div className="entrada-dato">
          <span>Fecha del evento</span>
          <strong>{fechaEventoStr} {horaEvento && `· ${horaEvento}hs`}</strong>
        </div>
        <div className="entrada-dato">
          <span>Comprada el</span>
          <strong>{fechaCompra}</strong>
        </div>
        <div className="entrada-dato">
          <span>Precio pagado</span>
          <strong className="texto-acento">
            {entrada.precio_pagado === 0
              ? 'GRATIS'
              : `$${entrada.precio_pagado.toLocaleString('es-AR')}`}
          </strong>
        </div>
      </div>

      {puedeAccionar && (
        <div className="entrada-acciones">
          <button
            className="btn btn-secundario btn-sm"
            onClick={() => onTransporte(entrada)}
          >
            Transporte
          </button>
          <button
            className="btn btn-secundario btn-sm"
            onClick={() => onTransferir(entrada)}
          >
            Transferir
          </button>
          <button
            className="btn btn-peligro btn-sm"
            onClick={() => onCancelar(entrada)}
          >
            Cancelar
          </button>
        </div>
      )}
    </article>
  )
}

/* ── Página Principal ────────────────────────────────────────────── */
export default function PaginaMisEntradas() {
  const [entradas, setEntradas] = useState([])
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')
  const [mensajeExito, setMensajeExito] = useState('')

  const [entradaACancelar, setEntradaACancelar] = useState(null)
  const [entradaATransferir, setEntradaATransferir] = useState(null)
  const [entradaTransporte, setEntradaTransporte] = useState(null)
  const [accionCargando, setAccionCargando] = useState(false)

  const cargarEntradas = async () => {
    setCargando(true)
    setError('')
    try {
      const response = await entradasAPI.misEntradas()
      setEntradas(response.data.datos?.entradas || [])
    } catch {
      setError('No se pudieron cargar tus entradas.')
    } finally {
      setCargando(false)
    }
  }

  useEffect(() => { cargarEntradas() }, [])

  const handleCancelarConfirmar = async () => {
    setAccionCargando(true)
    try {
      await entradasAPI.cancelar(entradaACancelar.id)
      setMensajeExito('Entrada cancelada correctamente.')
      setEntradaACancelar(null)
      cargarEntradas()
    } catch (err) {
      setError(err.response?.data?.error || 'Error al cancelar la entrada.')
      setEntradaACancelar(null)
    } finally {
      setAccionCargando(false)
    }
  }

  const handleTransferirConfirmar = async (emailDestinatario) => {
    setAccionCargando(true)
    try {
      await entradasAPI.transferir(entradaATransferir.id, emailDestinatario)
      setMensajeExito(`Entrada transferida a ${emailDestinatario} correctamente.`)
      setEntradaATransferir(null)
      cargarEntradas()
    } catch (err) {
      setError(err.response?.data?.error || 'Error al transferir la entrada.')
      setEntradaATransferir(null)
    } finally {
      setAccionCargando(false)
    }
  }

  const activas = entradas.filter(e => e.estado === 'activa')
  const resto = entradas.filter(e => e.estado !== 'activa')

  return (
    <div className="pagina-mis-entradas">
      <div className="contenedor">
        <div className="mis-entradas-header">
          <div>
            <h1>Mis Entradas</h1>
            <p className="texto-secundario">Historial de todas tus entradas adquiridas</p>
          </div>
          <Link to="/" className="btn btn-secundario btn-sm">Explorar eventos</Link>
        </div>

        {error && (
          <div className="alerta alerta-error" style={{ marginBottom: 24 }}>
            {error}
            <button style={{ marginLeft: 12, background: 'none', border: 'none', color: 'inherit', cursor: 'pointer' }}
              onClick={() => setError('')}>X</button>
          </div>
        )}
        {mensajeExito && (
          <div className="alerta alerta-exito" style={{ marginBottom: 24 }}>
            {mensajeExito}
            <button style={{ marginLeft: 12, background: 'none', border: 'none', color: 'inherit', cursor: 'pointer' }}
              onClick={() => setMensajeExito('')}>X</button>
          </div>
        )}

        {cargando ? (
          <div className="cargando-contenedor"><div className="spinner" /><span>Cargando entradas...</span></div>
        ) : entradas.length === 0 ? (
          <div className="sin-entradas">
            <h3>Todavía no compraste entradas</h3>
            <p>Explorá el catálogo y conseguí tus tickets favoritos.</p>
            <Link to="/" className="btn btn-primario" style={{ marginTop: 16 }}>Ver eventos</Link>
          </div>
        ) : (
          <>
            {activas.length > 0 && (
              <section className="mis-entradas-seccion">
                <h2 className="seccion-titulo">Entradas activas ({activas.length})</h2>
                <div className="entradas-grilla">
                  {activas.map(e => (
                    <TarjetaEntrada key={e.id} entrada={e}
                      onCancelar={setEntradaACancelar}
                      onTransferir={setEntradaATransferir}
                      onTransporte={setEntradaTransporte} />
                  ))}
                </div>
              </section>
            )}
            {resto.length > 0 && (
              <section className="mis-entradas-seccion">
                <h2 className="seccion-titulo">Historial ({resto.length})</h2>
                <div className="entradas-grilla">
                  {resto.map(e => (
                    <TarjetaEntrada key={e.id} entrada={e}
                      onCancelar={setEntradaACancelar}
                      onTransferir={setEntradaATransferir}
                      onTransporte={setEntradaTransporte} />
                  ))}
                </div>
              </section>
            )}
          </>
        )}
      </div>

      {entradaACancelar && (
        <ModalCancelar entrada={entradaACancelar}
          onConfirmar={handleCancelarConfirmar}
          onCerrar={() => setEntradaACancelar(null)}
          cargando={accionCargando} />
      )}
      {entradaATransferir && (
        <ModalTransferir entrada={entradaATransferir}
          onConfirmar={handleTransferirConfirmar}
          onCerrar={() => setEntradaATransferir(null)}
          cargando={accionCargando} />
      )}
      {entradaTransporte && (
        <ModalAsistenteTransporte entrada={entradaTransporte}
          onCerrar={() => setEntradaTransporte(null)} />
      )}
    </div>
  )
}
