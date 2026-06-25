package services_test

import (
	"errors"
	"testing"

	"ticketsya/domain"
	"ticketsya/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ============================================================
// Mock del TransporteDAO
// ============================================================

type MockTransporteDAO struct {
	mock.Mock
}

func (m *MockTransporteDAO) Crear(asistente *domain.AsistenteTransporte) error {
	args := m.Called(asistente)
	return args.Error(0)
}

func (m *MockTransporteDAO) BuscarPorEntradaID(entradaID uint) (*domain.AsistenteTransporte, error) {
	args := m.Called(entradaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AsistenteTransporte), args.Error(1)
}

func (m *MockTransporteDAO) BuscarPorID(id uint) (*domain.AsistenteTransporte, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AsistenteTransporte), args.Error(1)
}

func (m *MockTransporteDAO) Actualizar(asistente *domain.AsistenteTransporte) error {
	args := m.Called(asistente)
	return args.Error(0)
}

func (m *MockTransporteDAO) ListarComparteAutoPorEvento(eventoID uint, excluirUsuarioID uint) ([]domain.AsistenteTransporte, error) {
	args := m.Called(eventoID, excluirUsuarioID)
	return args.Get(0).([]domain.AsistenteTransporte), args.Error(1)
}

func (m *MockTransporteDAO) ListarSolicitudesPendientesPorDueno(usuarioID uint) ([]domain.AsistenteTransporte, error) {
	args := m.Called(usuarioID)
	return args.Get(0).([]domain.AsistenteTransporte), args.Error(1)
}

// ============================================================
// Helpers
// ============================================================

func entradaActivaTest(id, usuarioID, eventoID uint) *domain.Entrada {
	return &domain.Entrada{
		ID:        id,
		UsuarioID: usuarioID,
		EventoID:  eventoID,
		Estado:    domain.EstadoEntradaActiva,
	}
}

// ============================================================
// Tests de ConfigurarTransporte
// ============================================================

func TestConfigurarTransporte_ColectivoExitoso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10)
	dto := domain.DTOCrearAsistenteTransporte{
		EntradaID:      1,
		Modo:           domain.ModoColectivo,
		LineaColectivo: "Linea 7",
	}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockTransporteDAO.On("BuscarPorEntradaID", uint(1)).Return(nil, errors.New("record not found"))
	mockTransporteDAO.On("Crear", mock.AnythingOfType("*domain.AsistenteTransporte")).Return(nil)

	respuesta, err := svc.ConfigurarTransporte(5, dto)

	assert.NoError(t, err)
	assert.NotNil(t, respuesta)
	assert.NotNil(t, respuesta.Asistente)
	assert.Equal(t, domain.ModoColectivo, respuesta.Asistente.Modo)
	assert.Equal(t, "Linea 7", respuesta.Asistente.LineaColectivo)
	assert.NotEmpty(t, respuesta.LineasColectivo)
	assert.Empty(t, respuesta.Estacionamientos)
}

func TestConfigurarTransporte_AutoPropioExitoso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(2, 5, 10)
	dto := domain.DTOCrearAsistenteTransporte{
		EntradaID:    2,
		Modo:         domain.ModoAutoPropio,
		ComparteAuto: true,
	}

	mockEntradaDAO.On("BuscarPorID", uint(2)).Return(entrada, nil)
	mockTransporteDAO.On("BuscarPorEntradaID", uint(2)).Return(nil, errors.New("record not found"))
	mockTransporteDAO.On("Crear", mock.AnythingOfType("*domain.AsistenteTransporte")).Return(nil)

	respuesta, err := svc.ConfigurarTransporte(5, dto)

	assert.NoError(t, err)
	assert.Equal(t, domain.ModoAutoPropio, respuesta.Asistente.Modo)
	assert.True(t, respuesta.Asistente.ComparteAuto)
	assert.NotEmpty(t, respuesta.Estacionamientos)
	assert.Empty(t, respuesta.LineasColectivo)
}

