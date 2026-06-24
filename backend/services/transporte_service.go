package services

import (
	"errors"
	"fmt"

	"ticketsya/dao"
	"ticketsya/domain"

	"gorm.io/gorm"
)

// TransporteService define el contrato del servicio de asistente de transporte
type TransporteService interface {
	ConfigurarTransporte(usuarioID uint, dto domain.DTOCrearAsistenteTransporte) (*domain.DTORespuestaAsistente, error)
	ObtenerPorEntrada(usuarioID, entradaID uint) (*domain.DTORespuestaAsistente, error)
	// ListarOfertasAuto devuelve quién ofrece compartir auto para un evento,
	// para que el usuario elija con quién viajar.
	ListarOfertasAuto(usuarioID, eventoID uint) ([]domain.AsistenteTransporte, error)
	// SolicitarCompartir crea (o reutiliza) la solicitud de un usuario para
	// unirse al auto de otro. Queda en estado "pendiente" hasta que el dueño responda.
	SolicitarCompartir(usuarioID uint, asistenteOfertaID uint) (*domain.AsistenteTransporte, error)
	// ResponderSolicitud es la acción del dueño del auto: aprobar o rechazar
	// a quien pidió compartir el viaje.
	ResponderSolicitud(duenoID uint, asistenteID uint, aprobar bool) (*domain.AsistenteTransporte, error)
}

type transporteServiceImpl struct {
	transporteDAO dao.TransporteDAO
	entradaDAO    dao.EntradaDAO
	usuarioDAO    dao.UsuarioDAO
}

// NuevoTransporteService crea una nueva instancia del servicio de transporte.
// usuarioDAO se agrega en la Parte 2 para poder validar que el usuario que
// solicita compartir existe y está activo antes de crear el match.
func NuevoTransporteService(transporteDAO dao.TransporteDAO, entradaDAO dao.EntradaDAO, usuarioDAO dao.UsuarioDAO) TransporteService {
	return &transporteServiceImpl{
		transporteDAO: transporteDAO,
		entradaDAO:    entradaDAO,
		usuarioDAO:    usuarioDAO,
	}
}

// catalogoLineasColectivo es un listado fijo en código (no requiere tabla
// propia ni gestión desde el admin) con líneas de ejemplo y sus links de
// horarios. Para la demo alcanza con unas pocas líneas representativas.
var catalogoLineasColectivo = []domain.LineaColectivoInfo{
	{Linea: "Línea 1", Recorrido: "Centro - Estadio", URLHorario: "https://www.tusubeargentina.com.ar/horarios"},
	{Linea: "Línea 7", Recorrido: "Terminal - Centro", URLHorario: "https://www.tusubeargentina.com.ar/horarios"},
	{Linea: "Línea 14", Recorrido: "Barrio Norte - Estadio", URLHorario: "https://www.tusubeargentina.com.ar/horarios"},
	{Linea: "Línea 20", Recorrido: "Circunvalación", URLHorario: "https://www.tusubeargentina.com.ar/horarios"},
}

// catalogoEstacionamientos es un listado fijo de estacionamientos de ejemplo
// para mostrar a quienes eligen ir en auto propio. En un sistema real esto
// vendría de una API de mapas; para el alcance del bonus alcanza con datos
// representativos fijos.
var catalogoEstacionamientos = []domain.EstacionamientoInfo{
	{Nombre: "Parking Centro", Direccion: "Av. Principal 450", Distancia: "200m del lugar", Latitud: -31.4201, Longitud: -64.1888},
	{Nombre: "Estacionamiento Plaza", Direccion: "San Martín 120", Distancia: "350m del lugar", Latitud: -31.4180, Longitud: -64.1850},
	{Nombre: "Parking Estadio", Direccion: "Av. del Estadio 10", Distancia: "100m del lugar", Latitud: -31.4220, Longitud: -64.1900},
}

