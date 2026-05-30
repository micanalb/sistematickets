package controllers

import (
	"github.com/gin-gonic/gin"
	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"
)

// EntradaController maneja los endpoints de compra, cancelación y transferencia de entradas
type EntradaController struct {
	entradaService services.EntradaService
}

// NuevoEntradaController crea una nueva instancia del controlador de entradas
func NuevoEntradaController(entradaService services.EntradaService) *EntradaController {
	return &EntradaController{entradaService: entradaService}
}

// RegistrarRutas registra las rutas de entradas en el router
// Todas las rutas de entradas requieren autenticación (son operaciones del cliente)
func (ctrl *EntradaController) RegistrarRutas(router *gin.RouterGroup) {
	entradas := router.Group("/entradas")
	entradas.Use(MiddlewareAutenticacion())
	{
		entradas.POST("/comprar", ctrl.ComprarEntrada)
		entradas.GET("/mis-entradas", ctrl.MisEntradas)
		entradas.PUT("/:id/cancelar", ctrl.CancelarEntrada)
		entradas.PUT("/:id/transferir", ctrl.TransferirEntrada)
	}
}

// ComprarEntrada procesa la compra de una entrada
// POST /api/v1/entradas/comprar
func (ctrl *EntradaController) ComprarEntrada(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	var dto domain.DTOComprarEntrada
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos de compra inválidos: "+err.Error())
		return
	}

	entrada, err := ctrl.entradaService.ComprarEntrada(usuarioID, dto)
	if err != nil {
		// Distintos errores de negocio que pueden ocurrir
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderCreado(c, entrada)
}

// MisEntradas retorna el historial de entradas del usuario autenticado
// GET /api/v1/entradas/mis-entradas
func (ctrl *EntradaController) MisEntradas(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	entradas, err := ctrl.entradaService.MisEntradas(usuarioID)
	if err != nil {
		utils.ResponderErrorInterno(c, "error al obtener entradas")
		return
	}

	utils.ResponderExito(c, gin.H{
		"entradas": entradas,
		"total":    len(entradas),
	})
}

// CancelarEntrada procesa la cancelación de un ticket
// PUT /api/v1/entradas/:id/cancelar
func (ctrl *EntradaController) CancelarEntrada(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	entradaID, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}

	if err := ctrl.entradaService.CancelarEntrada(entradaID, usuarioID); err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderExitoConMensaje(c, nil, "entrada cancelada correctamente")
}

// TransferirEntrada transfiere un ticket a otro usuario
// PUT /api/v1/entradas/:id/transferir
func (ctrl *EntradaController) TransferirEntrada(c *gin.Context) {
	usuarioID, ok := obtenerUsuarioIDDelContexto(c)
	if !ok {
		utils.ResponderNoAutorizado(c, "usuario no autenticado")
		return
	}

	entradaID, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}

	var dto domain.DTOTransferirEntrada
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.ResponderBadRequest(c, "datos de transferencia inválidos: "+err.Error())
		return
	}

	entrada, err := ctrl.entradaService.TransferirEntrada(entradaID, usuarioID, dto)
	if err != nil {
		utils.ResponderBadRequest(c, err.Error())
		return
	}

	utils.ResponderExitoConMensaje(c, entrada, "entrada transferida correctamente")
}
