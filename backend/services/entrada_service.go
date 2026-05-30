package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"ticketsya/dao"
	"ticketsya/domain"
	"ticketsya/utils"
)

// EntradaService define el contrato del servicio de entradas
type EntradaService interface {
	ComprarEntrada(usuarioID uint, dto domain.DTOComprarEntrada) (*domain.Entrada, error)
	MisEntradas(usuarioID uint) ([]domain.Entrada, error)
	CancelarEntrada(entradaID, usuarioID uint) error
	TransferirEntrada(entradaID, usuarioID uint, dto domain.DTOTransferirEntrada) (*domain.Entrada, error)
}

// entradaServiceImpl es la implementación concreta
type entradaServiceImpl struct {
	entradaDAO dao.EntradaDAO
	eventoDAO  dao.EventoDAO
	usuarioDAO dao.UsuarioDAO
}

// NuevoEntradaService crea una nueva instancia del servicio de entradas
func NuevoEntradaService(entradaDAO dao.EntradaDAO, eventoDAO dao.EventoDAO, usuarioDAO dao.UsuarioDAO) EntradaService {
	return &entradaServiceImpl{
		entradaDAO: entradaDAO,
		eventoDAO:  eventoDAO,
		usuarioDAO: usuarioDAO,
	}
}

// ComprarEntrada procesa la adquisición de una entrada para un evento.
// Incluye todas las validaciones de negocio necesarias.
func (s *entradaServiceImpl) ComprarEntrada(usuarioID uint, dto domain.DTOComprarEntrada) (*domain.Entrada, error) {
	// 1. Verificar que el evento existe y está activo
	evento, err := s.eventoDAO.BuscarPorID(dto.EventoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el evento no existe")
		}
		return nil, fmt.Errorf("error al verificar evento: %w", err)
	}

	// 2. Verificar que el evento está activo (no cancelado ni finalizado)
	if evento.Estado != domain.EstadoActivo {
		return nil, fmt.Errorf("el evento no está disponible para compra (estado: %s)", evento.Estado)
	}

	// 3. Verificar disponibilidad de entradas
	if !evento.TieneDisponibilidad() {
		return nil, errors.New("no hay entradas disponibles para este evento")
	}

	// 4. Crear la entrada con código QR único
	codigoQR := utils.GenerarCodigoQR(evento.ID, usuarioID)
	entrada := &domain.Entrada{
		CodigoQR:     codigoQR,
		UsuarioID:    usuarioID,
		EventoID:     dto.EventoID,
		PrecioPagado: evento.PrecioBase,
		Estado:       domain.EstadoEntradaActiva,
		FechaCompra:  time.Now(),
	}

	if err := s.entradaDAO.Crear(entrada); err != nil {
		return nil, fmt.Errorf("error al crear entrada: %w", err)
	}

	// 5. Incrementar el contador de ventas del evento (operación atómica)
	if err := s.eventoDAO.IncrementarVentas(dto.EventoID); err != nil {
		// En un sistema real aquí haría rollback de la transacción
		return nil, fmt.Errorf("error al actualizar disponibilidad: %w", err)
	}

	// 6. Cargar los datos del evento en la respuesta
	entrada.Evento = evento
	return entrada, nil
}

// MisEntradas retorna el historial de entradas del usuario autenticado
func (s *entradaServiceImpl) MisEntradas(usuarioID uint) ([]domain.Entrada, error) {
	entradas, err := s.entradaDAO.ListarPorUsuario(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener entradas: %w", err)
	}
	return entradas, nil
}

// CancelarEntrada procesa la anulación de un ticket.
// Solo el propietario puede cancelar su entrada y esta debe estar activa.
func (s *entradaServiceImpl) CancelarEntrada(entradaID, usuarioID uint) error {
	// 1. Obtener la entrada con validación de existencia
	entrada, err := s.entradaDAO.BuscarPorID(entradaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("entrada no encontrada")
		}
		return err
	}

	// 2. Verificar que la entrada pertenece al usuario que solicita la cancelación
	if entrada.UsuarioID != usuarioID {
		return errors.New("no tenés permisos para cancelar esta entrada")
	}

	// 3. Verificar que la entrada puede ser cancelada (solo activas)
	if entrada.Estado != domain.EstadoEntradaActiva {
		return fmt.Errorf("la entrada no puede ser cancelada (estado: %s)", entrada.Estado)
	}

	// 4. Marcar como cancelada
	ahora := time.Now()
	entrada.Estado = domain.EstadoEntradaCancelada
	entrada.FechaCancelacion = &ahora

	if err := s.entradaDAO.Actualizar(entrada); err != nil {
		return fmt.Errorf("error al cancelar entrada: %w", err)
	}

	// 5. Devolver el cupo al evento (el stock se libera)
	if err := s.eventoDAO.DecrementarVentas(entrada.EventoID); err != nil {
		return fmt.Errorf("error al liberar cupo del evento: %w", err)
	}

	return nil
}

// TransferirEntrada transfiere la titularidad de un ticket a otro usuario registrado.
// La entrada original queda con estado "transferida" y se crea una nueva para el destinatario.
func (s *entradaServiceImpl) TransferirEntrada(entradaID, usuarioID uint, dto domain.DTOTransferirEntrada) (*domain.Entrada, error) {
	// 1. Obtener la entrada
	entrada, err := s.entradaDAO.BuscarPorID(entradaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("entrada no encontrada")
		}
		return nil, err
	}

	// 2. Verificar propiedad
	if entrada.UsuarioID != usuarioID {
		return nil, errors.New("no tenés permisos para transferir esta entrada")
	}

	// 3. Verificar estado de la entrada
	if entrada.Estado != domain.EstadoEntradaActiva {
		return nil, fmt.Errorf("solo se pueden transferir entradas activas (estado: %s)", entrada.Estado)
	}

	// 4. Verificar que el destinatario existe en el sistema
	destinatario, err := s.usuarioDAO.BuscarPorEmail(dto.EmailDestinatario)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el email destinatario no está registrado en el sistema")
		}
		return nil, err
	}

	// 5. No puede transferirse a sí mismo
	if destinatario.ID == usuarioID {
		return nil, errors.New("no podés transferir una entrada a vos mismo")
	}

	// 6. Marcar la entrada original como transferida
	entrada.Estado = domain.EstadoEntradaTransferida
	if err := s.entradaDAO.Actualizar(entrada); err != nil {
		return nil, fmt.Errorf("error al marcar entrada como transferida: %w", err)
	}

	// 7. Crear nueva entrada para el destinatario con nuevo código QR
	nuevaEntrada := &domain.Entrada{
		CodigoQR:     utils.GenerarCodigoQR(entrada.EventoID, destinatario.ID),
		UsuarioID:    destinatario.ID,
		EventoID:     entrada.EventoID,
		PrecioPagado: entrada.PrecioPagado,
		Estado:       domain.EstadoEntradaActiva,
		FechaCompra:  time.Now(),
	}

	if err := s.entradaDAO.Crear(nuevaEntrada); err != nil {
		return nil, fmt.Errorf("error al crear entrada para destinatario: %w", err)
	}

	nuevaEntrada.Usuario = &domain.Usuario{
		ID:       destinatario.ID,
		Nombre:   destinatario.Nombre,
		Apellido: destinatario.Apellido,
		Email:    destinatario.Email,
	}
	nuevaEntrada.Evento = entrada.Evento

	return nuevaEntrada, nil
}