// ConfigurarTransporte crea (o reemplaza) la preferencia de transporte de una
// entrada. Valida que la entrada exista, pertenezca al usuario y esté activa
// — no tiene sentido planear el transporte de un ticket cancelado.
func (s *transporteServiceImpl) ConfigurarTransporte(usuarioID uint, dto domain.DTOCrearAsistenteTransporte) (*domain.DTORespuestaAsistente, error) {
	// 1. Verificar que la entrada existe y pertenece al usuario
	entrada, err := s.entradaDAO.BuscarPorID(dto.EntradaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("entrada no encontrada")
		}
		return nil, fmt.Errorf("error al verificar entrada: %w", err)
	}

	if entrada.UsuarioID != usuarioID {
		return nil, errors.New("no tenés permisos para configurar el transporte de esta entrada")
	}

	if entrada.Estado != domain.EstadoEntradaActiva {
		return nil, fmt.Errorf("solo se puede configurar transporte para entradas activas (estado: %s)", entrada.Estado)
	}

	// 2. Validar el modo y los campos específicos de cada modo
	if dto.Modo == domain.ModoColectivo && dto.LineaColectivo == "" {
		return nil, errors.New("debés indicar la línea de colectivo elegida")
	}

	// 3. Si ya existe una configuración previa para esta entrada, la actualizamos
	// en lugar de crear un duplicado (entrada_id tiene uniqueIndex en la tabla)
	existente, err := s.transporteDAO.BuscarPorEntradaID(dto.EntradaID)
	var asistente *domain.AsistenteTransporte

	if err == nil && existente != nil {
		existente.Modo = dto.Modo
		existente.LineaColectivo = dto.LineaColectivo
		existente.ComparteAuto = dto.ComparteAuto
		if err := s.transporteDAO.Actualizar(existente); err != nil {
			return nil, fmt.Errorf("error al actualizar configuración de transporte: %w", err)
		}
		asistente = existente
	} else {
		asistente = &domain.AsistenteTransporte{
			EntradaID:      dto.EntradaID,
			UsuarioID:      usuarioID,
			EventoID:       entrada.EventoID,
			Modo:           dto.Modo,
			LineaColectivo: dto.LineaColectivo,
			ComparteAuto:   dto.ComparteAuto,
		}
		if err := s.transporteDAO.Crear(asistente); err != nil {
			return nil, fmt.Errorf("error al crear configuración de transporte: %w", err)
		}
	}

	return s.armarRespuesta(asistente), nil
}

// ObtenerPorEntrada recupera la configuración de transporte de una entrada,
// si existe. Devuelve también el catálogo de apoyo (líneas o estacionamientos)
// según el modo elegido, para que el frontend tenga todo en una sola llamada.
func (s *transporteServiceImpl) ObtenerPorEntrada(usuarioID, entradaID uint) (*domain.DTORespuestaAsistente, error) {
	entrada, err := s.entradaDAO.BuscarPorID(entradaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("entrada no encontrada")
		}
		return nil, err
	}
	if entrada.UsuarioID != usuarioID {
		return nil, errors.New("no tenés permisos para ver el transporte de esta entrada")
	}

	asistente, err := s.transporteDAO.BuscarPorEntradaID(entradaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Todavía no configuró nada — no es un error, simplemente no hay datos aún
			return &domain.DTORespuestaAsistente{Asistente: nil}, nil
		}
		return nil, err
	}

	return s.armarRespuesta(asistente), nil
}

// ListarOfertasAuto devuelve las ofertas de auto compartido disponibles para
// un evento, para que el usuario elija con quién viajar. Excluye sus propias
// ofertas (no tiene sentido que se vea a sí mismo en la lista).
func (s *transporteServiceImpl) ListarOfertasAuto(usuarioID, eventoID uint) ([]domain.AsistenteTransporte, error) {
	ofertas, err := s.transporteDAO.ListarComparteAutoPorEvento(eventoID, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error al listar ofertas de auto compartido: %w", err)
	}
	return ofertas, nil
}

