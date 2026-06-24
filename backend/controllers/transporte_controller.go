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
