package domain

import (
	"time"

	"gorm.io/gorm"
)

// CategoriaEvento define las categorías posibles de un evento
type CategoriaEvento string

const (
	CategoriaMusica     CategoriaEvento = "musica"
	CategoriaDeporte    CategoriaEvento = "deporte"
	CategoriaCultura    CategoriaEvento = "cultura"
	CategoriaTeatroCine CategoriaEvento = "teatro_cine"
	CategoriaConferencia CategoriaEvento = "conferencia"
	CategoriaOtro       CategoriaEvento = "otro"
)

// EstadoEvento define el ciclo de vida de un evento
type EstadoEvento string

const (
	EstadoActivo    EstadoEvento = "activo"
	EstadoCancelado EstadoEvento = "cancelado"
	EstadoAgotado   EstadoEvento = "agotado"
	EstadoFinalizado EstadoEvento = "finalizado"
)

// Evento representa la entidad central del sistema.
// Contiene toda la información de un evento al que se pueden comprar entradas.
type Evento struct {
	ID               uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	Titulo           string          `json:"titulo" gorm:"type:varchar(200);not null"`
	Descripcion      string          `json:"descripcion" gorm:"type:text"`
	FechaHora        time.Time       `json:"fecha_hora" gorm:"not null"`
	DuracionMinutos  int             `json:"duracion_minutos" gorm:"not null;default:120"`
	Lugar            string          `json:"lugar" gorm:"type:varchar(200);not null"`
	Direccion        string          `json:"direccion" gorm:"type:varchar(300)"`
	Ciudad           string          `json:"ciudad" gorm:"type:varchar(100)"`
	Categoria        CategoriaEvento `json:"categoria" gorm:"type:enum('musica','deporte','cultura','teatro_cine','conferencia','otro');not null"`
	CapacidadTotal   int             `json:"capacidad_total" gorm:"not null"`
	EntradasVendidas int             `json:"entradas_vendidas" gorm:"default:0"`
	PrecioBase       float64         `json:"precio_base" gorm:"type:decimal(10,2);not null"`
	ImagenURL        string          `json:"imagen_url" gorm:"type:varchar(500)"`
	Estado           EstadoEvento    `json:"estado" gorm:"type:enum('activo','cancelado','agotado','finalizado');default:'activo'"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"-" gorm:"index"` // Soft delete

	// Relaciones
	Entradas []Entrada `json:"entradas,omitempty" gorm:"foreignKey:EventoID"`
}

// TableName sobreescribe el nombre de tabla
func (Evento) TableName() string {
	return "eventos"
}

// DisponibilidadEntradas retorna cuántas entradas quedan disponibles
func (e *Evento) DisponibilidadEntradas() int {
	return e.CapacidadTotal - e.EntradasVendidas
}

// TieneDisponibilidad verifica si hay entradas disponibles
func (e *Evento) TieneDisponibilidad() bool {
	return e.DisponibilidadEntradas() > 0 && e.Estado == EstadoActivo
}

// DTOCrearEvento es el objeto de transferencia para crear un evento
type DTOCrearEvento struct {
	Titulo          string          `json:"titulo" binding:"required,min=3,max=200"`
	Descripcion     string          `json:"descripcion"`
	FechaHora       time.Time       `json:"fecha_hora" binding:"required"`
	DuracionMinutos int             `json:"duracion_minutos" binding:"required,min=1"`
	Lugar           string          `json:"lugar" binding:"required"`
	Direccion       string          `json:"direccion"`
	Ciudad          string          `json:"ciudad"`
	Categoria       CategoriaEvento `json:"categoria" binding:"required"`
	CapacidadTotal  int             `json:"capacidad_total" binding:"required,min=1"`
	PrecioBase      float64         `json:"precio_base" binding:"required,min=0"`
	ImagenURL       string          `json:"imagen_url"`
}

// DTOActualizarEvento es el objeto para actualizar un evento (todos los campos son opcionales)
type DTOActualizarEvento struct {
	Titulo          *string          `json:"titulo"`
	Descripcion     *string          `json:"descripcion"`
	FechaHora       *time.Time       `json:"fecha_hora"`
	DuracionMinutos *int             `json:"duracion_minutos"`
	Lugar           *string          `json:"lugar"`
	Direccion       *string          `json:"direccion"`
	Ciudad          *string          `json:"ciudad"`
	Categoria       *CategoriaEvento `json:"categoria"`
	CapacidadTotal  *int             `json:"capacidad_total"`
	PrecioBase      *float64         `json:"precio_base"`
	ImagenURL       *string          `json:"imagen_url"`
	Estado          *EstadoEvento    `json:"estado"`
}

// FiltrosEvento encapsula los criterios de búsqueda del catálogo
type FiltrosEvento struct {
	Busqueda  string          `form:"busqueda"`
	Categoria CategoriaEvento `form:"categoria"`
	Ciudad    string          `form:"ciudad"`
	FechaDesde *time.Time     `form:"fecha_desde" time_format:"2006-01-02"`
	FechaHasta *time.Time     `form:"fecha_hasta" time_format:"2006-01-02"`
	SoloDisponibles bool      `form:"solo_disponibles"`
}
