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
// Mock del EntradaDAO
// ════════════════════════════════════════════════════════════════════

type MockEntradaDAO struct {
	mock.Mock
}

func (m *MockEntradaDAO) Crear(entrada *domain.Entrada) error {
	args := m.Called(entrada)
	return args.Error(0)
}

func (m *MockEntradaDAO) BuscarPorID(id uint) (*domain.Entrada, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entrada), args.Error(1)
}

func (m *MockEntradaDAO) BuscarPorCodigoQR(codigoQR string) (*domain.Entrada, error) {
	args := m.Called(codigoQR)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entrada), args.Error(1)
}

func (m *MockEntradaDAO) ListarPorUsuario(usuarioID uint) ([]domain.Entrada, error) {
	args := m.Called(usuarioID)
	return args.Get(0).([]domain.Entrada), args.Error(1)
}

func (m *MockEntradaDAO) Actualizar(entrada *domain.Entrada) error {
	args := m.Called(entrada)
	return args.Error(0)
}

func (m *MockEntradaDAO) ContarActivas(eventoID, usuarioID uint) (int64, error) {
	args := m.Called(eventoID, usuarioID)
	return args.Get(0).(int64), args.Error(1)
}

// Helper para crear un evento activo con disponibilidad
func eventoActivoConDisponibilidad() *domain.Evento {
	return &domain.Evento{
		ID:               1,
		Titulo:           "Evento Test",
		CapacidadTotal:   100,
		EntradasVendidas: 50,
		PrecioBase:       5000.00,
		Estado:           domain.EstadoActivo,
		FechaHora:        time.Now().Add(7 * 24 * time.Hour),
	}
}

// ════════════════════════════════════════════════════════════════════
// Tests de Comprar Entrada
// ════════════════════════════════════════════════════════════════════

func TestComprarEntrada_Exitoso(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	evento := eventoActivoConDisponibilidad()
	dto := domain.DTOComprarEntrada{EventoID: 1}

	mockEventoDAO.On("BuscarPorID", uint(1)).Return(evento, nil)
	mockEntradaDAO.On("Crear", mock.AnythingOfType("*domain.Entrada")).Return(nil)
	mockEventoDAO.On("IncrementarVentas", uint(1)).Return(nil)

	entrada, err := svc.ComprarEntrada(1, dto)

	assert.NoError(t, err)
	assert.NotNil(t, entrada)
	assert.Equal(t, uint(1), entrada.UsuarioID)
	assert.Equal(t, uint(1), entrada.EventoID)
	assert.Equal(t, domain.EstadoEntradaActiva, entrada.Estado)
	assert.Equal(t, evento.PrecioBase, entrada.PrecioPagado)
	assert.NotEmpty(t, entrada.CodigoQR)
}

func TestComprarEntrada_ErrorEventoNoExiste(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	mockEventoDAO.On("BuscarPorID", uint(999)).Return(nil, errors.New("record not found"))

	entrada, err := svc.ComprarEntrada(1, domain.DTOComprarEntrada{EventoID: 999})

	assert.Error(t, err)
	assert.Nil(t, entrada)
	assert.Contains(t, err.Error(), "record not found")
}

func TestComprarEntrada_ErrorEventoCancelado(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	eventoCancelado := eventoActivoConDisponibilidad()
	eventoCancelado.Estado = domain.EstadoCancelado // <-- cancelado

	mockEventoDAO.On("BuscarPorID", uint(1)).Return(eventoCancelado, nil)

	entrada, err := svc.ComprarEntrada(1, domain.DTOComprarEntrada{EventoID: 1})

	assert.Error(t, err)
	assert.Nil(t, entrada)
	assert.Contains(t, err.Error(), "no está disponible")
}

func TestComprarEntrada_ErrorEventoAgotado(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	eventoAgotado := eventoActivoConDisponibilidad()
	eventoAgotado.EntradasVendidas = eventoAgotado.CapacidadTotal // sin disponibilidad

	mockEventoDAO.On("BuscarPorID", uint(1)).Return(eventoAgotado, nil)

	entrada, err := svc.ComprarEntrada(1, domain.DTOComprarEntrada{EventoID: 1})

	assert.Error(t, err)
	assert.Nil(t, entrada)
	assert.Contains(t, err.Error(), "no hay entradas disponibles")
}

// ════════════════════════════════════════════════════════════════════
// Tests de Mis Entradas
// ════════════════════════════════════════════════════════════════════

