package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"ticketsya/controllers"
	"ticketsya/domain"
	"ticketsya/utils"
)

func setupRouterMiddleware() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Ruta protegida solo con autenticación
	router.GET("/protegida",
		controllers.MiddlewareAutenticacion(),
		func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) },
	)

	// Ruta protegida con autenticación + rol admin
	router.GET("/solo-admin",
		controllers.MiddlewareAutenticacion(),
		controllers.MiddlewareRolAdmin(),
		func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) },
	)

	return router
}

func TestMiddlewareAutenticacion_SinHeader_401(t *testing.T) {
	router := setupRouterMiddleware()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protegida", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddlewareAutenticacion_FormatoInvalido_401(t *testing.T) {
	router := setupRouterMiddleware()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protegida", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddlewareAutenticacion_TokenInvalido_401(t *testing.T) {
	router := setupRouterMiddleware()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protegida", nil)
	req.Header.Set("Authorization", "Bearer token.invalido.aqui")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddlewareAutenticacion_TokenValido_200(t *testing.T) {
	router := setupRouterMiddleware()
	token, _ := utils.GenerarTokenJWT(1, "test@test.com", domain.RolCliente)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protegida", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddlewareRolAdmin_ConTokenCliente_403(t *testing.T) {
	router := setupRouterMiddleware()
	token, _ := utils.GenerarTokenJWT(1, "cliente@test.com", domain.RolCliente)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/solo-admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddlewareRolAdmin_ConTokenAdmin_200(t *testing.T) {
	router := setupRouterMiddleware()
	token, _ := utils.GenerarTokenJWT(1, "admin@test.com", domain.RolAdministrador)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/solo-admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
