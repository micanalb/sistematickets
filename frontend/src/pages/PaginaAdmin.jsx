import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { eventosAPI } from '../services/api'
import { useAuth } from '../context/AuthContext'
import './PaginaAdmin.css'

const CATEGORIAS = [
  { valor: 'musica', etiqueta: 'Musica' },
  { valor: 'deporte', etiqueta: 'Deporte' },
  { valor: 'cultura', etiqueta: 'Cultura' },
  { valor: 'teatro_cine', etiqueta: 'Teatro y Cine' },
  { valor: 'conferencia', etiqueta: 'Conferencia' },
  { valor: 'otro', etiqueta: 'Otro' },
]

const ESTADOS = [
  { valor: 'activo', etiqueta: 'Activo' },
  { valor: 'cancelado', etiqueta: 'Cancelado' },
  { valor: 'agotado', etiqueta: 'Agotado' },
  { valor: 'finalizado', etiqueta: 'Finalizado' },
]

const EXTENSIONES_PERMITIDAS = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
const TAMANIO_MAXIMO_MB = 5

const FORM_VACIO = {
  titulo: '', descripcion: '', fecha_hora: '', duracion_minutos: 120,
  lugar: '', direccion: '', ciudad: '', categoria: 'musica',
  capacidad_total: '', precio_base: '', imagen_url: '', estado: 'activo',
}

