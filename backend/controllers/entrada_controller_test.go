package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticketsya/controllers"
	"ticketsya/domain"
	"ticketsya/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ════════════════════════════════════════════════════════════════════
// Mock del EntradaService
// ════════════════════════════════════════════════════════════════════

type MockEntradaService struct {
	mock.Mock
}

func (m *MockEntradaService) ComprarEntrada(usuarioID uint, dto domain.DTOComprarEntrada) ([]domain.Entrada, error) {
	args := m.Called(usuarioID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Entrada), args.Error(1)
}

func (m *MockEntradaService) MisEntradas(usuarioID uint) ([]domain.Entrada, error) {
	args := m.Called(usuarioID)
	return args.Get(0).([]domain.Entrada), args.Error(1)
}

func (m *MockEntradaService) CancelarEntrada(entradaID, usuarioID uint) error {
	args := m.Called(entradaID, usuarioID)
	return args.Error(0)
}

func (m *MockEntradaService) TransferirEntrada(entradaID, usuarioID uint, dto domain.DTOTransferirEntrada) (*domain.Entrada, error) {
	args := m.Called(entradaID, usuarioID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entrada), args.Error(1)
}

func setupRouterEntrada(svc *MockEntradaService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NuevoEntradaController(svc)
	api := router.Group("/api/v1")
	ctrl.RegistrarRutas(api)
	return router
}

func generarTokenClienteTest(usuarioID uint) string {
	token, _ := utils.GenerarTokenJWT(usuarioID, "cliente@test.com", domain.RolCliente)
	return "Bearer " + token
}

// ── Comprar ───────────────────────────────────────────────────────

func TestComprarEntrada_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/entradas/comprar", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestComprarEntrada_HTTP_201_Exitoso(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	mockSvc.On("ComprarEntrada", uint(5), mock.AnythingOfType("domain.DTOComprarEntrada")).
		Return([]domain.Entrada{
			{ID: 1, CodigoQR: "TKT-1-5-abc", Estado: domain.EstadoEntradaActiva},
		}, nil)

	body, _ := json.Marshal(map[string]interface{}{"evento_id": 1})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/entradas/comprar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestComprarEntrada_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	mockSvc.On("ComprarEntrada", uint(5), mock.AnythingOfType("domain.DTOComprarEntrada")).
		Return(nil, assert.AnError)

	body, _ := json.Marshal(map[string]interface{}{"evento_id": 99})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/entradas/comprar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── Mis Entradas ──────────────────────────────────────────────────

func TestMisEntradas_HTTP_200_ConHistorial(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	mockSvc.On("MisEntradas", uint(7)).Return([]domain.Entrada{
		{ID: 1, UsuarioID: 7, Estado: domain.EstadoEntradaActiva},
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/entradas/mis-entradas", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(7))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMisEntradas_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/entradas/mis-entradas", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ── Cancelar ──────────────────────────────────────────────────────

func TestCancelarEntrada_HTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	mockSvc.On("CancelarEntrada", uint(3), uint(5)).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/3/cancelar", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCancelarEntrada_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/3/cancelar", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCancelarEntrada_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	mockSvc.On("CancelarEntrada", uint(3), uint(5)).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/3/cancelar", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelarEntrada_HTTP_400_IDInvalido(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/abc/cancelar", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── Transferir ────────────────────────────────────────────────────

func TestTransferirEntrada_HTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	mockSvc.On("TransferirEntrada", uint(4), uint(5), mock.AnythingOfType("domain.DTOTransferirEntrada")).
		Return(&domain.Entrada{ID: 99, UsuarioID: 10, Estado: domain.EstadoEntradaActiva}, nil)

	body, _ := json.Marshal(map[string]string{"email_destinatario": "destino@example.com"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/4/transferir", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTransferirEntrada_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/4/transferir", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTransferirEntrada_HTTP_400_EmailFaltante(t *testing.T) {
	mockSvc := new(MockEntradaService)
	router := setupRouterEntrada(mockSvc)

	body := []byte(`{}`) // sin email_destinatario
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/entradas/4/transferir", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
