package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"
)

type EventoController struct {
	eventoService services.EventoService
}

func NuevoEventoController(eventoService services.EventoService) *EventoController {
	return &EventoController{eventoService: eventoService}
}

func (ctrl *EventoController) RegistrarRutas(router *gin.RouterGroup) {
	// Rutas públicas
	eventos := router.Group("/eventos")
	{
		eventos.GET("", ctrl.ListarEventos)
		eventos.GET("/:id", ctrl.ObtenerEvento)
	}

	// Rutas admin
	admin := router.Group("/admin/eventos")
	admin.Use(MiddlewareAutenticacion(), MiddlewareRolAdmin())
	{
		admin.GET("", ctrl.ListarEventosAdmin)
		admin.POST("", ctrl.CrearEvento)
		admin.PUT("/:id", ctrl.ActualizarEvento)
		admin.DELETE("/:id", ctrl.EliminarEvento)
		admin.GET("/:id/reporte", ctrl.ObtenerReporte)
	}
}

func (ctrl *EventoController) ListarEventos(c *gin.Context) {
	var filtros domain.FiltrosEvento
	if err := c.ShouldBindQuery(&filtros); err != nil {
		utils.ResponderBadRequest(c, "filtros inválidos: "+err.Error())
		return
	}
	eventos, err := ctrl.eventoService.ListarEventos(filtros)
	if err != nil {
		utils.ResponderErrorInterno(c, "error al obtener eventos")
		return
	}
	utils.ResponderExito(c, gin.H{"eventos": eventos, "total": len(eventos)})
}

func (ctrl *EventoController) ObtenerEvento(c *gin.Context) {
	id, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}
	evento, err := ctrl.eventoService.ObtenerEventoPorID(id)
	if err != nil {
		utils.ResponderNoEncontrado(c, err.Error())
		return
	}
	utils.ResponderExito(c, evento)
}

func (ctrl *EventoController) ListarEventosAdmin(c *gin.Context) {
	var filtros domain.FiltrosEvento
	if err := c.ShouldBindQuery(&filtros); err != nil {
		utils.ResponderBadRequest(c, "filtros inválidos: "+err.Error())
		return
	}
	eventos, err := ctrl.eventoService.ListarEventos(filtros)
	if err != nil {
		utils.ResponderErrorInterno(c, "error al obtener eventos")
		return
	}
	utils.ResponderExito(c, gin.H{"eventos": eventos, "total": len(eventos)})
}

func (ctrl *EventoController) CrearEvento(c *gin.Context) {
	var dto domain.DTOCrearEvento
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos del evento inválidos: "+err.Error())
		return
	}
	evento, err := ctrl.eventoService.CrearEvento(dto)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}
	utils.ResponderCreado(c, evento)
}

func (ctrl *EventoController) ActualizarEvento(c *gin.Context) {
	id, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}
	var dto domain.DTOActualizarEvento
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos inválidos: "+err.Error())
		return
	}
	evento, err := ctrl.eventoService.ActualizarEvento(id, dto)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}
	utils.ResponderExito(c, evento)
}

func (ctrl *EventoController) EliminarEvento(c *gin.Context) {
	id, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}
	if err := ctrl.eventoService.EliminarEvento(id); err != nil {
		utils.ResponderNoEncontrado(c, err.Error())
		return
	}
	utils.ResponderExitoConMensaje(c, nil, "evento eliminado correctamente")
}

func (ctrl *EventoController) ObtenerReporte(c *gin.Context) {
	id, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}
	evento, err := ctrl.eventoService.ObtenerReporte(id)
	if err != nil {
		utils.ResponderNoEncontrado(c, err.Error())
		return
	}
	porcentaje := 0.0
	if evento.CapacidadTotal > 0 {
		porcentaje = float64(evento.EntradasVendidas) / float64(evento.CapacidadTotal) * 100
	}
	utils.ResponderExito(c, gin.H{
		"evento":               evento,
		"capacidad_total":      evento.CapacidadTotal,
		"entradas_vendidas":    evento.EntradasVendidas,
		"entradas_disponibles": evento.DisponibilidadEntradas(),
		"porcentaje_ocupacion": porcentaje,
		"compradores":          evento.Entradas,
	})
}

func parsearIDParam(c *gin.Context, nombreParam string) (uint, error) {
	idStr := c.Param(nombreParam)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponderBadRequest(c, "ID inválido")
		return 0, err
	}
	return uint(id), nil
}
