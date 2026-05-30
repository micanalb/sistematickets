package domain

import (
	"time"

	"gorm.io/gorm"
)

// RolUsuario define el tipo de usuario en el sistema
type RolUsuario string

const (
	RolCliente       RolUsuario = "cliente"
	RolAdministrador RolUsuario = "administrador"
)

// Usuario representa la entidad de usuario del sistema.
// Almacena los datos de autenticación y perfil de cada persona registrada.
type Usuario struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Nombre         string         `json:"nombre" gorm:"type:varchar(100);not null"`
	Apellido       string         `json:"apellido" gorm:"type:varchar(100);not null"`
	Email          string         `json:"email" gorm:"type:varchar(150);uniqueIndex;not null"`
	PasswordHash   string         `json:"-" gorm:"type:varchar(255);not null"` // Se omite en JSON por seguridad
	Rol            RolUsuario     `json:"rol" gorm:"type:enum('cliente','administrador');default:'cliente';not null"`
	Telefono       string         `json:"telefono" gorm:"type:varchar(20)"`
	FechaRegistro  time.Time      `json:"fecha_registro" gorm:"autoCreateTime"`
	Activo         bool           `json:"activo" gorm:"default:true"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete

	// Relaciones
	Entradas []Entrada `json:"entradas,omitempty" gorm:"foreignKey:UsuarioID"`
}

// TableName sobreescribe el nombre de tabla por convención en español
func (Usuario) TableName() string {
	return "usuarios"
}

// DTORegistro es el objeto de transferencia para el registro de un nuevo usuario
type DTORegistro struct {
	Nombre    string `json:"nombre" binding:"required,min=2,max=100"`
	Apellido  string `json:"apellido" binding:"required,min=2,max=100"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Telefono  string `json:"telefono"`
}

// DTOLogin es el objeto de transferencia para el inicio de sesión
type DTOLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// DTORespuestaLogin es la respuesta que devuelve el sistema al autenticarse exitosamente
type DTORespuestaLogin struct {
	Token   string   `json:"token"`
	Usuario UsuarioPublico `json:"usuario"`
}

// UsuarioPublico expone solo los datos no sensibles del usuario
type UsuarioPublico struct {
	ID       uint       `json:"id"`
	Nombre   string     `json:"nombre"`
	Apellido string     `json:"apellido"`
	Email    string     `json:"email"`
	Rol      RolUsuario `json:"rol"`
	Telefono string     `json:"telefono"`
}
