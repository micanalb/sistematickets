package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ticketsya/domain"
	"ticketsya/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// generarTokenAdminTest crea un token JWT valido con rol administrador,
// para probar las rutas /admin/eventos que requieren MiddlewareRolAdmin.
func generarTokenAdminTest(usuarioID uint) string {
	token, _ := utils.GenerarTokenJWT(usuarioID, "admin@test.com", domain.RolAdministrador)
	return "Bearer " + token
}

// ── ListarEventosAdmin ───────────────────────────────────────────────

func TestListarEventosAdmin_HTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ListarEventos", mock.AnythingOfType("domain.FiltrosEvento")).
		Return([]domain.Evento{{ID: 1, Titulo: "Evento Admin"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos", nil)
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListarEventosAdmin_HTTP_403_RolCliente(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListarEventosAdmin_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ── CrearEvento ──────────────────────────────────────────────────────

func TestCrearEvento_HTTP_201_Exitoso(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("CrearEvento", mock.AnythingOfType("domain.DTOCrearEvento")).
		Return(&domain.Evento{ID: 1, Titulo: "Nuevo Evento", Estado: domain.EstadoActivo}, nil)

	body, _ := json.Marshal(map[string]interface{}{
		"titulo": "Nuevo Evento", "fecha_hora": time.Now().Add(48 * time.Hour),
		"duracion_minutos": 120, "lugar": "Estadio", "categoria": "musica",
		"capacidad_total": 1000, "precio_base": 5000,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/eventos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCrearEvento_HTTP_403_RolCliente(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	body, _ := json.Marshal(map[string]interface{}{"titulo": "Evento"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/eventos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCrearEvento_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("CrearEvento", mock.AnythingOfType("domain.DTOCrearEvento")).
		Return(nil, assert.AnError)

	body, _ := json.Marshal(map[string]interface{}{
		"titulo": "Evento Invalido", "fecha_hora": time.Now().Add(48 * time.Hour),
		"duracion_minutos": 120, "lugar": "Estadio", "categoria": "musica",
		"capacidad_total": 0, "precio_base": 5000,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/eventos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCrearEvento_HTTP_400_BodyInvalido(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	body := []byte(`{}`) // faltan campos requeridos
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/eventos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── ActualizarEvento ─────────────────────────────────────────────────

func TestActualizarEventoHTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ActualizarEvento", uint(1), mock.AnythingOfType("domain.DTOActualizarEvento")).
		Return(&domain.Evento{ID: 1, Titulo: "Titulo Actualizado"}, nil)

	body, _ := json.Marshal(map[string]string{"titulo": "Titulo Actualizado"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/admin/eventos/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestActualizarEventoHTTP_403_RolCliente(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	body, _ := json.Marshal(map[string]string{"titulo": "Titulo"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/admin/eventos/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestActualizarEventoHTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ActualizarEvento", uint(99), mock.AnythingOfType("domain.DTOActualizarEvento")).
		Return(nil, assert.AnError)

	body, _ := json.Marshal(map[string]string{"titulo": "No existe"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/admin/eventos/99", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestActualizarEventoHTTP_400_IDInvalido(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	body, _ := json.Marshal(map[string]string{"titulo": "Titulo"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/admin/eventos/abc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── EliminarEvento ───────────────────────────────────────────────────

func TestEliminarEventoHTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("EliminarEvento", uint(1)).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/admin/eventos/1", nil)
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEliminarEventoHTTP_403_RolCliente(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/admin/eventos/1", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestEliminarEventoHTTP_404_NoEncontrado(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("EliminarEvento", uint(99)).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/admin/eventos/99", nil)
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEliminarEventoHTTP_400_IDInvalido(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/admin/eventos/abc", nil)
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── ObtenerReporte ───────────────────────────────────────────────────

func TestObtenerReporteHTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerReporte", uint(1)).Return(&domain.Evento{
		ID: 1, Titulo: "Evento con reporte",
		CapacidadTotal: 100, EntradasVendidas: 40,
		Entradas: []domain.Entrada{{ID: 1, UsuarioID: 5}},
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos/1/reporte", nil)
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestObtenerReporteHTTP_403_RolCliente(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos/1/reporte", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestObtenerReporteHTTP_404_NoEncontrado(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerReporte", uint(99)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos/99/reporte", nil)
	req.Header.Set("Authorization", generarTokenAdminTest(1))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestObtenerReporteHTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/eventos/1/reporte", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
