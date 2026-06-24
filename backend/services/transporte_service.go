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
}

type transporteServiceImpl struct {
	transporteDAO dao.TransporteDAO
	entradaDAO    dao.EntradaDAO
}

// NuevoTransporteService crea una nueva instancia del servicio de transporte
func NuevoTransporteService(transporteDAO dao.TransporteDAO, entradaDAO dao.EntradaDAO) TransporteService {
	return &transporteServiceImpl{
		transporteDAO: transporteDAO,
		entradaDAO:    entradaDAO,
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
