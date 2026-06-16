package dao

import (
	"time"

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

type eventoDAOImpl struct {
	db *gorm.DB
}

func NuevoEventoDAO(db *gorm.DB) EventoDAO {
	return &eventoDAOImpl{db: db}
}

func (dao *eventoDAOImpl) Crear(evento *domain.Evento) error {
	return dao.db.Create(evento).Error
}

func (dao *eventoDAOImpl) BuscarPorID(id uint) (*domain.Evento, error) {
	var evento domain.Evento
	resultado := dao.db.First(&evento, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &evento, nil
}

// ListarConFiltros recupera eventos aplicando los filtros recibidos.
// IMPORTANTE: solo devuelve eventos con fecha_hora >= ahora (no vencidos)
func (dao *eventoDAOImpl) ListarConFiltros(filtros domain.FiltrosEvento) ([]domain.Evento, error) {
	var eventos []domain.Evento
	query := dao.db.Model(&domain.Evento{})

	// ── Solo eventos futuros (no mostrar eventos vencidos en el catálogo) ──────
	query = query.Where("fecha_hora >= ?", time.Now())

	// Filtro de texto
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

	// Solo eventos con entradas disponibles
	if filtros.SoloDisponibles {
		query = query.Where("estado = ? AND entradas_vendidas < capacidad_total",
			domain.EstadoActivo)
	}

	resultado := query.Order("fecha_hora ASC").Find(&eventos)
	return eventos, resultado.Error
}

func (dao *eventoDAOImpl) Actualizar(evento *domain.Evento) error {
	return dao.db.Save(evento).Error
}

func (dao *eventoDAOImpl) Eliminar(id uint) error {
	return dao.db.Delete(&domain.Evento{}, id).Error
}

func (dao *eventoDAOImpl) IncrementarVentas(eventoID uint) error {
	return dao.db.Model(&domain.Evento{}).
		Where("id = ?", eventoID).
		UpdateColumn("entradas_vendidas", gorm.Expr("entradas_vendidas + 1")).Error
}

func (dao *eventoDAOImpl) DecrementarVentas(eventoID uint) error {
	return dao.db.Model(&domain.Evento{}).
		Where("id = ? AND entradas_vendidas > 0", eventoID).
		UpdateColumn("entradas_vendidas", gorm.Expr("entradas_vendidas - 1")).Error
}

func (dao *eventoDAOImpl) ReportePorID(id uint) (*domain.Evento, error) {
	var evento domain.Evento
	resultado := dao.db.Preload("Entradas").Preload("Entradas.Usuario").First(&evento, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &evento, nil
}
