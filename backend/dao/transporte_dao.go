package dao

import (
	"ticketsya/domain"

	"gorm.io/gorm"
)

// TransporteDAO define el contrato de acceso a datos para AsistenteTransporte
type TransporteDAO interface {
	Crear(asistente *domain.AsistenteTransporte) error
	BuscarPorEntradaID(entradaID uint) (*domain.AsistenteTransporte, error)
	BuscarPorID(id uint) (*domain.AsistenteTransporte, error)
	Actualizar(asistente *domain.AsistenteTransporte) error
	// ListarComparteAutoPorEvento devuelve todos los registros donde algún
	// usuario ofreció compartir su auto para ese evento. Se usa en la Parte 2
	// (matching) para mostrarle a un usuario las ofertas disponibles.
	ListarComparteAutoPorEvento(eventoID uint, excluirUsuarioID uint) ([]domain.AsistenteTransporte, error)
}

type transporteDAOImpl struct {
	db *gorm.DB
}

// NuevoTransporteDAO crea una nueva instancia del DAO de transporte
func NuevoTransporteDAO(db *gorm.DB) TransporteDAO {
	return &transporteDAOImpl{db: db}
}

func (dao *transporteDAOImpl) Crear(asistente *domain.AsistenteTransporte) error {
	return dao.db.Create(asistente).Error
}

// BuscarPorEntradaID es la consulta principal: cada entrada tiene a lo sumo
// un asistente de transporte configurado (relación 1 a 1).
func (dao *transporteDAOImpl) BuscarPorEntradaID(entradaID uint) (*domain.AsistenteTransporte, error) {
	var asistente domain.AsistenteTransporte
	resultado := dao.db.
		Preload("Entrada").
		Preload("Evento").
		Where("entrada_id = ?", entradaID).
		First(&asistente)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &asistente, nil
}

func (dao *transporteDAOImpl) BuscarPorID(id uint) (*domain.AsistenteTransporte, error) {
	var asistente domain.AsistenteTransporte
	resultado := dao.db.
		Preload("Entrada").
		Preload("Evento").
		Preload("Usuario").
		Preload("UsuarioMatch").
		First(&asistente, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &asistente, nil
}

func (dao *transporteDAOImpl) Actualizar(asistente *domain.AsistenteTransporte) error {
	return dao.db.Save(asistente).Error
}

func (dao *transporteDAOImpl) ListarComparteAutoPorEvento(eventoID uint, excluirUsuarioID uint) ([]domain.AsistenteTransporte, error) {
	var asistentes []domain.AsistenteTransporte
	resultado := dao.db.
		Preload("Usuario").
		Where("evento_id = ? AND modo = ? AND comparte_auto = ? AND usuario_id != ?",
			eventoID, domain.ModoAutoPropio, true, excluirUsuarioID).
		Find(&asistentes)
	return asistentes, resultado.Error
}
