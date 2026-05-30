package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ticketsya/domain"
	"ticketsya/services"
)

// ════════════════════════════════════════════════════════════════════
// Mock del EventoDAO
// ════════════════════════════════════════════════════════════════════

type MockEventoDAO struct {
	mock.Mock
}

func (m *MockEventoDAO) Crear(evento *domain.Evento) error {
	args := m.Called(evento)
	return args.Error(0)
}

func (m *MockEventoDAO) BuscarPorID(id uint) (*domain.Evento, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evento), args.Error(1)
}

func (m *MockEventoDAO) ListarConFiltros(filtros domain.FiltrosEvento) ([]domain.Evento, error) {
	args := m.Called(filtros)
	return args.Get(0).([]domain.Evento), args.Error(1)
}

func (m *MockEventoDAO) Actualizar(evento *domain.Evento) error {
	args := m.Called(evento)
	return args.Error(0)
}

func (m *MockEventoDAO) Eliminar(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEventoDAO) IncrementarVentas(eventoID uint) error {
	args := m.Called(eventoID)
	return args.Error(0)
}

func (m *MockEventoDAO) DecrementarVentas(eventoID uint) error {
	args := m.Called(eventoID)
	return args.Error(0)
}

func (m *MockEventoDAO) ReportePorID(id uint) (*domain.Evento, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Evento), args.Error(1)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Listar Eventos
// ════════════════════════════════════════════════════════════════════

func TestListarEventos_RetornaListaVacia(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	filtros := domain.FiltrosEvento{}
	mockDAO.On("ListarConFiltros", filtros).Return([]domain.Evento{}, nil)

	eventos, err := svc.ListarEventos(filtros)

	assert.NoError(t, err)
	assert.Empty(t, eventos)
}

func TestListarEventos_RetornaEventosDisponibles(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventosEsperados := []domain.Evento{
		{ID: 1, Titulo: "Lollapalooza 2024", Estado: domain.EstadoActivo},
		{ID: 2, Titulo: "River vs Boca", Estado: domain.EstadoActivo},
	}

	filtros := domain.FiltrosEvento{}
	mockDAO.On("ListarConFiltros", filtros).Return(eventosEsperados, nil)

	eventos, err := svc.ListarEventos(filtros)

	assert.NoError(t, err)
	assert.Len(t, eventos, 2)
	assert.Equal(t, "Lollapalooza 2024", eventos[0].Titulo)
}

func TestListarEventos_ErrorDelDAO(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	filtros := domain.FiltrosEvento{}
	mockDAO.On("ListarConFiltros", filtros).Return([]domain.Evento{}, errors.New("error de conexión"))

	eventos, err := svc.ListarEventos(filtros)

	assert.Error(t, err)
	assert.Nil(t, eventos)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Obtener Evento por ID
// ════════════════════════════════════════════════════════════════════

func TestObtenerEventoPorID_Exitoso(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventoEsperado := &domain.Evento{
		ID:     5,
		Titulo: "Tech Talk 2024",
		Estado: domain.EstadoActivo,
	}

	mockDAO.On("BuscarPorID", uint(5)).Return(eventoEsperado, nil)

	evento, err := svc.ObtenerEventoPorID(5)

	assert.NoError(t, err)
	assert.NotNil(t, evento)
	assert.Equal(t, uint(5), evento.ID)
	assert.Equal(t, "Tech Talk 2024", evento.Titulo)
}

func TestObtenerEventoPorID_NoEncontrado(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	mockDAO.On("BuscarPorID", uint(999)).Return(nil, errors.New("record not found"))

	evento, err := svc.ObtenerEventoPorID(999)

	assert.Error(t, err)
	assert.Nil(t, evento)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Crear Evento
// ════════════════════════════════════════════════════════════════════

func TestCrearEvento_Exitoso(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	dto := domain.DTOCrearEvento{
		Titulo:          "Coldplay Argentina 2024",
		Descripcion:     "Gira Music of the Spheres",
		FechaHora:       time.Now().Add(30 * 24 * time.Hour),
		DuracionMinutos: 180,
		Lugar:           "Estadio Monumental",
		Categoria:       domain.CategoriaMusica,
		CapacidadTotal:  80000,
		PrecioBase:      45000.00,
	}

	mockDAO.On("Crear", mock.AnythingOfType("*domain.Evento")).Return(nil)

	evento, err := svc.CrearEvento(dto)

	assert.NoError(t, err)
	assert.NotNil(t, evento)
	assert.Equal(t, dto.Titulo, evento.Titulo)
	assert.Equal(t, domain.EstadoActivo, evento.Estado)
}

func TestCrearEvento_ErrorCapacidadCero(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	dto := domain.DTOCrearEvento{
		Titulo:         "Evento sin capacidad",
		CapacidadTotal: 0, // inválido
		PrecioBase:     100.0,
		DuracionMinutos: 60,
		Categoria:      domain.CategoriaOtro,
		FechaHora:      time.Now(),
		Lugar:          "Algún lugar",
	}

	evento, err := svc.CrearEvento(dto)

	assert.Error(t, err)
	assert.Nil(t, evento)
	assert.Contains(t, err.Error(), "capacidad")
	mockDAO.AssertNotCalled(t, "Crear", mock.Anything)
}

func TestCrearEvento_ErrorPrecioNegativo(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	dto := domain.DTOCrearEvento{
		Titulo:          "Evento precio negativo",
		CapacidadTotal:  100,
		PrecioBase:      -50.0, // inválido
		DuracionMinutos: 60,
		Categoria:       domain.CategoriaOtro,
		FechaHora:       time.Now(),
		Lugar:           "Algún lugar",
	}

	evento, err := svc.CrearEvento(dto)

	assert.Error(t, err)
	assert.Nil(t, evento)
	assert.Contains(t, err.Error(), "precio")
}

// ════════════════════════════════════════════════════════════════════
// Tests de Eliminar Evento
// ════════════════════════════════════════════════════════════════════

func TestEliminarEvento_Exitoso(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	evento := &domain.Evento{ID: 1, Titulo: "Evento a eliminar"}
	mockDAO.On("BuscarPorID", uint(1)).Return(evento, nil)
	mockDAO.On("Eliminar", uint(1)).Return(nil)

	err := svc.EliminarEvento(1)

	assert.NoError(t, err)
	mockDAO.AssertExpectations(t)
}

func TestEliminarEvento_NoEncontrado(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	mockDAO.On("BuscarPorID", uint(999)).Return(nil, errors.New("record not found"))

	err := svc.EliminarEvento(999)

	assert.Error(t, err)
	mockDAO.AssertNotCalled(t, "Eliminar", mock.Anything)
}
