package services

import (
	"errors"
	"fmt"
	"time"

	"ticketsya/dao"
	"ticketsya/domain"
	"ticketsya/utils"

	"gorm.io/gorm"
)

// EntradaService define el contrato del servicio de entradas
type EntradaService interface {
	ComprarEntrada(usuarioID uint, dto domain.DTOComprarEntrada) ([]domain.Entrada, error)
	MisEntradas(usuarioID uint) ([]domain.Entrada, error)
	CancelarEntrada(entradaID, usuarioID uint) error
	TransferirEntrada(entradaID, usuarioID uint, dto domain.DTOTransferirEntrada) (*domain.Entrada, error)
}

// entradaServiceImpl es la implementación concreta
type entradaServiceImpl struct {
	db         *gorm.DB
	entradaDAO dao.EntradaDAO
	eventoDAO  dao.EventoDAO
	usuarioDAO dao.UsuarioDAO
}

// NuevoEntradaService crea una nueva instancia del servicio de entradas.
// Recibe la conexión *gorm.DB para poder envolver operaciones multi-paso
// (compra, cancelación, transferencia) en transacciones atómicas.
func NuevoEntradaService(db *gorm.DB, entradaDAO dao.EntradaDAO, eventoDAO dao.EventoDAO, usuarioDAO dao.UsuarioDAO) EntradaService {
	return &entradaServiceImpl{
		db:         db,
		entradaDAO: entradaDAO,
		eventoDAO:  eventoDAO,
		usuarioDAO: usuarioDAO,
	}
}

// ComprarEntrada procesa la adquisición de una o más entradas para un evento.
// dto.Cantidad indica cuántos tickets comprar en la misma operación (mínimo 1,
// máximo 10 por compra). Si no se especifica, se normaliza a 1 para mantener
// compatible el comportamiento de compra individual.
//
// Las validaciones (evento existe, está activo, hay disponibilidad para TODA
// la cantidad solicitada) se hacen ANTES de abrir la transacción para fallar
// rápido sin tomar locks innecesarios. La escritura (crear las N entradas +
// incrementar ventas en N) se ejecuta dentro de db.Transaction: si cualquier
// paso falla, GORM hace ROLLBACK automático y ninguna de las entradas queda
// guardada — no puede pasar que se cobren 3 y se acrediten 2.
func (s *entradaServiceImpl) ComprarEntrada(usuarioID uint, dto domain.DTOComprarEntrada) ([]domain.Entrada, error) {
	// 0. Normalizar cantidad: si no vino o vino en 0, es una compra de 1 entrada
	cantidad := dto.Cantidad
	if cantidad <= 0 {
		cantidad = 1
	}

	// 1. Verificar que el evento existe y está activo
	evento, err := s.eventoDAO.BuscarPorID(dto.EventoID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, fmt.Errorf("error al verificar evento: %w", err)
	}

	// 2. Verificar que el evento está activo (no cancelado ni finalizado)
	if evento.Estado != domain.EstadoActivo {
		return nil, fmt.Errorf("el evento no está disponible para compra (estado: %s)", evento.Estado)
	}

	// 3. Verificar que hay disponibilidad suficiente para TODA la cantidad pedida.
	//    No alcanza con TieneDisponibilidad() (que solo chequea >0); si alguien
	//    pide 5 y quedan 3, debe rechazarse la operación completa.
	disponibles := evento.DisponibilidadEntradas()
	if disponibles <= 0 {
		return nil, errors.New("no hay entradas disponibles para este evento")
	}
	if cantidad > disponibles {
		return nil, fmt.Errorf("solo quedan %d entradas disponibles, no se pueden comprar %d", disponibles, cantidad)
	}

	// 4. Construir las N entradas en memoria, cada una con su propio código QR único
	entradas := make([]domain.Entrada, cantidad)
	for i := 0; i < cantidad; i++ {
		entradas[i] = domain.Entrada{
			CodigoQR:     utils.GenerarCodigoQR(evento.ID, usuarioID),
			UsuarioID:    usuarioID,
			EventoID:     dto.EventoID,
			PrecioPagado: evento.PrecioBase,
			Estado:       domain.EstadoEntradaActiva,
			FechaCompra:  time.Now(),
		}
	}

	// 5. Crear las N entradas + incrementar ventas en N, en una única transacción
	//    atómica. Si falla cualquier paso a mitad de camino, GORM revierte todo
	//    lo creado hasta ese punto — no quedan entradas "fantasma" sin contabilizar.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// gorm.Create con un slice inserta todas las filas en una sola operación
		if err := tx.Create(&entradas).Error; err != nil {
			return fmt.Errorf("error al crear entradas: %w", err)
		}

		if err := tx.Model(&domain.Evento{}).
			Where("id = ?", dto.EventoID).
			UpdateColumn("entradas_vendidas", gorm.Expr("entradas_vendidas + ?", cantidad)).Error; err != nil {
			return fmt.Errorf("error al actualizar disponibilidad: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 6. Cargar los datos del evento en cada entrada de la respuesta
	for i := range entradas {
		entradas[i].Evento = evento
	}
	return entradas, nil
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

	// 4. Marcar como cancelada + devolver el cupo, en una sola transacción.
	//    Si falla la liberación del cupo, la cancelación de la entrada también
	//    se revierte — evita que quede "cancelada" sin liberar el cupo del evento.
	ahora := time.Now()
	eventoID := entrada.EventoID

	err = s.db.Transaction(func(tx *gorm.DB) error {
		entrada.Estado = domain.EstadoEntradaCancelada
		entrada.FechaCancelacion = &ahora

		if err := tx.Save(entrada).Error; err != nil {
			return fmt.Errorf("error al cancelar entrada: %w", err)
		}

		if err := tx.Model(&domain.Evento{}).
			Where("id = ? AND entradas_vendidas > 0", eventoID).
			UpdateColumn("entradas_vendidas", gorm.Expr("entradas_vendidas - 1")).Error; err != nil {
			return fmt.Errorf("error al liberar cupo del evento: %w", err)
		}

		return nil
	})

	return err
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
			return nil, errors.New("record not found")
		}
		return nil, err
	}

	// 5. No puede transferirse a sí mismo
	if destinatario.ID == usuarioID {
		return nil, errors.New("no podés transferir una entrada a vos mismo")
	}

	// 6. Marcar la original como transferida + crear la nueva entrada del
	//    destinatario en una sola transacción. Si falla la creación de la
	//    nueva entrada, la original NO queda marcada como transferida sin
	//    que exista una entrada válida del otro lado — se revierte todo.
	nuevaEntrada := &domain.Entrada{
		CodigoQR:     utils.GenerarCodigoQR(entrada.EventoID, destinatario.ID),
		UsuarioID:    destinatario.ID,
		EventoID:     entrada.EventoID,
		PrecioPagado: entrada.PrecioPagado,
		Estado:       domain.EstadoEntradaActiva,
		FechaCompra:  time.Now(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		entrada.Estado = domain.EstadoEntradaTransferida
		if err := tx.Save(entrada).Error; err != nil {
			return fmt.Errorf("error al marcar entrada como transferida: %w", err)
		}

		if err := tx.Create(nuevaEntrada).Error; err != nil {
			return fmt.Errorf("error al crear entrada para destinatario: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
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