func TestConfigurarTransporte_ActualizaConfiguracionExistente(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(3, 5, 10)
	existente := &domain.AsistenteTransporte{
		ID:        1,
		EntradaID: 3,
		UsuarioID: 5,
		EventoID:  10,
		Modo:      domain.ModoColectivo,
	}

	dto := domain.DTOCrearAsistenteTransporte{
		EntradaID:    3,
		Modo:         domain.ModoAutoPropio,
		ComparteAuto: false,
	}

	mockEntradaDAO.On("BuscarPorID", uint(3)).Return(entrada, nil)
	mockTransporteDAO.On("BuscarPorEntradaID", uint(3)).Return(existente, nil)
	mockTransporteDAO.On("Actualizar", mock.AnythingOfType("*domain.AsistenteTransporte")).Return(nil)

	respuesta, err := svc.ConfigurarTransporte(5, dto)

	assert.NoError(t, err)
	assert.Equal(t, domain.ModoAutoPropio, respuesta.Asistente.Modo)
	mockTransporteDAO.AssertNotCalled(t, "Crear", mock.Anything)
}

func TestConfigurarTransporte_ErrorEntradaNoEncontrada(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	mockEntradaDAO.On("BuscarPorID", uint(99)).Return(nil, gorm.ErrRecordNotFound)

	dto := domain.DTOCrearAsistenteTransporte{EntradaID: 99, Modo: domain.ModoColectivo, LineaColectivo: "Linea 1"}

	respuesta, err := svc.ConfigurarTransporte(5, dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "no encontrada")
}

func TestConfigurarTransporte_ErrorPropietarioDistinto(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10) // pertenece al usuario 5
	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	dto := domain.DTOCrearAsistenteTransporte{EntradaID: 1, Modo: domain.ModoColectivo, LineaColectivo: "Linea 1"}

	respuesta, err := svc.ConfigurarTransporte(99, dto) // usuario 99 intenta configurar la entrada de otro

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "permisos")
}

func TestConfigurarTransporte_ErrorEntradaNoActiva(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10)
	entrada.Estado = domain.EstadoEntradaCancelada

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	dto := domain.DTOCrearAsistenteTransporte{EntradaID: 1, Modo: domain.ModoColectivo, LineaColectivo: "Linea 1"}

	respuesta, err := svc.ConfigurarTransporte(5, dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "activas")
}

func TestConfigurarTransporte_ErrorColectivoSinLinea(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10)
	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	dto := domain.DTOCrearAsistenteTransporte{EntradaID: 1, Modo: domain.ModoColectivo, LineaColectivo: ""}

	respuesta, err := svc.ConfigurarTransporte(5, dto)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "línea")
}

// ============================================================
// Tests de ObtenerPorEntrada
// ============================================================

func TestObtenerPorEntrada_Exitoso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10)
	asistente := &domain.AsistenteTransporte{
		ID: 1, EntradaID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: false,
	}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockTransporteDAO.On("BuscarPorEntradaID", uint(1)).Return(asistente, nil)

	respuesta, err := svc.ObtenerPorEntrada(5, 1)

	assert.NoError(t, err)
	assert.NotNil(t, respuesta.Asistente)
	assert.NotEmpty(t, respuesta.Estacionamientos)
}

func TestObtenerPorEntrada_SinConfiguracionPrevia(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10)

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockTransporteDAO.On("BuscarPorEntradaID", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	respuesta, err := svc.ObtenerPorEntrada(5, 1)

	assert.NoError(t, err)
	assert.NotNil(t, respuesta)
	assert.Nil(t, respuesta.Asistente)
}

func TestObtenerPorEntrada_ErrorPropietarioDistinto(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	entrada := entradaActivaTest(1, 5, 10)
	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	respuesta, err := svc.ObtenerPorEntrada(99, 1)

	assert.Error(t, err)
	assert.Nil(t, respuesta)
	assert.Contains(t, err.Error(), "permisos")
}

// ============================================================
// Tests de ListarOfertasAuto
// ============================================================

func TestListarOfertasAuto_RetornaOfertas(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	ofertasEsperadas := []domain.AsistenteTransporte{
		{ID: 1, UsuarioID: 5, EventoID: 10, Modo: domain.ModoAutoPropio, ComparteAuto: true},
	}
	mockTransporteDAO.On("ListarComparteAutoPorEvento", uint(10), uint(7)).Return(ofertasEsperadas, nil)

	ofertas, err := svc.ListarOfertasAuto(7, 10)

	assert.NoError(t, err)
	assert.Len(t, ofertas, 1)
}

func TestListarOfertasAuto_ListaVacia(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	mockTransporteDAO.On("ListarComparteAutoPorEvento", uint(10), uint(7)).Return([]domain.AsistenteTransporte{}, nil)

	ofertas, err := svc.ListarOfertasAuto(7, 10)

	assert.NoError(t, err)
	assert.Empty(t, ofertas)
}

