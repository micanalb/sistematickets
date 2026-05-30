package services

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"ticketsya/dao"
	"ticketsya/domain"
)

// EventoService define el contrato del servicio de eventos
type EventoService interface {
	ListarEventos(filtros domain.FiltrosEvento) ([]domain.Evento, error)
	ObtenerEventoPorID(id uint) (*domain.Evento, error)
	CrearEvento(dto domain.DTOCrearEvento) (*domain.Evento, error)
	ActualizarEvento(id uint, dto domain.DTOActualizarEvento) (*domain.Evento, error)
	EliminarEvento(id uint) error
	ObtenerReporte(id uint) (*domain.Evento, error)
}

// eventoServiceImpl es la implementación concreta
type eventoServiceImpl struct {
	eventoDAO dao.EventoDAO
}

// NuevoEventoService crea una nueva instancia del servicio de eventos
func NuevoEventoService(eventoDAO dao.EventoDAO) EventoService {
	return &eventoServiceImpl{eventoDAO: eventoDAO}
}

// ListarEventos retorna el catálogo de eventos aplicando los filtros recibidos.
// Este endpoint es público (no requiere autenticación).
func (s *eventoServiceImpl) ListarEventos(filtros domain.FiltrosEvento) ([]domain.Evento, error) {
	eventos, err := s.eventoDAO.ListarConFiltros(filtros)
	if err != nil {
		return nil, fmt.Errorf("error al listar eventos: %w", err)
	}
	return eventos, nil
}

// ObtenerEventoPorID retorna el detalle completo de un evento específico.
// Este endpoint es público (no requiere autenticación).
func (s *eventoServiceImpl) ObtenerEventoPorID(id uint) (*domain.Evento, error) {
	evento, err := s.eventoDAO.BuscarPorID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("evento no encontrado")
		}
		return nil, fmt.Errorf("error al obtener evento: %w", err)
	}
	return evento, nil
}

// CrearEvento registra un nuevo evento en el catálogo (solo administradores).
func (s *eventoServiceImpl) CrearEvento(dto domain.DTOCrearEvento) (*domain.Evento, error) {
	// Validación de negocio: la capacidad debe ser positiva
	if dto.CapacidadTotal <= 0 {
		return nil, errors.New("la capacidad del evento debe ser mayor a cero")
	}

	// Validación de negocio: precio no puede ser negativo
	if dto.PrecioBase < 0 {
		return nil, errors.New("el precio base no puede ser negativo")
	}

	evento := &domain.Evento{
		Titulo:          dto.Titulo,
		Descripcion:     dto.Descripcion,
		FechaHora:       dto.FechaHora,
		DuracionMinutos: dto.DuracionMinutos,
		Lugar:           dto.Lugar,
		Direccion:       dto.Direccion,
		Ciudad:          dto.Ciudad,
		Categoria:       dto.Categoria,
		CapacidadTotal:  dto.CapacidadTotal,
		PrecioBase:      dto.PrecioBase,
		ImagenURL:       dto.ImagenURL,
		Estado:          domain.EstadoActivo,
	}

	if err := s.eventoDAO.Crear(evento); err != nil {
		return nil, fmt.Errorf("error al crear evento: %w", err)
	}

	return evento, nil
}

// ActualizarEvento modifica los atributos de un evento existente.
// Solo se actualizan los campos que vienen en el DTO (actualización parcial).
func (s *eventoServiceImpl) ActualizarEvento(id uint, dto domain.DTOActualizarEvento) (*domain.Evento, error) {
	evento, err := s.eventoDAO.BuscarPorID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("evento no encontrado")
		}
		return nil, err
	}

	// Actualización selectiva: solo modificamos los campos que no son nil
	if dto.Titulo != nil {
		evento.Titulo = *dto.Titulo
	}
	if dto.Descripcion != nil {
		evento.Descripcion = *dto.Descripcion
	}
	if dto.FechaHora != nil {
		evento.FechaHora = *dto.FechaHora
	}
	if dto.DuracionMinutos != nil {
		evento.DuracionMinutos = *dto.DuracionMinutos
	}
	if dto.Lugar != nil {
		evento.Lugar = *dto.Lugar
	}
	if dto.Direccion != nil {
		evento.Direccion = *dto.Direccion
	}
	if dto.Ciudad != nil {
		evento.Ciudad = *dto.Ciudad
	}
	if dto.Categoria != nil {
		evento.Categoria = *dto.Categoria
	}
	if dto.CapacidadTotal != nil {
		// Validar que la nueva capacidad no sea menor a las ventas actuales
		if *dto.CapacidadTotal < evento.EntradasVendidas {
			return nil, errors.New("la nueva capacidad no puede ser menor a las entradas ya vendidas")
		}
		evento.CapacidadTotal = *dto.CapacidadTotal
	}
	if dto.PrecioBase != nil {
		evento.PrecioBase = *dto.PrecioBase
	}
	if dto.ImagenURL != nil {
		evento.ImagenURL = *dto.ImagenURL
	}
	if dto.Estado != nil {
		evento.Estado = *dto.Estado
	}

	if err := s.eventoDAO.Actualizar(evento); err != nil {
		return nil, fmt.Errorf("error al actualizar evento: %w", err)
	}

	return evento, nil
}

// EliminarEvento realiza un soft-delete del evento.
// Decisión de diseño: usamos soft-delete para mantener el historial de entradas
// compradas, ya que eliminar el evento no debe borrar los tickets de los usuarios.
func (s *eventoServiceImpl) EliminarEvento(id uint) error {
	_, err := s.eventoDAO.BuscarPorID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("evento no encontrado")
		}
		return err
	}

	return s.eventoDAO.Eliminar(id)
}

// ObtenerReporte retorna el evento con todas sus entradas para el reporte de ocupación
func (s *eventoServiceImpl) ObtenerReporte(id uint) (*domain.Evento, error) {
	evento, err := s.eventoDAO.ReportePorID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("evento no encontrado")
		}
		return nil, fmt.Errorf("error al obtener reporte: %w", err)
	}
	return evento, nil
}
