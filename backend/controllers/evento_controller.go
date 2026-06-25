package controllers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ticketsya/domain"
	"ticketsya/services"
	"ticketsya/utils"

	"github.com/gin-gonic/gin"
)

// directorioUploadsEventos es la carpeta donde se guardan fisicamente las
// imagenes subidas. Relativa a la raiz del backend (donde corre main.go).
const directorioUploadsEventos = "uploads/eventos"

// extensionesPermitidas restringe que tipos de archivo se aceptan, tanto
// para evitar subir archivos arbitrarios (ej: .exe, .php) como para
// asegurar que el navegador pueda renderizarlo como imagen.
var extensionesPermitidas = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// tamanioMaximoImagen limita el peso del archivo a 5 MB para evitar que
// alguien llene el disco del servidor con uploads gigantes.
const tamanioMaximoImagen = 5 << 20 // 5 MB en bytes

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
		admin.POST("/:id/imagen", ctrl.SubirImagen)
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

// SubirImagen recibe un archivo de imagen (multipart/form-data, campo "imagen"),
// lo valida, lo guarda en disco con un nombre único y actualiza el campo
// imagen_url del evento para que apunte a la ruta pública servida por el backend.
// POST /api/v1/admin/eventos/:id/imagen
func (ctrl *EventoController) SubirImagen(c *gin.Context) {
	id, err := parsearIDParam(c, "id")
	if err != nil {
		return
	}

	// Verificar que el evento exista antes de aceptar el archivo
	_, err = ctrl.eventoService.ObtenerEventoPorID(id)
	if err != nil {
		utils.ResponderNoEncontrado(c, "evento no encontrado")
		return
	}

	// Limitar el tamaño máximo del body para esta request específica
	c.Request.Body = utils.LimitarTamanioBody(c.Request.Body, tamanioMaximoImagen)

	archivo, err := c.FormFile("imagen")
	if err != nil {
		utils.ResponderBadRequest(c, "no se recibió ningún archivo en el campo 'imagen'")
		return
	}

	if archivo.Size > tamanioMaximoImagen {
		utils.ResponderBadRequest(c, "la imagen supera el tamaño máximo permitido (5 MB)")
		return
	}

	extension := strings.ToLower(filepath.Ext(archivo.Filename))
	if !extensionesPermitidas[extension] {
		utils.ResponderBadRequest(c, "formato no permitido. Usá JPG, PNG o WEBP")
		return
	}

	// Nombre único para evitar colisiones entre eventos o re-subidas:
	// evento-<id>-<timestamp><extensión>
	nombreArchivo := fmt.Sprintf("evento-%d-%d%s", id, time.Now().UnixNano(), extension)
	rutaDestino := filepath.Join(directorioUploadsEventos, nombreArchivo)

	if err := c.SaveUploadedFile(archivo, rutaDestino); err != nil {
		utils.ResponderErrorInterno(c, "error al guardar la imagen en el servidor")
		return
	}

	// La URL pública depende de cómo main.go sirve el directorio estático;
	// se sirve bajo /uploads/eventos/<archivo> (ver router.Static en main.go)
	urlPublica := fmt.Sprintf("/uploads/eventos/%s", nombreArchivo)

	dto := domain.DTOActualizarEvento{ImagenURL: &urlPublica}
	eventoActualizado, err := ctrl.eventoService.ActualizarEvento(id, dto)
	if err != nil {
		utils.ResponderErrorInterno(c, "la imagen se guardó pero no se pudo actualizar el evento")
		return
	}

	utils.ResponderExito(c, gin.H{
		"imagen_url": urlPublica,
		"evento":     eventoActualizado,
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
