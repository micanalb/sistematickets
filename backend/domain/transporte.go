package domain

import (
	"time"

	"gorm.io/gorm"
)

// ModoTransporte define cómo el usuario planea llegar al evento
type ModoTransporte string

const (
	ModoColectivo  ModoTransporte = "colectivo"
	ModoAutoPropio ModoTransporte = "auto_propio"
	ModoCompartido ModoTransporte = "compartido"
)

// EstadoMatch define el estado de una solicitud para unirse a un auto compartido.
// Es nil/vacío mientras nadie pidió unirse; toma valor cuando alguien solicita
// compartir el viaje con el dueño del auto.
type EstadoMatch string

const (
	EstadoMatchPendiente EstadoMatch = "pendiente"
	EstadoMatchAprobado  EstadoMatch = "aprobado"
	EstadoMatchRechazado EstadoMatch = "rechazado"
)

// AsistenteTransporte representa la preferencia de transporte que un usuario
// configuró para una entrada/evento específico. Es 1 a 1 con la entrada:
// cada ticket puede tener su propio plan de cómo llegar.
//
// Los campos de "compartir auto" (ComparteAuto, EstadoMatch, UsuarioMatchID)
// solo tienen sentido cuando Modo == ModoAutoPropio o ModoCompartido — se
// completa en la Parte 2 del bonus (funcionalidad de matching).
type AsistenteTransporte struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	EntradaID uint           `json:"entrada_id" gorm:"not null;uniqueIndex"`
	UsuarioID uint           `json:"usuario_id" gorm:"not null;index"`
	EventoID  uint           `json:"evento_id" gorm:"not null;index"`
	Modo      ModoTransporte `json:"modo" gorm:"type:varchar(20);not null"`

	// Específico de modo=colectivo
	LineaColectivo string `json:"linea_colectivo,omitempty" gorm:"type:varchar(100)"`

	// Específico de modo=auto_propio / compartido (se usa en la Parte 2)
	ComparteAuto   bool         `json:"comparte_auto" gorm:"default:false"`
	EstadoMatch    *EstadoMatch `json:"estado_match,omitempty" gorm:"type:varchar(20)"`
	UsuarioMatchID *uint        `json:"usuario_match_id,omitempty" gorm:"index"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relaciones
	Entrada      *Entrada `json:"entrada,omitempty" gorm:"foreignKey:EntradaID"`
	Usuario      *Usuario `json:"usuario,omitempty" gorm:"foreignKey:UsuarioID"`
	Evento       *Evento  `json:"evento,omitempty" gorm:"foreignKey:EventoID"`
	UsuarioMatch *Usuario `json:"usuario_match,omitempty" gorm:"foreignKey:UsuarioMatchID"`
}

// TableName sobreescribe el nombre de tabla
func (AsistenteTransporte) TableName() string {
	return "asistentes_transporte"
}

// LineaColectivoInfo representa una línea de colectivo disponible para
// mostrarle al usuario, con el link a sus horarios. Por ahora es un catálogo
// fijo en código — no requiere tabla propia porque no lo gestiona el admin.
type LineaColectivoInfo struct {
	Linea      string `json:"linea"`
	Recorrido  string `json:"recorrido"`
	URLHorario string `json:"url_horario"`
}

// EstacionamientoInfo representa un estacionamiento cercano sugerido para
// quienes van en auto propio. Catálogo fijo en código, igual que las líneas.
type EstacionamientoInfo struct {
	Nombre    string  `json:"nombre"`
	Direccion string  `json:"direccion"`
	Distancia string  `json:"distancia"` // ej: "350m del lugar"
	Latitud   float64 `json:"latitud"`
	Longitud  float64 `json:"longitud"`
}

// ── DTOs ────────────────────────────────────────────────────────────

// DTOCrearAsistenteTransporte es el objeto de transferencia para configurar
// el transporte de una entrada por primera vez.
type DTOCrearAsistenteTransporte struct {
	EntradaID      uint           `json:"entrada_id" binding:"required"`
	Modo           ModoTransporte `json:"modo" binding:"required"`
	LineaColectivo string         `json:"linea_colectivo"`
	ComparteAuto   bool           `json:"comparte_auto"`
}

// DTORespuestaAsistente es lo que se devuelve al consultar el asistente de
// una entrada: la configuración guardada + la info de apoyo según el modo
// (líneas de colectivo o estacionamientos), para que el frontend no tenga
// que pedirla por separado.
type DTORespuestaAsistente struct {
	Asistente        *AsistenteTransporte  `json:"asistente"`
	LineasColectivo  []LineaColectivoInfo  `json:"lineas_colectivo,omitempty"`
	Estacionamientos []EstacionamientoInfo `json:"estacionamientos,omitempty"`
}
