package services_test

import (
	"errors"
	"testing"
	"time"

	"ticketsya/domain"
	"ticketsya/services"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
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

// ════════════════════════════════════════════════════════════════════
// BD de prueba en memoria (SQLite)
//
// db.Transaction() necesita una conexión *gorm.DB real para abrir un
// BEGIN/COMMIT/ROLLBACK de verdad — no se puede mockear con testify.
// SQLite en memoria nos da esa conexión real sin depender de MySQL
// levantado ni de credenciales: cada test arranca con una BD limpia
// y descartable.
// ════════════════════════════════════════════════════════════════════

// dbTestEnMemoria crea una conexión SQLite en memoria y crea las tablas
// necesarias a mano con SQL crudo.
//
// No usamos db.AutoMigrate(&domain.Usuario{}, ...) acá porque el struct
// domain.Usuario tiene la tag `gorm:"type:enum('cliente','administrador')"`,
// que es sintaxis específica de MySQL — SQLite no soporta ENUM y el
// AutoMigrate falla. Definimos las tablas a mano, con los mismos nombres
// y columnas relevantes (rol queda como TEXT, que en SQLite acepta
// cualquier string igual que el enum en MySQL).
func dbTestEnMemoria(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("error al abrir sqlite en memoria: %v", err)
	}

	esquema := `
	CREATE TABLE usuarios (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL,
		apellido TEXT NOT NULL,
		email TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		rol TEXT NOT NULL DEFAULT 'cliente',
		telefono TEXT,
		fecha_registro DATETIME,
		activo BOOLEAN DEFAULT 1,
		deleted_at DATETIME
	);

	CREATE TABLE eventos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		titulo TEXT NOT NULL,
		descripcion TEXT,
		fecha_hora DATETIME NOT NULL,
		duracion_minutos INTEGER,
		lugar TEXT,
		direccion TEXT,
		ciudad TEXT,
		categoria TEXT,
		capacidad_total INTEGER NOT NULL,
		entradas_vendidas INTEGER DEFAULT 0,
		precio_base REAL NOT NULL,
		imagen_url TEXT,
		estado TEXT DEFAULT 'activo',
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);

	CREATE TABLE entradas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		codigo_qr TEXT NOT NULL,
		usuario_id INTEGER NOT NULL,
		evento_id INTEGER NOT NULL,
		precio_pagado REAL NOT NULL,
		estado TEXT DEFAULT 'activa',
		fecha_compra DATETIME,
		fecha_cancelacion DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);
	`

	if err := db.Exec(esquema).Error; err != nil {
		t.Fatalf("error al crear esquema de prueba: %v", err)
	}

	return db
}

// insertarEventoTest inserta un evento directamente en la BD de prueba
// (vía db.Create, sin pasar por el DAO) para que esté disponible cuando
// el service ejecute el UpdateColumn de entradas_vendidas dentro de la
// transacción.
func insertarEventoTest(t *testing.T, db *gorm.DB, evento *domain.Evento) {
	if err := db.Create(evento).Error; err != nil {
		t.Fatalf("error al insertar evento de prueba: %v", err)
	}
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
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	evento := eventoActivoConDisponibilidad()
	insertarEventoTest(t, db, evento) // necesario para el UpdateColumn dentro de la transacción
	dto := domain.DTOComprarEntrada{EventoID: 1}

	mockEventoDAO.On("BuscarPorID", uint(1)).Return(evento, nil)

	entrada, err := svc.ComprarEntrada(1, dto)

	assert.NoError(t, err)
	assert.NotNil(t, entrada)
	assert.Equal(t, uint(1), entrada.UsuarioID)
	assert.Equal(t, uint(1), entrada.EventoID)
	assert.Equal(t, domain.EstadoEntradaActiva, entrada.Estado)
	assert.Equal(t, evento.PrecioBase, entrada.PrecioPagado)
	assert.NotEmpty(t, entrada.CodigoQR)

	// Verificamos que la transacción de verdad incrementó el contador en la BD
	var eventoActualizado domain.Evento
	db.First(&eventoActualizado, 1)
	assert.Equal(t, 51, eventoActualizado.EntradasVendidas)
}