/* -- Modal Formulario (Crear / Editar) --
   Flujo de imagen: el formulario guarda los datos del evento (texto/numeros)
   igual que siempre. La imagen se maneja por separado en estado local
   (archivoImagen + previewImagen) porque el backend necesita el ID del
   evento para asociarla -- recien existe despues de crear, o ya existia
   si estamos editando. Por eso la subida ocurre DESPUES de onGuardar,
   orquestada desde el componente padre (PaginaAdmin).
*/
function ModalFormEvento({ evento, onGuardar, onCerrar, cargando, error, subiendoImagen }) {
  const [form, setForm] = useState(FORM_VACIO)
  const [archivoImagen, setArchivoImagen] = useState(null)
  const [previewImagen, setPreviewImagen] = useState('')
  const [errorImagen, setErrorImagen] = useState('')
  const inputArchivoRef = useRef(null)

  useEffect(() => {
    if (evento) {
      const fechaLocal = evento.fecha_hora
        ? new Date(evento.fecha_hora).toISOString().slice(0, 16)
        : ''
      setForm({
        titulo: evento.titulo || '',
        descripcion: evento.descripcion || '',
        fecha_hora: fechaLocal,
        duracion_minutos: evento.duracion_minutos || 120,
        lugar: evento.lugar || '',
        direccion: evento.direccion || '',
        ciudad: evento.ciudad || '',
        categoria: evento.categoria || 'musica',
        capacidad_total: evento.capacidad_total || '',
        precio_base: evento.precio_base || '',
        imagen_url: evento.imagen_url || '',
        estado: evento.estado || 'activo',
      })
      setPreviewImagen(evento.imagen_url || '')
    } else {
      setForm(FORM_VACIO)
      setPreviewImagen('')
    }
    setArchivoImagen(null)
    setErrorImagen('')
  }, [evento])

  const handleChange = (e) => {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  const handleSeleccionarImagen = (e) => {
    const archivo = e.target.files?.[0]
    setErrorImagen('')

    if (!archivo) return

    if (!EXTENSIONES_PERMITIDAS.includes(archivo.type)) {
      setErrorImagen('Formato no permitido. Usa JPG, PNG o WEBP.')
      return
    }

    const tamanioMB = archivo.size / (1024 * 1024)
    if (tamanioMB > TAMANIO_MAXIMO_MB) {
      setErrorImagen(`La imagen supera el tamano maximo (${TAMANIO_MAXIMO_MB}MB).`)
      return
    }

    setArchivoImagen(archivo)
    setPreviewImagen(URL.createObjectURL(archivo))
  }

  const handleQuitarImagen = () => {
    setArchivoImagen(null)
    setPreviewImagen('')
    if (inputArchivoRef.current) inputArchivoRef.current.value = ''
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    onGuardar(form, archivoImagen)
  }

  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal modal-admin" onClick={e => e.stopPropagation()}>
        <div className="modal-admin-header">
          <h2>{evento ? 'Editar evento' : 'Nuevo evento'}</h2>
          <button className="btn-cerrar" onClick={onCerrar}>X</button>
        </div>

        {error && <div className="alerta alerta-error">{error}</div>}

        <form onSubmit={handleSubmit} className="admin-form">
          <div className="admin-form-fila">
            <div className="form-grupo">
              <label className="form-etiqueta">Titulo *</label>
              <input name="titulo" className="form-input" value={form.titulo}
                onChange={handleChange} required placeholder="Nombre del evento" />
            </div>
            <div className="form-grupo">
              <label className="form-etiqueta">Categoria *</label>
              <select name="categoria" className="form-input" value={form.categoria} onChange={handleChange}>
                {CATEGORIAS.map(c => <option key={c.valor} value={c.valor}>{c.etiqueta}</option>)}
              </select>
            </div>
          </div>

          <div className="form-grupo">
            <label className="form-etiqueta">Descripcion</label>
            <textarea name="descripcion" className="form-input" rows={3}
              value={form.descripcion} onChange={handleChange}
              placeholder="Descripcion del evento" />
          </div>

          <div className="admin-form-fila">
            <div className="form-grupo">
              <label className="form-etiqueta">Fecha y hora *</label>
              <input name="fecha_hora" type="datetime-local" className="form-input"
                value={form.fecha_hora} onChange={handleChange} required />
            </div>
            <div className="form-grupo">
              <label className="form-etiqueta">Duracion (minutos) *</label>
              <input name="duracion_minutos" type="number" className="form-input"
                value={form.duracion_minutos} onChange={handleChange} min={1} required />
            </div>
          </div>

          <div className="admin-form-fila">
            <div className="form-grupo">
              <label className="form-etiqueta">Lugar *</label>
              <input name="lugar" className="form-input" value={form.lugar}
                onChange={handleChange} required placeholder="Estadio, teatro, etc." />
            </div>
            <div className="form-grupo">
              <label className="form-etiqueta">Ciudad</label>
              <input name="ciudad" className="form-input" value={form.ciudad}
                onChange={handleChange} placeholder="Buenos Aires" />
            </div>
          </div>

          <div className="form-grupo">
            <label className="form-etiqueta">Direccion</label>
            <input name="direccion" className="form-input" value={form.direccion}
              onChange={handleChange} placeholder="Av. Corrientes 1234" />
          </div>

          <div className="admin-form-fila">
            <div className="form-grupo">
              <label className="form-etiqueta">Capacidad total *</label>
              <input name="capacidad_total" type="number" className="form-input"
                value={form.capacidad_total} onChange={handleChange} min={1} required />
            </div>
            <div className="form-grupo">
              <label className="form-etiqueta">Precio base (ARS) *</label>
              <input name="precio_base" type="number" className="form-input"
                value={form.precio_base} onChange={handleChange} min={0} step="0.01" required />
            </div>
          </div>

          {/* -- Imagen del evento (subida real de archivo) -- */}
          <div className="form-grupo">
            <label className="form-etiqueta">Imagen del evento</label>

            {previewImagen && (
              <div className="imagen-preview-wrapper">
                <img src={previewImagen} alt="Vista previa" className="imagen-preview" />
                <button type="button" className="btn-quitar-imagen" onClick={handleQuitarImagen}>
                  X Quitar
                </button>
              </div>
            )}

            <input
              ref={inputArchivoRef}
              type="file"
              accept="image/jpeg,image/jpg,image/png,image/webp"
              onChange={handleSeleccionarImagen}
              className="form-input-file"
            />
            <span className="form-hint">JPG, PNG o WEBP -- maximo {TAMANIO_MAXIMO_MB}MB</span>
            {errorImagen && <span className="form-error-msg">{errorImagen}</span>}
            {subiendoImagen && (
              <span className="form-hint texto-acento">
                <div className="spinner" style={{ width: 14, height: 14, borderWidth: 2, display: 'inline-block', marginRight: 6 }} />
                Subiendo imagen...
              </span>
            )}
          </div>

          {evento && (
            <div className="form-grupo">
              <label className="form-etiqueta">Estado</label>
              <select name="estado" className="form-input" value={form.estado} onChange={handleChange}>
                {ESTADOS.map(e => <option key={e.valor} value={e.valor}>{e.etiqueta}</option>)}
              </select>
            </div>
          )}

          <div className="modal-admin-footer">
            <button type="button" className="btn btn-secundario" onClick={onCerrar} disabled={cargando}>
              Cancelar
            </button>
            <button type="submit" className="btn btn-primario" disabled={cargando}>
              {cargando
                ? <><div className="spinner" style={{ width: 16, height: 16, borderWidth: 2 }} /> Guardando...</>
                : evento ? 'Guardar cambios' : 'Crear evento'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

/* -- Modal Confirmar Eliminar -- */
function ModalConfirmarEliminar({ evento, onConfirmar, onCerrar, cargando }) {
  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2 style={{ marginBottom: 12 }}>Eliminar evento</h2>
        <p className="texto-secundario" style={{ marginBottom: 20 }}>
          Confirmas que queres eliminar <strong style={{ color: 'var(--color-texto)' }}>{evento?.titulo}</strong>?
          El evento se dara de baja pero las entradas ya vendidas se conservan.
        </p>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
          <button className="btn btn-secundario" onClick={onCerrar} disabled={cargando}>Cancelar</button>
          <button className="btn btn-peligro" onClick={onConfirmar} disabled={cargando}>
            {cargando ? 'Eliminando...' : 'Si, eliminar'}
          </button>
        </div>
      </div>
    </div>
  )
}

/* -- Modal Reporte -- */
function ModalReporte({ reporte, onCerrar }) {
  if (!reporte) return null
  const { evento, entradas_vendidas, capacidad_total, entradas_disponibles, porcentaje_ocupacion, compradores } = reporte
  return (
    <div className="modal-overlay" onClick={onCerrar}>
      <div className="modal modal-reporte" onClick={e => e.stopPropagation()}>
        <div className="modal-admin-header">
          <h2>Reporte -- {evento?.titulo}</h2>
          <button className="btn-cerrar" onClick={onCerrar}>X</button>
        </div>

        <div className="reporte-metricas">
          <div className="metrica-card">
            <span className="metrica-valor">{entradas_vendidas}</span>
            <span className="metrica-label">Vendidas</span>
          </div>
          <div className="metrica-card">
            <span className="metrica-valor">{entradas_disponibles}</span>
            <span className="metrica-label">Disponibles</span>
          </div>
          <div className="metrica-card">
            <span className="metrica-valor">{capacidad_total}</span>
            <span className="metrica-label">Capacidad</span>
          </div>
          <div className="metrica-card metrica-highlight">
            <span className="metrica-valor">{porcentaje_ocupacion?.toFixed(1)}%</span>
            <span className="metrica-label">Ocupacion</span>
          </div>
        </div>

        <div className="ocupacion-barra" style={{ height: 8, margin: '16px 0' }}>
          <div className="ocupacion-relleno" style={{ width: `${Math.min(porcentaje_ocupacion, 100)}%` }} />
        </div>

        <h3 style={{ marginBottom: 12, fontSize: '0.95rem' }}>
          Compradores ({compradores?.length || 0})
        </h3>

        {compradores?.length > 0 ? (
          <div className="reporte-tabla-wrapper">
            <table className="reporte-tabla">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Nombre</th>
                  <th>Email</th>
                  <th>Estado</th>
                  <th>Precio</th>
                </tr>
              </thead>
              <tbody>
                {compradores.map((entrada, i) => (
                  <tr key={entrada.id}>
                    <td>{i + 1}</td>
                    <td>{entrada.usuario ? `${entrada.usuario.nombre} ${entrada.usuario.apellido}` : '-'}</td>
                    <td>{entrada.usuario?.email || '-'}</td>
                    <td><span className={`badge badge-${entrada.estado}`}>{entrada.estado}</span></td>
                    <td>${entrada.precio_pagado?.toLocaleString('es-AR')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="texto-secundario">Todavia no hay entradas vendidas.</p>
        )}

        <div style={{ marginTop: 20, textAlign: 'right' }}>
          <button className="btn btn-secundario" onClick={onCerrar}>Cerrar</button>
        </div>
      </div>
    </div>
  )
}

/* -- Pagina Principal Admin -- */
export default function PaginaAdmin() {
  const { esAdmin, estaAutenticado } = useAuth()
  const navigate = useNavigate()

  const [eventos, setEventos] = useState([])
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')
  const [exito, setExito] = useState('')

  const [modalForm, setModalForm] = useState(false)
  const [eventoEditando, setEventoEditando] = useState(null)
  const [eventoEliminando, setEventoEliminando] = useState(null)
  const [reporte, setReporte] = useState(null)
  const [accionCargando, setAccionCargando] = useState(false)
  const [subiendoImagen, setSubiendoImagen] = useState(false)
  const [errorForm, setErrorForm] = useState('')

  useEffect(() => {
    if (!estaAutenticado || !esAdmin) {
      navigate('/')
      return
    }
    cargarEventos()
  }, [])

  const cargarEventos = async () => {
    setCargando(true)
    try {
      const res = await eventosAPI.listar()
      setEventos(res.data.datos?.eventos || [])
    } catch {
      setError('No se pudieron cargar los eventos.')
    } finally {
      setCargando(false)
    }
  }

  /* handleGuardar recibe tambien el archivo de imagen (puede ser null si
     no se seleccion ninguna). Flujo:
       1. Crear o actualizar el evento con los datos de texto/numero
       2. Si hay archivo, recien ahi subirlo con el ID del evento ya creado/existente
     Si el paso 2 falla, el evento ya quedo guardado igual -- se avisa por
     separado para no perder los datos de texto que si se guardaron bien. */
  const handleGuardar = async (form, archivoImagen) => {
    setAccionCargando(true)
    setErrorForm('')
    try {
      const datos = {
        ...form,
        duracion_minutos: parseInt(form.duracion_minutos),
        capacidad_total: parseInt(form.capacidad_total),
        precio_base: parseFloat(form.precio_base),
        fecha_hora: new Date(form.fecha_hora).toISOString(),
      }

      let eventoGuardado
      if (eventoEditando) {
        const res = await eventosAPI.actualizar(eventoEditando.id, datos)
        eventoGuardado = res.data.datos
      } else {
        const res = await eventosAPI.crear(datos)
        eventoGuardado = res.data.datos
      }

      if (archivoImagen && eventoGuardado?.id) {
        setAccionCargando(false)
        setSubiendoImagen(true)
        try {
          await eventosAPI.subirImagen(eventoGuardado.id, archivoImagen)
        } catch (errImg) {
          setError(
            errImg.response?.data?.error ||
            'El evento se guardo, pero la imagen no se pudo subir. Podes reintentarlo editando el evento.'
          )
        } finally {
          setSubiendoImagen(false)
        }
      }

      setExito(eventoEditando ? 'Evento actualizado correctamente.' : 'Evento creado correctamente.')
      setModalForm(false)
      setEventoEditando(null)
      cargarEventos()
    } catch (err) {
      setErrorForm(err.response?.data?.error || 'Error al guardar el evento.')
    } finally {
      setAccionCargando(false)
    }
  }

  const handleEliminar = async () => {
    setAccionCargando(true)
    try {
      await eventosAPI.eliminar(eventoEliminando.id)
      setExito('Evento eliminado correctamente.')
      setEventoEliminando(null)
      cargarEventos()
    } catch (err) {
      setError(err.response?.data?.error || 'Error al eliminar.')
      setEventoEliminando(null)
    } finally {
      setAccionCargando(false)
    }
  }

  const handleVerReporte = async (evento) => {
    try {
      const res = await eventosAPI.reporte(evento.id)
      setReporte(res.data.datos)
    } catch {
      setError('No se pudo cargar el reporte.')
    }
  }

  const abrirCrear = () => { setEventoEditando(null); setErrorForm(''); setModalForm(true) }
  const abrirEditar = (ev) => { setEventoEditando(ev); setErrorForm(''); setModalForm(true) }

  const fecha = (f) => new Date(f).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' })
  const ocupacion = (ev) => ev.capacidad_total > 0
    ? Math.round((ev.entradas_vendidas / ev.capacidad_total) * 100) : 0

  return (
    <div className="pagina-admin">
      <div className="contenedor">

        <div className="admin-header">
          <div>
            <h1>Panel de Administracion</h1>
            <p className="texto-secundario">Gestion de eventos del sistema</p>
          </div>
          <button className="btn btn-primario" onClick={abrirCrear}>
            + Nuevo evento
          </button>
        </div>

        {error && (
          <div className="alerta alerta-error" style={{ marginBottom: 20 }}>
            {error} <button className="btn-x" onClick={() => setError('')}>X</button>
          </div>
        )}
        {exito && (
          <div className="alerta alerta-exito" style={{ marginBottom: 20 }}>
            {exito} <button className="btn-x" onClick={() => setExito('')}>X</button>
          </div>
        )}

        {cargando ? (
          <div className="cargando-contenedor"><div className="spinner" /><span>Cargando...</span></div>
        ) : eventos.length === 0 ? (
          <div className="sin-resultados">
            <h3>No hay eventos cargados</h3>
            <p>Crea el primero con el boton de arriba.</p>
          </div>
        ) : (
          <div className="admin-tabla-wrapper">
            <table className="admin-tabla">
              <thead>
                <tr>
                  <th>Evento</th>
                  <th>Fecha</th>
                  <th>Categoria</th>
                  <th>Precio</th>
                  <th>Ocupacion</th>
                  <th>Estado</th>
                  <th>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {eventos.map(ev => (
                  <tr key={ev.id}>
                    <td>
                      <div className="tabla-evento-nombre">{ev.titulo}</div>
                      <div className="tabla-evento-lugar">{ev.lugar}</div>
                    </td>
                    <td>{fecha(ev.fecha_hora)}</td>
                    <td><span className={`badge badge-${ev.categoria}`}>{ev.categoria}</span></td>
                    <td>${ev.precio_base?.toLocaleString('es-AR')}</td>
                    <td>
                      <div className="tabla-ocupacion">
                        <div className="ocupacion-barra" style={{ height: 4, width: 80 }}>
                          <div className="ocupacion-relleno" style={{ width: `${ocupacion(ev)}%` }} />
                        </div>
                        <span>{ocupacion(ev)}%</span>
                      </div>
                    </td>
                    <td>
                      <span className={`badge badge-estado-${ev.estado}`}>{ev.estado}</span>
                    </td>
                    <td>
                      <div className="tabla-acciones">
                        <button className="btn btn-secundario btn-sm" onClick={() => handleVerReporte(ev)}
                          title="Ver reporte">Rep</button>
                        <button className="btn btn-secundario btn-sm" onClick={() => abrirEditar(ev)}
                          title="Editar">Edit</button>
                        <button className="btn btn-peligro btn-sm" onClick={() => setEventoEliminando(ev)}
                          title="Eliminar">Del</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {modalForm && (
        <ModalFormEvento
          evento={eventoEditando}
          onGuardar={handleGuardar}
          onCerrar={() => { setModalForm(false); setEventoEditando(null) }}
          cargando={accionCargando}
          subiendoImagen={subiendoImagen}
          error={errorForm}
        />
      )}
      {eventoEliminando && (
        <ModalConfirmarEliminar
          evento={eventoEliminando}
          onConfirmar={handleEliminar}
          onCerrar={() => setEventoEliminando(null)}
          cargando={accionCargando}
        />
      )}
      {reporte && (
        <ModalReporte reporte={reporte} onCerrar={() => setReporte(null)} />
      )}
    </div>
  )
}
