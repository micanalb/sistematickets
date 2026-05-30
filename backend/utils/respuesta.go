package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespuestaExito estructura estándar para respuestas exitosas
type RespuestaExito struct {
	Exito  bool        `json:"exito"`
	Datos  interface{} `json:"datos,omitempty"`
	Mensaje string     `json:"mensaje,omitempty"`
}

// RespuestaError estructura estándar para respuestas de error
type RespuestaError struct {
	Exito   bool   `json:"exito"`
	Error   string `json:"error"`
	Codigo  int    `json:"codigo"`
}

// ResponderExito envía una respuesta JSON 200 OK con datos
func ResponderExito(c *gin.Context, datos interface{}) {
	c.JSON(http.StatusOK, RespuestaExito{
		Exito: true,
		Datos: datos,
	})
}

// ResponderExitoConMensaje envía una respuesta JSON 200 OK con datos y mensaje
func ResponderExitoConMensaje(c *gin.Context, datos interface{}, mensaje string) {
	c.JSON(http.StatusOK, RespuestaExito{
		Exito:  true,
		Datos:  datos,
		Mensaje: mensaje,
	})
}

// ResponderCreado envía una respuesta JSON 201 Created
func ResponderCreado(c *gin.Context, datos interface{}) {
	c.JSON(http.StatusCreated, RespuestaExito{
		Exito: true,
		Datos: datos,
	})
}

// ResponderError envía una respuesta JSON de error con código HTTP dado
func ResponderError(c *gin.Context, codigoHTTP int, mensaje string) {
	c.JSON(codigoHTTP, RespuestaError{
		Exito:  false,
		Error:  mensaje,
		Codigo: codigoHTTP,
	})
}

// ResponderBadRequest envía 400 Bad Request
func ResponderBadRequest(c *gin.Context, mensaje string) {
	ResponderError(c, http.StatusBadRequest, mensaje)
}

// ResponderNoAutorizado envía 401 Unauthorized
func ResponderNoAutorizado(c *gin.Context, mensaje string) {
	ResponderError(c, http.StatusUnauthorized, mensaje)
}

// ResponderForbidden envía 403 Forbidden
func ResponderForbidden(c *gin.Context, mensaje string) {
	ResponderError(c, http.StatusForbidden, mensaje)
}

// ResponderNoEncontrado envía 404 Not Found
func ResponderNoEncontrado(c *gin.Context, mensaje string) {
	ResponderError(c, http.StatusNotFound, mensaje)
}

// ResponderErrorInterno envía 500 Internal Server Error
func ResponderErrorInterno(c *gin.Context, mensaje string) {
	ResponderError(c, http.StatusInternalServerError, mensaje)
}

// ResponderConflicto envía 409 Conflict
func ResponderConflicto(c *gin.Context, mensaje string) {
	ResponderError(c, http.StatusConflict, mensaje)
}
