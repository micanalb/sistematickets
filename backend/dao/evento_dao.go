package dao

import (
	"ticketsya/domain"

	"gorm.io/gorm"
)

// EventoDAO define el contrato de acceso a datos para la entidad Evento
type EventoDAO interface {
	Crear(evento *domain.Evento) error
	BuscarPorID(id uint) (*domain.Evento, error)
	ListarConFiltros(filtros domain.FiltrosEvento) ([]domain.Evento, error)
	Actualizar(evento *domain.Evento) error
	Eliminar(id uint) error
	IncrementarVentas(eventoID uint) error
	DecrementarVentas(eventoID uint) error
	ReportePorID(id uint) (*domain.Evento, error)
}

// eventoDAOImpl es la implementación concreta con GORM
type eventoDAOImpl struct {
	db *gorm.DB
}

// NuevoEventoDAO crea una nueva instancia del DAO de evento
func NuevoEventoDAO(db *gorm.DB) EventoDAO {
	return &eventoDAOImpl{db: db}
}

// Crear inserta un nuevo evento
func (dao *eventoDAOImpl) Crear(evento *domain.Evento) error {
	return dao.db.Create(evento).Error
}

// BuscarPorID recupera un evento con sus datos completos
func (dao *eventoDAOImpl) BuscarPorID(id uint) (*domain.Evento, error) {
	var evento domain.Evento
	resultado := dao.db.First(&evento, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &evento, nil
}

// ListarConFiltros recupera el catálogo de eventos aplicando los filtros especificados.
// Solo retorna eventos con soft-delete = nil (no eliminados).
func (dao *eventoDAOImpl) ListarConFiltros(filtros domain.FiltrosEvento) ([]domain.Evento, error) {
	var eventos []domain.Evento
	query := dao.db.Model(&domain.Evento{})

	// Filtro de texto: busca en título, descripción y lugar
	if filtros.Busqueda != "" {
		termino := "%" + filtros.Busqueda + "%"
		query = query.Where(
			"titulo LIKE ? OR descripcion LIKE ? OR lugar LIKE ?",
			termino, termino, termino,
		)
	}

	// Filtro por categoría
	if filtros.Categoria != "" {
		query = query.Where("categoria = ?", filtros.Categoria)
	}

	// Filtro por ciudad
	if filtros.Ciudad != "" {
		query = query.Where("ciudad LIKE ?", "%"+filtros.Ciudad+"%")
	}

	// Filtro por rango de fechas
	if filtros.FechaDesde != nil {
		query = query.Where("fecha_hora >= ?", filtros.FechaDesde)
	}
	if filtros.FechaHasta != nil {
		query = query.Where("fecha_hora <= ?", filtros.FechaHasta)
	}

	// Filtro de solo eventos con entradas disponibles
	if filtros.SoloDisponibles {
		query = query.Where("estado = ? AND entradas_vendidas < capacidad_total",
			domain.EstadoActivo)
	}

	// Ordenamos por fecha de forma ascendente (próximos primero)
	resultado := query.Order("fecha_hora ASC").Find(&eventos)
	return eventos, resultado.Error
}

// Actualizar persiste los cambios de un evento
func (dao *eventoDAOImpl) Actualizar(evento *domain.Evento) error {
	return dao.db.Save(evento).Error
}

// Eliminar realiza un soft-delete del evento (GORM lo maneja automáticamente
// cuando el modelo tiene el campo DeletedAt de tipo gorm.DeletedAt)
func (dao *eventoDAOImpl) Eliminar(id uint) error {
	return dao.db.Delete(&domain.Evento{}, id).Error
}

// IncrementarVentas aumenta atómicamente el contador de entradas vendidas
func (dao *eventoDAOImpl) IncrementarVentas(eventoID uint) error {
	return dao.db.Model(&domain.Evento{}).
		Where("id = ?", eventoID).
		UpdateColumn("entradas_vendidas", gorm.Expr("entradas_vendidas + 1")).Error
}

// DecrementarVentas reduce atómicamente el contador de entradas vendidas (al cancelar)
func (dao *eventoDAOImpl) DecrementarVentas(eventoID uint) error {
	return dao.db.Model(&domain.Evento{}).
		Where("id = ? AND entradas_vendidas > 0", eventoID).
		UpdateColumn("entradas_vendidas", gorm.Expr("entradas_vendidas - 1")).Error
}

// ReportePorID recupera el evento con todas sus entradas para generar el reporte de ocupación
func (dao *eventoDAOImpl) ReportePorID(id uint) (*domain.Evento, error) {
	var evento domain.Evento
	resultado := dao.db.Preload("Entradas").Preload("Entradas.Usuario").First(&evento, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &evento, nil
}
