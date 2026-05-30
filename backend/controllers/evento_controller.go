package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"
)

// EventoController maneja los endpoints relacionados a eventos
type EventoController struct {
	eventoService services.EventoService
}

// NuevoEventoController crea una nueva instancia del controlador de eventos
func NuevoEventoController(eventoService services.EventoService) *EventoController {
	return &EventoController{eventoService: eventoService}
}

// RegistrarRutas registra las rutas públicas de eventos
// Para la primera entrega solo se exponen las rutas del cliente (sin admin)
func (ctrl *EventoController) RegistrarRutas(router *gin.RouterGroup) {
	eventos := router.Group("/eventos")
	{
		eventos.GET("", ctrl.ListarEventos)
		eventos.GET("/:id", ctrl.ObtenerEvento)
	}
}

// ListarEventos retorna el catálogo de eventos con filtros opcionales
// GET /api/v1/eventos?busqueda=rock&categoria=musica&solo_disponibles=true
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

	utils.ResponderExito(c, gin.H{
		"eventos": eventos,
		"total":   len(eventos),
	})
}

// ObtenerEvento retorna el detalle completo de un evento por ID
// GET /api/v1/eventos/:id
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

// parsearIDParam es un helper para parsear el parámetro :id de la URL
func parsearIDParam(c *gin.Context, nombreParam string) (uint, error) {
	idStr := c.Param(nombreParam)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponderBadRequest(c, "ID inválido")
		return 0, err
	}
	return uint(id), nil
}