func TestComprarEntrada_ErrorEventoNoExiste(t *testing.T) {
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	mockEventoDAO.On("BuscarPorID", uint(999)).Return(nil, errors.New("record not found"))

	entrada, err := svc.ComprarEntrada(1, domain.DTOComprarEntrada{EventoID: 999})

	assert.Error(t, err)
	assert.Nil(t, entrada)
	assert.Contains(t, err.Error(), "record not found")
}

func TestComprarEntrada_ErrorEventoCancelado(t *testing.T) {
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	eventoCancelado := eventoActivoConDisponibilidad()
	eventoCancelado.Estado = domain.EstadoCancelado // <-- cancelado

	mockEventoDAO.On("BuscarPorID", uint(1)).Return(eventoCancelado, nil)

	entrada, err := svc.ComprarEntrada(1, domain.DTOComprarEntrada{EventoID: 1})

	assert.Error(t, err)
	assert.Nil(t, entrada)
	assert.Contains(t, err.Error(), "no está disponible")
}

func TestComprarEntrada_ErrorEventoAgotado(t *testing.T) {
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	mockEntradaDAO.On("ListarPorUsuario", uint(99)).Return([]domain.Entrada{}, nil)

	entradas, err := svc.MisEntradas(99)

	assert.NoError(t, err)
	assert.Empty(t, entradas)
}

// ════════════════════════════════════════════════════════════════════
// Tests de Cancelar Entrada
// ════════════════════════════════════════════════════════════════════

func TestCancelarEntrada_Exitoso(t *testing.T) {
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

	// El evento necesita existir en la BD de prueba porque CancelarEntrada
	// ejecuta un UpdateColumn real sobre la tabla eventos dentro de la transacción.
	insertarEventoTest(t, db, eventoActivoConDisponibilidad())

	entrada := &domain.Entrada{
		ID:        1,
		UsuarioID: 5,
		EventoID:  1,
		Estado:    domain.EstadoEntradaActiva,
	}

	mockEntradaDAO.On("BuscarPorID", uint(1)).Return(entrada, nil)

	err := svc.CancelarEntrada(1, 5)

	assert.NoError(t, err)
	assert.Equal(t, domain.EstadoEntradaCancelada, entrada.Estado)
	assert.NotNil(t, entrada.FechaCancelacion)

	// Verificamos que la transacción de verdad liberó el cupo en la BD
	var eventoActualizado domain.Evento
	db.First(&eventoActualizado, 1)
	assert.Equal(t, 49, eventoActualizado.EntradasVendidas)
}

func TestCancelarEntrada_ErrorPropietarioDistinto(t *testing.T) {
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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

	nuevaEntrada, err := svc.TransferirEntrada(1, propietario, dto)

	assert.NoError(t, err)
	assert.NotNil(t, nuevaEntrada)
	assert.Equal(t, destinatario.ID, nuevaEntrada.UsuarioID)
	assert.Equal(t, domain.EstadoEntradaTransferida, entrada.Estado)

	// Verificamos que la nueva entrada quedó persistida de verdad en la BD
	var entradaEnDB domain.Entrada
	db.First(&entradaEnDB, nuevaEntrada.ID)
	assert.Equal(t, destinatario.ID, entradaEnDB.UsuarioID)
	assert.Equal(t, domain.EstadoEntradaActiva, entradaEnDB.Estado)
}

func TestTransferirEntrada_ErrorAutoTransferencia(t *testing.T) {
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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
	db := dbTestEnMemoria(t)
	mockEntradaDAO := new(MockEntradaDAO)
	mockEventoDAO := new(MockEventoDAO)
	mockUsuarioDAO := new(MockUsuarioDAO)
	svc := services.NuevoEntradaService(db, mockEntradaDAO, mockEventoDAO, mockUsuarioDAO)

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
