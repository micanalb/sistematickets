package services_test

import (
	"errors"
	"testing"
	"time"

	"ticketsya/domain"
	"ticketsya/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ============================================================
// Tests de Actualizar Evento
// ============================================================

func TestActualizarEvento_ExitosoActualizacionParcial(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventoExistente := &domain.Evento{
		ID:               1,
		Titulo:           "Titulo Original",
		Descripcion:      "Descripcion original",
		Lugar:            "Lugar original",
		CapacidadTotal:   1000,
		EntradasVendidas: 100,
		PrecioBase:       5000.0,
		Estado:           domain.EstadoActivo,
	}

	nuevoTitulo := "Titulo Actualizado"
	dto := domain.DTOActualizarEvento{
		Titulo: &nuevoTitulo,
		// el resto de los campos queda nil -> no deberían tocarse
	}

	mockDAO.On("BuscarPorID", uint(1)).Return(eventoExistente, nil)
	mockDAO.On("Actualizar", mock.AnythingOfType("*domain.Evento")).Return(nil)

	evento, err := svc.ActualizarEvento(1, dto)

	assert.NoError(t, err)
	assert.Equal(t, "Titulo Actualizado", evento.Titulo)
	// Los campos no enviados deben permanecer iguales
	assert.Equal(t, "Descripcion original", evento.Descripcion)
	assert.Equal(t, "Lugar original", evento.Lugar)
	assert.Equal(t, 1000, evento.CapacidadTotal)
}

func TestActualizarEvento_ExitosoTodosLosCampos(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventoExistente := &domain.Evento{
		ID:               1,
		Titulo:           "Original",
		CapacidadTotal:   1000,
		EntradasVendidas: 50,
		Estado:           domain.EstadoActivo,
	}

	nuevoTitulo := "Nuevo Titulo"
	nuevaDescripcion := "Nueva descripcion"
	nuevaFecha := time.Now().Add(60 * 24 * time.Hour)
	nuevaDuracion := 240
	nuevoLugar := "Nuevo Lugar"
	nuevaDireccion := "Nueva Direccion 123"
	nuevaCiudad := "Cordoba"
	nuevaCategoria := domain.CategoriaDeporte
	nuevaCapacidad := 2000
	nuevoPrecio := 7500.0
	nuevaImagen := "https://example.com/imagen.jpg"
	nuevoEstado := domain.EstadoCancelado

	dto := domain.DTOActualizarEvento{
		Titulo:          &nuevoTitulo,
		Descripcion:     &nuevaDescripcion,
		FechaHora:       &nuevaFecha,
		DuracionMinutos: &nuevaDuracion,
		Lugar:           &nuevoLugar,
		Direccion:       &nuevaDireccion,
		Ciudad:          &nuevaCiudad,
		Categoria:       &nuevaCategoria,
		CapacidadTotal:  &nuevaCapacidad,
		PrecioBase:      &nuevoPrecio,
		ImagenURL:       &nuevaImagen,
		Estado:          &nuevoEstado,
	}

	mockDAO.On("BuscarPorID", uint(1)).Return(eventoExistente, nil)
	mockDAO.On("Actualizar", mock.AnythingOfType("*domain.Evento")).Return(nil)

	evento, err := svc.ActualizarEvento(1, dto)

	assert.NoError(t, err)
	assert.Equal(t, "Nuevo Titulo", evento.Titulo)
	assert.Equal(t, "Nueva descripcion", evento.Descripcion)
	assert.Equal(t, 240, evento.DuracionMinutos)
	assert.Equal(t, "Nuevo Lugar", evento.Lugar)
	assert.Equal(t, "Nueva Direccion 123", evento.Direccion)
	assert.Equal(t, "Cordoba", evento.Ciudad)
	assert.Equal(t, domain.CategoriaDeporte, evento.Categoria)
	assert.Equal(t, 2000, evento.CapacidadTotal)
	assert.Equal(t, 7500.0, evento.PrecioBase)
	assert.Equal(t, "https://example.com/imagen.jpg", evento.ImagenURL)
	assert.Equal(t, domain.EstadoCancelado, evento.Estado)
}

func TestActualizarEvento_ErrorEventoNoEncontrado(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	mockDAO.On("BuscarPorID", uint(99)).Return(nil, gorm.ErrRecordNotFound)

	nuevoTitulo := "No importa"
	dto := domain.DTOActualizarEvento{Titulo: &nuevoTitulo}

	evento, err := svc.ActualizarEvento(99, dto)

	assert.Error(t, err)
	assert.Nil(t, evento)
	mockDAO.AssertNotCalled(t, "Actualizar", mock.Anything)
}

func TestActualizarEvento_ErrorCapacidadMenorAVentas(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventoExistente := &domain.Evento{
		ID:               1,
		CapacidadTotal:   1000,
		EntradasVendidas: 800, // ya vendió 800
	}

	nuevaCapacidadMenor := 500 // quiere bajar a 500, menos que las 800 vendidas
	dto := domain.DTOActualizarEvento{CapacidadTotal: &nuevaCapacidadMenor}

	mockDAO.On("BuscarPorID", uint(1)).Return(eventoExistente, nil)

	evento, err := svc.ActualizarEvento(1, dto)

	assert.Error(t, err)
	assert.Nil(t, evento)
	assert.Contains(t, err.Error(), "entradas ya vendidas")
	mockDAO.AssertNotCalled(t, "Actualizar", mock.Anything)
}

func TestActualizarEvento_ErrorAlGuardar(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventoExistente := &domain.Evento{ID: 1, Titulo: "Original"}
	nuevoTitulo := "Nuevo"
	dto := domain.DTOActualizarEvento{Titulo: &nuevoTitulo}

	mockDAO.On("BuscarPorID", uint(1)).Return(eventoExistente, nil)
	mockDAO.On("Actualizar", mock.AnythingOfType("*domain.Evento")).Return(errors.New("error de conexion"))

	evento, err := svc.ActualizarEvento(1, dto)

	assert.Error(t, err)
	assert.Nil(t, evento)
}

// ============================================================
// Tests de Obtener Reporte
// ============================================================

func TestObtenerReporte_Exitoso(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	eventoConEntradas := &domain.Evento{
		ID:               1,
		Titulo:           "Festival de Jazz",
		CapacidadTotal:   500,
		EntradasVendidas: 320,
		Entradas: []domain.Entrada{
			{ID: 1, UsuarioID: 10, EventoID: 1, Estado: domain.EstadoEntradaActiva},
			{ID: 2, UsuarioID: 11, EventoID: 1, Estado: domain.EstadoEntradaActiva},
		},
	}

	mockDAO.On("ReportePorID", uint(1)).Return(eventoConEntradas, nil)

	reporte, err := svc.ObtenerReporte(1)

	assert.NoError(t, err)
	assert.NotNil(t, reporte)
	assert.Equal(t, "Festival de Jazz", reporte.Titulo)
	assert.Len(t, reporte.Entradas, 2)
}

func TestObtenerReporte_ErrorNoEncontrado(t *testing.T) {
	mockDAO := new(MockEventoDAO)
	svc := services.NuevoEventoService(mockDAO)

	mockDAO.On("ReportePorID", uint(99)).Return(nil, gorm.ErrRecordNotFound)

	reporte, err := svc.ObtenerReporte(99)

	assert.Error(t, err)
	assert.Nil(t, reporte)
}
