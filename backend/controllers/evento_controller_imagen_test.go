package controllers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticketsya/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// crearRequestConArchivo arma un request multipart/form-data con un
// archivo de prueba en el campo "imagen", simulando lo que hace el
// navegador al subir una foto desde <input type="file">.
func crearRequestConArchivo(t *testing.T, url string, nombreArchivo string, contenido []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	parte, err := writer.CreateFormFile("imagen", nombreArchivo)
	if err != nil {
		t.Fatalf("error al crear form file: %v", err)
	}
	parte.Write(contenido)
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// ── SubirImagen ──────────────────────────────────────────────────────

func TestSubirImagen_HTTP_200_Exitoso(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerEventoPorID", uint(1)).
		Return(&domain.Evento{ID: 1, Titulo: "Evento Test"}, nil)
	mockSvc.On("ActualizarEvento", uint(1), mock.AnythingOfType("domain.DTOActualizarEvento")).
		Return(&domain.Evento{ID: 1, Titulo: "Evento Test", ImagenURL: "/uploads/eventos/evento-1-123.jpg"}, nil)

	// Contenido binario simple simulando una imagen (no necesita ser un
	// JPG real válido -- el controller solo valida la extensión del
	// nombre de archivo, no el contenido binario)
	contenidoFalso := []byte("contenido-de-imagen-de-prueba")
	req := crearRequestConArchivo(t, "/api/v1/admin/eventos/1/imagen", "foto.jpg", contenidoFalso)
	req.Header.Set("Authorization", generarTokenAdminTest(1))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubirImagen_HTTP_403_RolCliente(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	contenidoFalso := []byte("contenido-de-imagen-de-prueba")
	req := crearRequestConArchivo(t, "/api/v1/admin/eventos/1/imagen", "foto.jpg", contenidoFalso)
	req.Header.Set("Authorization", generarTokenClienteTest(1))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSubirImagen_HTTP_401_SinToken(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	contenidoFalso := []byte("contenido-de-imagen-de-prueba")
	req := crearRequestConArchivo(t, "/api/v1/admin/eventos/1/imagen", "foto.jpg", contenidoFalso)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubirImagen_HTTP_404_EventoNoEncontrado(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerEventoPorID", uint(99)).Return(nil, assert.AnError)

	contenidoFalso := []byte("contenido-de-imagen-de-prueba")
	req := crearRequestConArchivo(t, "/api/v1/admin/eventos/99/imagen", "foto.jpg", contenidoFalso)
	req.Header.Set("Authorization", generarTokenAdminTest(1))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubirImagen_HTTP_400_ExtensionNoPermitida(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerEventoPorID", uint(1)).
		Return(&domain.Evento{ID: 1, Titulo: "Evento Test"}, nil)

	contenidoFalso := []byte("contenido-cualquiera")
	// .exe no está en el mapa de extensiones permitidas (jpg/jpeg/png/webp)
	req := crearRequestConArchivo(t, "/api/v1/admin/eventos/1/imagen", "archivo.exe", contenidoFalso)
	req.Header.Set("Authorization", generarTokenAdminTest(1))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubirImagen_HTTP_400_SinArchivo(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	mockSvc.On("ObtenerEventoPorID", uint(1)).
		Return(&domain.Evento{ID: 1, Titulo: "Evento Test"}, nil)

	// Request multipart vacío, sin ningún archivo en el campo "imagen"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/admin/eventos/1/imagen", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", generarTokenAdminTest(1))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubirImagen_HTTP_400_IDInvalido(t *testing.T) {
	mockSvc := new(MockEventoService)
	router := setupRouterEvento(mockSvc)

	contenidoFalso := []byte("contenido-de-imagen-de-prueba")
	req := crearRequestConArchivo(t, "/api/v1/admin/eventos/abc/imagen", "foto.jpg", contenidoFalso)
	req.Header.Set("Authorization", generarTokenAdminTest(1))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
