package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ticketsya/controllers"
	"ticketsya/domain"
)

// ════════════════════════════════════════════════════════════════════
// Mock del EventoService
// ════════════════════════════════════════════════════════════════════

type MockEventoService struct {
	mock.Mock
}

func (m *MockEventoService) ListarEventos(filtros domain.FiltrosEvento) ([]domain.Evento, error) {
	args := m.Called(filtros)
	return args.Get(0).([]domain.Evento), args.Error(1)
}

func (m *MockEventoService) ObtenerEventoPorID(id uint) (*domain.Evento, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evento), args.Error(1)
}

func (m *MockEventoService) CrearEvento(dto domain.DTOCrearEvento) (*domain.Evento, error) {
	args := m.Called(dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evento), args.Error(1)
}

func (m *MockEventoService) ActualizarEvento(id uint, dto domain.DTOActualizarEvento) (*domain.Evento, error) {
	args := m.Called(id, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evento), args.Error(1)
}

func (m *MockEventoService) EliminarEvento(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEventoService) ObtenerReporte(id uint) (*domain.Evento, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evento), args.Error(1)
}

func setupRouterEvento(svc *MockEventoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NuevoEventoController(svc)
	api := router.Group("/api/v1")
	ctrl.RegistrarRutas(api)
	return router
}

// ════════════════════════════════════════════════════════════════════
// Tests de rutas públicas
// ════════════════════════════════════════════════════════════════════

func TestListarEventos_HTTP_200_SinFiltros(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ListarEventos", mock.AnythingOfType("domain.FiltrosEvento")).
		Return([]domain.Evento{
			{ID: 1, Titulo: "Evento A"},
			{ID: 2, Titulo: "Evento B"},
		}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/eventos", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListarEventos_HTTP_200_ConFiltros(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ListarEventos", mock.AnythingOfType("domain.FiltrosEvento")).
		Return([]domain.Evento{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/eventos?categoria=musica&solo_disponibles=true", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestObtenerEvento_HTTP_200_Encontrado(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerEventoPorID", uint(3)).Return(&domain.Evento{
		ID: 3, Titulo: "Evento Test", Estado: domain.EstadoActivo,
	}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/eventos/3", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestObtenerEvento_HTTP_404_NoExiste(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerEventoPorID", uint(999)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/eventos/999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestObtenerEvento_HTTP_400_IDInvalido(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/eventos/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
