package controllers

import (
	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"

	"github.com/gin-gonic/gin"
)

// TransporteController maneja los endpoints del asistente de transporte (bonus track)
type TransporteController struct {
	transporteService services.TransporteService
}

// NuevoTransporteController crea una nueva instancia del controlador de transporte
func NuevoTransporteController(transporteService services.TransporteService) *TransporteController {
	return &TransporteController{transporteService: transporteService}
}

// RegistrarRutas registra las rutas del asistente de transporte.
// Todas requieren autenticación porque el transporte se configura sobre
// una entrada propia del usuario — es una funcionalidad de Cliente, no admin.
func (ctrl *TransporteController) RegistrarRutas(router *gin.RouterGroup) {
	transporte := router.Group("/transporte")
	transporte.Use(MiddlewareAutenticacion())
	{
		transporte.POST("", ctrl.ConfigurarTransporte)
		transporte.GET("/entrada/:entradaId", ctrl.ObtenerPorEntrada)
		transporte.GET("/ofertas/:eventoId", ctrl.ListarOfertasAuto)
		transporte.POST("/:id/solicitar", ctrl.SolicitarCompartir)
		transporte.PUT("/:id/responder", ctrl.ResponderSolicitud)
	}
}

// ConfigurarTransporte crea o actualiza la preferencia de transporte de una entrada
// POST /api/v1/transporte
func (ctrl *TransporteController) ConfigurarTransporte(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	var dto domain.DTOCrearAsistenteTransporte
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos de transporte inválidos: "+err.Error())
		return
	}

	respuesta, err := ctrl.transporteService.ConfigurarTransporte(usuarioID, dto)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderCreado(c, respuesta)
}

// ObtenerPorEntrada devuelve la configuración de transporte de una entrada,
// junto con el catálogo de apoyo (líneas de colectivo o estacionamientos)
// GET /api/v1/transporte/entrada/:entradaId
func (ctrl *TransporteController) ObtenerPorEntrada(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	entradaID, err := parsearIDParam(c, "entradaId")
	if err != nil {
		return
	}

	respuesta, err := ctrl.transporteService.ObtenerPorEntrada(usuarioID, entradaID)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderExito(c, respuesta)
}

// ListarOfertasAuto devuelve quiénes ofrecen compartir auto para un evento
// GET /api/v1/transporte/ofertas/:eventoId
func (ctrl *TransporteController) ListarOfertasAuto(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	eventoID, err := parsearIDParam(c, "eventoId")
	if err != nil {
		return
	}

	ofertas, err := ctrl.transporteService.ListarOfertasAuto(usuarioID, eventoID)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderExito(c, gin.H{
		"ofertas": ofertas,
		"total":   len(ofertas),
	})
}

// SolicitarCompartir registra la intención de un usuario de unirse al auto
// de otro. :id es el ID del registro AsistenteTransporte del DUEÑO del auto
// (el que aparece en ListarOfertasAuto), no del solicitante.
// POST /api/v1/transporte/:id/solicitar
func (ctrl *TransporteController) SolicitarCompartir(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	asistenteOfertaID, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}

	resultado, err := ctrl.transporteService.SolicitarCompartir(usuarioID, asistenteOfertaID)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderExitoConMensaje(c, resultado, "solicitud enviada, queda pendiente de aprobación del dueño")
}

// ResponderSolicitud es la acción del dueño del auto para aprobar o rechazar
// PUT /api/v1/transporte/:id/responder
// Body esperado: {"aprobar": true} o {"aprobar": false}
func (ctrl *TransporteController) ResponderSolicitud(c *gin.Context) {
	duenoID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	asistenteID, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Aprobar bool `json:"aprobar"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ResponderBadRequest(c, "se requiere el campo 'aprobar' (true/false)")
		return
	}

	resultado, err := ctrl.transporteService.ResponderSolicitud(duenoID, asistenteID, body.Aprobar)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	mensaje := "solicitud rechazada"
	if body.Aprobar {
		mensaje = "solicitud aprobada, ya pueden ver sus datos de contacto"
	}
	utils.ResponderExitoConMensaje(c, resultado, mensaje)
}