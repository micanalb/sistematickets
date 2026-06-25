package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticketsya/controllers"
	"ticketsya/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================
// Mock del TransporteService
// ============================================================

type MockTransporteService struct {
	mock.Mock
}

func (m *MockTransporteService) ConfigurarTransporte(usuarioID uint, dto domain.DTOCrearAsistenteTransporte) (*domain.DTORespuestaAsistente, error) {
	args := m.Called(usuarioID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DTORespuestaAsistente), args.Error(1)
}

func (m *MockTransporteService) ObtenerPorEntrada(usuarioID, entradaID uint) (*domain.DTORespuestaAsistente, error) {
	args := m.Called(usuarioID, entradaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DTORespuestaAsistente), args.Error(1)
}

func (m *MockTransporteService) ListarOfertasAuto(usuarioID, eventoID uint) ([]domain.AsistenteTransporte, error) {
	args := m.Called(usuarioID, eventoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.AsistenteTransporte), args.Error(1)
}

func (m *MockTransporteService) SolicitarCompartir(usuarioID uint, asistenteOfertaID uint) (*domain.AsistenteTransporte, error) {
	args := m.Called(usuarioID, asistenteOfertaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AsistenteTransporte), args.Error(1)
}

func (m *MockTransporteService) ResponderSolicitud(duenoID uint, asistenteID uint, aprobar bool) (*domain.AsistenteTransporte, error) {
	args := m.Called(duenoID, asistenteID, aprobar)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AsistenteTransporte), args.Error(1)
}

func setupRouterTransporte(svc *MockTransporteService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NuevoTransporteController(svc)
	api := router.Group("/api/v1")
	ctrl.RegistrarRutas(api)
	return router
}

// ── ConfigurarTransporte ────────────────────────────────────────────

func TestConfigurarTransporte_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestConfigurarTransporte_HTTP_201_Exitoso(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("ConfigurarTransporte", uint(5), mock.AnythingOfType("domain.DTOCrearAsistenteTransporte")).
		Return(&domain.DTORespuestaAsistente{
			Asistente: &domain.AsistenteTransporte{ID: 1, EntradaID: 1, Modo: domain.ModoColectivo},
		}, nil)

	body, _ := json.Marshal(map[string]interface{}{"entrada_id": 1, "modo": "colectivo", "linea_colectivo": "Linea 1"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestConfigurarTransporte_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("ConfigurarTransporte", uint(5), mock.AnythingOfType("domain.DTOCrearAsistenteTransporte")).
		Return(nil, assert.AnError)

	body, _ := json.Marshal(map[string]interface{}{"entrada_id": 99, "modo": "colectivo", "linea_colectivo": "Linea 1"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigurarTransporte_HTTP_400_BodyInvalido(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	body := []byte(`{}`) // sin entrada_id ni modo
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── ObtenerPorEntrada ────────────────────────────────────────────────

func TestObtenerPorEntradaTransporte_HTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("ObtenerPorEntrada", uint(5), uint(1)).
		Return(&domain.DTORespuestaAsistente{
			Asistente: &domain.AsistenteTransporte{ID: 1, EntradaID: 1, Modo: domain.ModoAutoPropio},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transporte/entrada/1", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestObtenerPorEntradaTransporte_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transporte/entrada/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestObtenerPorEntradaTransporte_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("ObtenerPorEntrada", uint(5), uint(99)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transporte/entrada/99", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── ListarOfertasAuto ────────────────────────────────────────────────

func TestListarOfertasAuto_HTTP_200_ConOfertas(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("ListarOfertasAuto", uint(7), uint(10)).Return([]domain.AsistenteTransporte{
		{ID: 1, UsuarioID: 5, EventoID: 10, Modo: domain.ModoAutoPropio, ComparteAuto: true},
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transporte/ofertas/10", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(7))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListarOfertasAuto_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/transporte/ofertas/10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ── SolicitarCompartir ───────────────────────────────────────────────

func TestSolicitarCompartir_HTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	pendiente := domain.EstadoMatchPendiente
	mockSvc.On("SolicitarCompartir", uint(7), uint(1)).
		Return(&domain.AsistenteTransporte{ID: 1, UsuarioID: 5, EstadoMatch: &pendiente}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte/1/solicitar", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(7))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSolicitarCompartir_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte/1/solicitar", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSolicitarCompartir_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("SolicitarCompartir", uint(7), uint(1)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte/1/solicitar", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(7))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSolicitarCompartir_HTTP_400_IDInvalido(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/transporte/abc/solicitar", nil)
	req.Header.Set("Authorization", generarTokenClienteTest(7))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── ResponderSolicitud ───────────────────────────────────────────────

func TestResponderSolicitudTransporte_HTTP_200_Aprobar(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	aprobado := domain.EstadoMatchAprobado
	mockSvc.On("ResponderSolicitud", uint(5), uint(1), true).
		Return(&domain.AsistenteTransporte{ID: 1, UsuarioID: 5, EstadoMatch: &aprobado}, nil)

	body, _ := json.Marshal(map[string]bool{"aprobar": true})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/transporte/1/responder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponderSolicitudTransporte_HTTP_200_Rechazar(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	rechazado := domain.EstadoMatchRechazado
	mockSvc.On("ResponderSolicitud", uint(5), uint(1), false).
		Return(&domain.AsistenteTransporte{ID: 1, UsuarioID: 5, EstadoMatch: &rechazado}, nil)

	body, _ := json.Marshal(map[string]bool{"aprobar": false})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/transporte/1/responder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponderSolicitudTransporte_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	body, _ := json.Marshal(map[string]bool{"aprobar": true})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/transporte/1/responder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestResponderSolicitudTransporte_HTTP_400_ErrorServicio(t *testing.T) {
	mockSvc := new(MockTransporteService)
	router := setupRouterTransporte(mockSvc)

	mockSvc.On("ResponderSolicitud", uint(5), uint(1), true).Return(nil, assert.AnError)

	body, _ := json.Marshal(map[string]bool{"aprobar": true})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/transporte/1/responder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generarTokenClienteTest(5))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
