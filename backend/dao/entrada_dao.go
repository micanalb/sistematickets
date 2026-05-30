package dao

import (
	"ticketsya/domain"

	"gorm.io/gorm"
)

// EntradaDAO define el contrato de acceso a datos para la entidad Entrada
type EntradaDAO interface {
	Crear(entrada *domain.Entrada) error
	BuscarPorID(id uint) (*domain.Entrada, error)
	BuscarPorCodigoQR(codigoQR string) (*domain.Entrada, error)
	ListarPorUsuario(usuarioID uint) ([]domain.Entrada, error)
	Actualizar(entrada *domain.Entrada) error
	ContarActivas(eventoID, usuarioID uint) (int64, error)
}

// entradaDAOImpl es la implementación concreta con GORM
type entradaDAOImpl struct {
	db *gorm.DB
}

// NuevoEntradaDAO crea una nueva instancia del DAO de entrada
func NuevoEntradaDAO(db *gorm.DB) EntradaDAO {
	return &entradaDAOImpl{db: db}
}

// Crear inserta una nueva entrada (ticket) en la base de datos
func (dao *entradaDAOImpl) Crear(entrada *domain.Entrada) error {
	return dao.db.Create(entrada).Error
}

// BuscarPorID recupera una entrada con sus relaciones (evento y usuario)
func (dao *entradaDAOImpl) BuscarPorID(id uint) (*domain.Entrada, error) {
	var entrada domain.Entrada
	resultado := dao.db.
		Preload("Evento").
		Preload("Usuario").
		First(&entrada, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &entrada, nil
}

// BuscarPorCodigoQR recupera una entrada por su código QR único
func (dao *entradaDAOImpl) BuscarPorCodigoQR(codigoQR string) (*domain.Entrada, error) {
	var entrada domain.Entrada
	resultado := dao.db.
		Preload("Evento").
		Preload("Usuario").
		Where("codigo_qr = ?", codigoQR).
		First(&entrada)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &entrada, nil
}

// ListarPorUsuario recupera todas las entradas de un usuario específico,
// con datos del evento cargados mediante Preload para el historial
func (dao *entradaDAOImpl) ListarPorUsuario(usuarioID uint) ([]domain.Entrada, error) {
	var entradas []domain.Entrada
	resultado := dao.db.
		Preload("Evento").
		Where("usuario_id = ?", usuarioID).
		Order("created_at DESC").
		Find(&entradas)
	return entradas, resultado.Error
}

// Actualizar persiste los cambios de una entrada existente
func (dao *entradaDAOImpl) Actualizar(entrada *domain.Entrada) error {
	return dao.db.Save(entrada).Error
}

// ContarActivas cuenta cuántas entradas activas tiene un usuario para un evento
// (útil para evitar duplicados si se quiere limitar a 1 entrada por persona por evento)
func (dao *entradaDAOImpl) ContarActivas(eventoID, usuarioID uint) (int64, error) {
	var count int64
	resultado := dao.db.Model(&domain.Entrada{}).
		Where("evento_id = ? AND usuario_id = ? AND estado = ?",
			eventoID, usuarioID, domain.EstadoEntradaActiva).
		Count(&count)
	return count, resultado.Error
}
