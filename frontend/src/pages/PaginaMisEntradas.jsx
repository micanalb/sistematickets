import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { entradasAPI } from '../services/api'
import './PaginaMisEntradas.css'

const ETIQUETA_ESTADO = {
  activa:      { label: 'Activa',      clase: 'badge-activa' },
  cancelada:   { label: 'Cancelada',   clase: 'badge-cancelada' },
  usada:       { label: 'Usada',       clase: 'badge-usada' },
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

/* ── Tarjeta Entrada ─────────────────────────────────────────────── */
function TarjetaEntrada({ entrada, onCancelar, onTransferir }) {
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
          <span>📅 Fecha del evento</span>
          <strong>{fechaEventoStr} {horaEvento && `· ${horaEvento}hs`}</strong>
        </div>
        <div className="entrada-dato">
          <span>🗓️ Comprada el</span>
          <strong>{fechaCompra}</strong>
        </div>
        <div className="entrada-dato">
          <span>💰 Precio pagado</span>
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
            onClick={() => onTransferir(entrada)}
          >
            ↗ Transferir
          </button>
          <button
            className="btn btn-peligro btn-sm"
            onClick={() => onCancelar(entrada)}
          >
            ✕ Cancelar
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
  const resto   = entradas.filter(e => e.estado !== 'activa')

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
              onClick={() => setError('')}>✕</button>
          </div>
        )}
        {mensajeExito && (
          <div className="alerta alerta-exito" style={{ marginBottom: 24 }}>
            {mensajeExito}
            <button style={{ marginLeft: 12, background: 'none', border: 'none', color: 'inherit', cursor: 'pointer' }}
              onClick={() => setMensajeExito('')}>✕</button>
          </div>
        )}

        {cargando ? (
          <div className="cargando-contenedor"><div className="spinner" /><span>Cargando entradas...</span></div>
        ) : entradas.length === 0 ? (
          <div className="sin-entradas">
            <span className="sin-entradas-icono">🎫</span>
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
                      onTransferir={setEntradaATransferir} />
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
                      onTransferir={setEntradaATransferir} />
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
    </div>
  )
}
