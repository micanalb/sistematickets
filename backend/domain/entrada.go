package domain

import (
	"time"

	"gorm.io/gorm"
)

// EstadoEntrada define el ciclo de vida de una entrada
type EstadoEntrada string

const (
	EstadoEntradaActiva    EstadoEntrada = "activa"
	EstadoEntradaCancelada EstadoEntrada = "cancelada"
	EstadoEntradaUsada     EstadoEntrada = "usada"
	EstadoEntradaTransferida EstadoEntrada = "transferida"
)

// Entrada representa un ticket adquirido por un usuario para un evento específico.
// Es la entidad transaccional principal del sistema.
type Entrada struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	CodigoQR        string         `json:"codigo_qr" gorm:"type:varchar(100);uniqueIndex;not null"`
	UsuarioID       uint           `json:"usuario_id" gorm:"not null;index"`
	EventoID        uint           `json:"evento_id" gorm:"not null;index"`
	PrecioPagado    float64        `json:"precio_pagado" gorm:"type:decimal(10,2);not null"`
	Estado          EstadoEntrada  `json:"estado" gorm:"type:enum('activa','cancelada','usada','transferida');default:'activa'"`
	FechaCompra     time.Time      `json:"fecha_compra" gorm:"autoCreateTime"`
	FechaCancelacion *time.Time    `json:"fecha_cancelacion,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones - se cargan con Preload
	Usuario *Usuario `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	Evento  *Evento  `json:"evento,omitempty" gorm:"foreignKey:EventoID"`
}

// TableName sobreescribe el nombre de tabla
func (Entrada) TableName() string {
	return "entradas"
}

// DTOComprarEntrada es el objeto de transferencia para la compra de una entrada
type DTOComprarEntrada struct {
	EventoID uint `json:"evento_id" binding:"required"`
}

// DTOTransferirEntrada es el objeto para transferir una entrada a otro usuario
type DTOTransferirEntrada struct {
	EmailDestinatario string `json:"email_destinatario" binding:"required,email"`
}

// DTORespuestaEntrada es la respuesta estructurada de una entrada con sus relaciones
type DTORespuestaEntrada struct {
	ID           uint          `json:"id"`
	CodigoQR     string        `json:"codigo_qr"`
	Estado       EstadoEntrada `json:"estado"`
	PrecioPagado float64       `json:"precio_pagado"`
	FechaCompra  time.Time     `json:"fecha_compra"`
	Evento       *Evento       `json:"evento,omitempty"`
	Usuario      *UsuarioPublico `json:"usuario,omitempty"`
}