func TestMisEntradas_RetornaEntradasDelUsuario(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	entradasEsperadas := []domain.Entrada{
		{ID: 1, UsuarioID: 5, Estado: domain.EstadoEntradaActiva},
		{ID: 2, UsuarioID: 5, Estado: domain.EstadoEntradaCancelada},
	}

	mockEntradaDAO.On("ListarPorUsuario", uint(5)).Return(entradasEsperadas, nil)

	entradas, err := svc.MisEntradas(5)

	assert.NoError(t, err)
	assert.Len(t, entradas, 2)
}

func TestMisEntradas_ListaVaciaParaUsuarioSinCompras(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	mockEntradaDAO.On("ListarPorUsuario", uint(99)).Return([]domain.Entrada{}, nil)

	entradas, err := svc.MisEntradas(99)

	assert.NoError(t, err)
	assert.Empty(t, entradas)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Cancelar Entrada
// ════════════════════════════════════════════════════════════════════

func TestCancelarEntrada_Exitoso(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: 5,
		EventoID:  1,
		Estado:    domain.EstadoEntradaActiva,
	}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockEntradaDAO.On("Actualizar", mock.AnythingOfType("*domain.Entrada")).Return(nil)
	mockEventoDAO.On("DecrementarVentas", uint(1)).Return(nil)

	err := svc.CancelarEntrada(1, 5)

	assert.NoError(t, err)
}

func TestCancelarEntrada_ErrorPropietarioDistinto(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: 5, // pertenece al usuario 5
		Estado:    domain.EstadoEntradaActiva,
	}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	// El usuario 99 intenta cancelar la entrada del usuario 5
	err := svc.CancelarEntrada(1, 99)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permisos")
}

func TestCancelarEntrada_ErrorEntradaYaCancelada(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: 5,
		Estado:    domain.EstadoEntradaCancelada, // ya cancelada
	}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	err := svc.CancelarEntrada(1, 5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")
}

// ════════════════════════════════════════════════════════════════════
// Tests de Transferir Entrada
// ════════════════════════════════════════════════════════════════════

func TestTransferirEntrada_Exitoso(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	propietario := uint(5)
	destinatario := &domain.Usuario{
		ID:       10,
		Nombre:   "Carlos",
		Apellido: "Ruiz",
		Email:    "carlos@example.com",
	}

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: propietario,
		EventoID:  2,
		Estado:    domain.EstadoEntradaActiva,
		Evento:    &domain.Evento{ID: 2, Titulo: "Evento Test"},
	}

	dto := domain.DTOTransferirEntrada{EmailDestinatario: "carlos@example.com"}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockUsuarioDAO.On("BuscarPorEmail", "carlos@example.com").Return(destinatario, nil)
	mockEntradaDAO.On("Actualizar", mock.AnythingOfType("*domain.Entrada")).Return(nil)
	mockEntradaDAO.On("Crear", mock.AnythingOfType("*domain.Entrada")).Return(nil)

	nuevaEntrada, err := svc.TransferirEntrada(1, propietario, dto)

	assert.NoError(t, err)
	assert.NotNil(t, nuevaEntrada)
	assert.Equal(t, destinatario.ID, nuevaEntrada.UsuarioID)
}

func TestTransferirEntrada_ErrorAutoTransferencia(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	usuarioID := uint(5)
	mismoUsuario := &domain.Usuario{
		ID:    usuarioID,
		Email: "yo@example.com",
	}

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: usuarioID,
		Estado:    domain.EstadoEntradaActiva,
	}

	dto := domain.DTOTransferirEntrada{EmailDestinatario: "yo@example.com"}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockUsuarioDAO.On("BuscarPorEmail", "yo@example.com").Return(mismoUsuario, nil)

	nuevaEntrada, err := svc.TransferirEntrada(1, usuarioID, dto)

	assert.Error(t, err)
	assert.Nil(t, nuevaEntrada)
	assert.Contains(t, err.Error(), "vos mismo")
}

func TestTransferirEntrada_ErrorDestinatarioNoRegistrado(t *testing.T) {
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: 5,
		Estado:    domain.EstadoEntradaActiva,
	}

	dto := domain.DTOTransferirEntrada{EmailDestinatario: "noexiste@example.com"}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)
	mockUsuarioDAO.On("BuscarPorEmail", "noexiste@example.com").Return(nil, errors.New("record not found"))

	nuevaEntrada, err := svc.TransferirEntrada(1, 5, dto)

	assert.Error(t, err)
	assert.Nil(t, nuevaEntrada)
	assert.Contains(t, err.Error(), "record not found")
}