// ============================================================
// Tests de SolicitarCompartir
// ============================================================

func TestSolicitarCompartir_Exitoso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	oferta := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
	}
	actualizado := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
	}

	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(oferta, nil).Once()
	mockUsuarioDAO.On("BuscarPorID", uint(7)).Return(&domain.Usuario{ID: 7}, nil)
	mockTransporteDAO.On("Actualizar", mock.AnythingOfType("*domain.AsistenteTransporte")).Return(nil)
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(actualizado, nil).Once()

	resultado, err := svc.SolicitarCompartir(7, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resultado)
}

func TestSolicitarCompartir_ErrorNoEsOfertaCompartida(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	oferta := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoColectivo, ComparteAuto: false,
	}
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(oferta, nil)

	resultado, err := svc.SolicitarCompartir(7, 1)

	assert.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "no es una oferta")
}

func TestSolicitarCompartir_ErrorPropioAuto(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	oferta := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
	}
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(oferta, nil)

	resultado, err := svc.SolicitarCompartir(5, 1) // el mismo dueño intenta solicitar su propio auto

	assert.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "propio auto")
}

func TestSolicitarCompartir_ErrorYaTieneSolicitudEnCurso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	pendiente := domain.EstadoMatchPendiente
	oferta := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
		EstadoMatch: &pendiente,
	}
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(oferta, nil)

	resultado, err := svc.SolicitarCompartir(7, 1)

	assert.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "en curso")
}

func TestSolicitarCompartir_ErrorOfertaNoEncontrada(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	mockTransporteDAO.On("BuscarPorID", uint(99)).Return(nil, gorm.ErrRecordNotFound)
	resultado, err := svc.SolicitarCompartir(7, 99)

	assert.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "no encontrada")
}

// ============================================================
// Tests de ResponderSolicitud
// ============================================================

func TestResponderSolicitud_AprobarExitoso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	pendiente := domain.EstadoMatchPendiente
	usuarioMatchID := uint(7)
	asistente := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
		EstadoMatch: &pendiente, UsuarioMatchID: &usuarioMatchID,
	}

	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(asistente, nil).Once()
	mockTransporteDAO.On("Actualizar", mock.AnythingOfType("*domain.AsistenteTransporte")).Return(nil)
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(asistente, nil).Once()

	resultado, err := svc.ResponderSolicitud(5, 1, true)

	assert.NoError(t, err)
	assert.NotNil(t, resultado)
}

func TestResponderSolicitud_RechazarExitoso(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	pendiente := domain.EstadoMatchPendiente
	usuarioMatchID := uint(7)
	asistente := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
		EstadoMatch: &pendiente, UsuarioMatchID: &usuarioMatchID,
	}

	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(asistente, nil).Once()
	mockTransporteDAO.On("Actualizar", mock.AnythingOfType("*domain.AsistenteTransporte")).Return(nil)
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(asistente, nil).Once()

	resultado, err := svc.ResponderSolicitud(5, 1, false)

	assert.NoError(t, err)
	assert.NotNil(t, resultado)
}

func TestResponderSolicitud_ErrorNoEsDueno(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	pendiente := domain.EstadoMatchPendiente
	asistente := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10, // pertenece al usuario 5
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
		EstadoMatch: &pendiente,
	}
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(asistente, nil)

	resultado, err := svc.ResponderSolicitud(99, 1, true) // usuario 99 no es el dueño

	assert.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "permisos")
}

func TestResponderSolicitud_ErrorSinSolicitudPendiente(t *testing.T) {
	mockTransporteDAO := new(MockTransporteDAO)
	mockEntradaDAO := new(MockEntradaDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoTransporteService(mockTransporteDAO, mockEntradaDAO, mockUsuarioDAO)

	asistente := &domain.AsistenteTransporte{
		ID: 1, UsuarioID: 5, EventoID: 10,
		Modo: domain.ModoAutoPropio, ComparteAuto: true,
		EstadoMatch: nil, // sin ninguna solicitud
	}
	mockTransporteDAO.On("BuscarPorID", uint(1)).Return(asistente, nil)

	resultado, err := svc.ResponderSolicitud(5, 1, true)

	assert.Error(t, err)
	assert.Nil(t, resultado)
	assert.Contains(t, err.Error(), "pendiente")
}
