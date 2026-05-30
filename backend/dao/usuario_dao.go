package dao

import (
	"ticketsya/domain"

	"gorm.io/gorm"
)

// UsuarioDAO define el contrato de acceso a datos para la entidad Usuario.
// El uso de interfaz permite mockear en los tests unitarios.
type UsuarioDAO interface {
	Crear(usuario *domain.Usuario) error
	BuscarPorID(id uint) (*domain.Usuario, error)
	BuscarPorEmail(email string) (*domain.Usuario, error)
	Actualizar(usuario *domain.Usuario) error
	ExisteEmail(email string) (bool, error)
}

// usuarioDAOImpl es la implementación concreta con GORM
type usuarioDAOImpl struct {
	db *gorm.DB
}

// NuevoUsuarioDAO crea una nueva instancia del DAO de usuario
func NuevoUsuarioDAO(db *gorm.DB) UsuarioDAO {
	return &usuarioDAOImpl{db: db}
}

// Crear inserta un nuevo usuario en la base de datos
func (dao *usuarioDAOImpl) Crear(usuario *domain.Usuario) error {
	return dao.db.Create(usuario).Error
}

// BuscarPorID recupera un usuario por su ID primario
func (dao *usuarioDAOImpl) BuscarPorID(id uint) (*domain.Usuario, error) {
	var usuario domain.Usuario
	resultado := dao.db.First(&usuario, id)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &usuario, nil
}

// BuscarPorEmail recupera un usuario por su email (único en el sistema)
func (dao *usuarioDAOImpl) BuscarPorEmail(email string) (*domain.Usuario, error) {
	var usuario domain.Usuario
	resultado := dao.db.Where("email = ?", email).First(&usuario)
	if resultado.Error != nil {
		return nil, resultado.Error
	}
	return &usuario, nil
}

// Actualizar persiste los cambios de un usuario existente
func (dao *usuarioDAOImpl) Actualizar(usuario *domain.Usuario) error {
	return dao.db.Save(usuario).Error
}

// ExisteEmail verifica si un email ya está registrado (útil para validación en registro)
func (dao *usuarioDAOImpl) ExisteEmail(email string) (bool, error) {
	var count int64
	resultado := dao.db.Model(&domain.Usuario{}).Where("email = ?", email).Count(&count)
	return count > 0, resultado.Error
}