// SolicitarCompartir registra que un usuario quiere unirse al auto de otro.
// asistenteOfertaID es el ID del registro AsistenteTransporte del DUEÑO del
// auto (no del solicitante) — es el que aparece en ListarOfertasAuto.
//
// Reglas de negocio:
//   - El registro de la oferta debe existir, estar en modo auto_propio y
//     tener comparte_auto = true
//   - No se puede solicitar el propio auto
//   - Si ya hay una solicitud pendiente o aprobada de OTRO usuario, no se
//     puede pedir de nuevo (el auto ya tiene acompañante o está en revisión)
//   - Queda en estado "pendiente" hasta que el dueño responda
func (s *transporteServiceImpl) SolicitarCompartir(usuarioID uint, asistenteOfertaID uint) (*domain.AsistenteTransporte, error) {
	oferta, err := s.transporteDAO.BuscarPorID(asistenteOfertaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("oferta de auto compartido no encontrada")
		}
		return nil, err
	}

	if oferta.Modo != domain.ModoAutoPropio || !oferta.ComparteAuto {
		return nil, errors.New("este registro no es una oferta de auto compartido")
	}

	if oferta.UsuarioID == usuarioID {
		return nil, errors.New("no podés solicitar tu propio auto")
	}

	if oferta.EstadoMatch != nil &&
		(*oferta.EstadoMatch == domain.EstadoMatchPendiente || *oferta.EstadoMatch == domain.EstadoMatchAprobado) {
		return nil, errors.New("esta oferta ya tiene una solicitud en curso")
	}

	// Verificar que el usuario solicitante exista (defensivo; el JWT ya lo
	// garantiza en la práctica, pero mantiene la capa de servicio autocontenida)
	if _, err := s.usuarioDAO.BuscarPorID(usuarioID); err != nil {
		return nil, errors.New("usuario solicitante no encontrado")
	}

	estadoPendiente := domain.EstadoMatchPendiente
	oferta.EstadoMatch = &estadoPendiente
	oferta.UsuarioMatchID = &usuarioID

	if err := s.transporteDAO.Actualizar(oferta); err != nil {
		return nil, fmt.Errorf("error al registrar la solicitud: %w", err)
	}

	// Recargamos con los preloads para devolver los datos de ambos usuarios
	actualizado, err := s.transporteDAO.BuscarPorID(oferta.ID)
	if err != nil {
		return oferta, nil // el cambio ya se guardó; si falla el reload no es crítico
	}
	return actualizado, nil
}

// ResponderSolicitud es la acción que ejecuta el DUEÑO del auto para aprobar
// o rechazar a quien pidió compartir viaje. Solo el dueño (asistente.UsuarioID)
// puede responder su propia oferta.
//
// Al aprobar, ambos usuarios quedan habilitados para ver los datos de
// contacto del otro (eso se resuelve en el controller/frontend leyendo
// Usuario y UsuarioMatch una vez que EstadoMatch == aprobado).
// Al rechazar, se libera la oferta (UsuarioMatchID vuelve a nil) para que
// otro usuario pueda solicitarla.
func (s *transporteServiceImpl) ResponderSolicitud(duenoID uint, asistenteID uint, aprobar bool) (*domain.AsistenteTransporte, error) {
	asistente, err := s.transporteDAO.BuscarPorID(asistenteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("configuración de transporte no encontrada")
		}
		return nil, err
	}

	if asistente.UsuarioID != duenoID {
		return nil, errors.New("no tenés permisos para responder esta solicitud")
	}

	if asistente.EstadoMatch == nil || *asistente.EstadoMatch != domain.EstadoMatchPendiente {
		return nil, errors.New("no hay una solicitud pendiente para responder")
	}

	if aprobar {
		estadoAprobado := domain.EstadoMatchAprobado
		asistente.EstadoMatch = &estadoAprobado
	} else {
		estadoRechazado := domain.EstadoMatchRechazado
		asistente.EstadoMatch = &estadoRechazado
		// Liberamos el cupo: si alguien más quiere solicitar este auto después,
		// que pueda hacerlo. UsuarioMatchID se mantiene para historial del
		// rechazo, pero el chequeo de SolicitarCompartir solo bloquea en
		// pendiente/aprobado, así que un rechazo no impide nuevas solicitudes.
	}

	if err := s.transporteDAO.Actualizar(asistente); err != nil {
		return nil, fmt.Errorf("error al responder la solicitud: %w", err)
	}

	actualizado, err := s.transporteDAO.BuscarPorID(asistente.ID)
	if err != nil {
		return asistente, nil
	}
	return actualizado, nil
}

// armarRespuesta adjunta el catálogo de apoyo correspondiente al modo elegido
func (s *transporteServiceImpl) armarRespuesta(asistente *domain.AsistenteTransporte) *domain.DTORespuestaAsistente {
	respuesta := &domain.DTORespuestaAsistente{Asistente: asistente}

	switch asistente.Modo {
	case domain.ModoColectivo:
		respuesta.LineasColectivo = catalogoLineasColectivo
	case domain.ModoAutoPropio, domain.ModoCompartido:
		respuesta.Estacionamientos = catalogoEstacionamientos
	}

	return respuesta
}